package models

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"code.google.com/p/google-api-go-client/drive/v2"

	"github.com/coopernurse/gorp"
)

// https://github.com/coopernurse/gorp#mapping-structs-to-tables
type App struct {
	Id          int       `db:"id"`
	Title       string    `db:"title"`
	FileId      string    `db:"file_id"`
	ApiToken    string    `db:"api_token"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func GetAppByApiToken(txn *gorp.Transaction, apiToken string) (*App, error) {
	var app App
	err := txn.SelectOne(&app, "SELECT * FROM app where api_token = ?", apiToken)

	if err != nil {
		return nil, err
	}

	return &app, nil
}

func (app *App) Bundles(txn *gorp.Transaction) ([]*Bundle, error) {
	var bundles []*Bundle
	_, err := txn.Select(&bundles, "SELECT * FROM bundle WHERE app_id = ? ORDER BY id DESC", app.Id)
	if err != nil {
		return nil, err
	}
	return bundles, nil
}

func (app *App) Authorities(txn *gorp.Transaction) ([]*Authority, error) {
	var authorities []*Authority
	_, err := txn.Select(&authorities, "SELECT * FROM authority WHERE app_id = ? ORDER BY id ASC", app.Id)
	if err != nil {
		return nil, err
	}
	return authorities, nil
}

func (app *App) GetMaxRevisionByBundleVersion(txn *gorp.Transaction, bundleVersion string) (int, error) {
	revision, err := txn.SelectInt(
		"SELECT IFNULL(MAX(revision), 0) FROM bundle WHERE app_id = ? AND bundle_version = ?",
		app.Id,
		bundleVersion,
	)
	return int(revision), err
}

func NewToken() string {
	uuid := uuid.NewRandom()
	mac := hmac.New(sha256.New, nil)
	mac.Write([]byte(uuid.String()))
	tokenBytes := mac.Sum(nil)
	return hex.EncodeToString(tokenBytes)
}

// https://github.com/coopernurse/gorp#hooks
func (app *App) PreInsert(s gorp.SqlExecutor) error {
	app.CreatedAt = time.Now()
	app.UpdatedAt = app.CreatedAt
	app.ApiToken = NewToken()
	return nil
}

func (app *App) PreUpdate(s gorp.SqlExecutor) error {
	app.UpdatedAt = time.Now()
	return nil
}

func (app *App) Save(txn *gorp.Transaction) error {
	return txn.Insert(app)
}

func (app *App) RefreshToken(txn *gorp.Transaction) error {
	current, err := GetApp(txn, app.Id)
	if err != nil {
		return err
	}

	current.ApiToken = NewToken()

	_, err = txn.Update(current)
	return err
}

func (app *App) Update(txn *gorp.Transaction) error {
	current, err := GetApp(txn, app.Id)
	if err != nil {
		return err
	}

	current.Title = app.Title
	current.Description = app.Description

	_, err = txn.Update(current)
	return err
}

func (app *App) DeleteFromDB(txn *gorp.Transaction) error {
	_, err := txn.Delete(app)
	return err
}

func (app *App) DeleteFromGoogleDrive(s *GoogleService) error {
	return s.DeleteFile(app.FileId)
}

func (app *App) Delete(txn *gorp.Transaction, s *GoogleService) error {
	if err := app.DeleteBundles(txn); err != nil {
		return err
	}
	if err := app.DeleteAuthorities(txn); err != nil {
		return err
	}
	if err := app.DeleteFromDB(txn); err != nil {
		return err
	}
	return app.DeleteFromGoogleDrive(s)
}

func (app *App) DeleteBundles(txn *gorp.Transaction) error {
	bundles, err := app.Bundles(txn)
	if err != nil {
		return err
	}

	args := make([]interface{}, len(bundles))
	for i, bundle := range bundles {
		args[i] = bundle
	}

	_, err = txn.Delete(args...)
	return err
}

func (app *App) DeleteAuthority(txn *gorp.Transaction, s *GoogleService, authority *Authority) error {
	if err := authority.DeleteFromDB(txn); err != nil {
		return err
	}

	return s.DeletePermission(app.FileId, authority.PermissionId)
}

func (app *App) DeleteAuthorities(txn *gorp.Transaction) error {
	authorities, err := app.Authorities(txn)
	if err != nil {
		return err
	}

	args := make([]interface{}, len(authorities))
	for i, authority := range authorities {
		args[i] = authority
	}

	_, err = txn.Delete(args...)
	return err
}

func (app *App) HasAuthorityForEmail(txn *gorp.Transaction, email string) (bool, error) {
	count, err := txn.SelectInt("SELECT COUNT(id) FROM authority WHERE app_id = ? AND email = ?", app.Id, email)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

func (app *App) ParentReference() *drive.ParentReference {
	return &drive.ParentReference{
		Id: app.FileId,
	}
}

func (app *App) CreateBundle(txn *gorp.Transaction, s *GoogleService, bundle *Bundle) error {
	bundle.AppId = app.Id

	bundleInfo, err := NewBundleInfo(bundle.File, bundle.PlatformType)
	if err != nil {
		return err
	}
	if len(bundleInfo.Version) == 0 {
		return &BundleParseError{}
	}
	bundle.BundleInfo = bundleInfo

	maxRevision, err := app.GetMaxRevisionByBundleVersion(txn, bundleInfo.Version)
	if err != nil {
		return err
	}
	bundle.Revision = maxRevision + 1
	bundle.FileName = bundle.BuildFileName()

	parent := app.ParentReference()
	driveFile, err := s.InsertFile(bundle.File, bundle.FileName, parent)
	if err != nil {
		return err
	}

	bundle.FileId = driveFile.Id
	if err := bundle.Save(txn); err != nil {
		return err
	}

	return nil
}

func (app *App) CreateAuthority(txn *gorp.Transaction, s *GoogleService, authority *Authority) error {
	authority.AppId = app.Id

	permission := s.CreateUserPermission(authority.Email, "reader")
	permissionInserted, err := s.InsertPermission(app.FileId, permission)
	if err != nil {
		return err
	}
	authority.PermissionId = permissionInserted.Id

	return authority.Save(txn)
}

func CreateApp(txn *gorp.Transaction, s *GoogleService, app *App) error {
	driveFolder, err := s.CreateFolder(app.Title)
	if err != nil {
		return err
	}
	app.FileId = driveFolder.Id

	return app.Save(txn)
}

func GetApp(txn *gorp.Transaction, id int) (*App, error) {
	app, err := txn.Get(App{}, id)
	if err != nil {
		return nil, err
	}
	return app.(*App), nil
}

func GetApps(txn *gorp.Transaction, fileIds []string) ([]*App, error) {
	if len(fileIds) <= 0 {
		return []*App{}, nil
	}

	args := make([]interface{}, len(fileIds))
	quarks := make([]string, len(fileIds))
	for i, fileId := range fileIds {
		args[i] = fileId
		quarks[i] = "?"
	}

	var apps []*App
	_, err := txn.Select(&apps, fmt.Sprintf("SELECT * FROM app WHERE file_id in (%s) ORDER BY id DESC", strings.Join(quarks, ",")), args...)
	if err != nil {
		return nil, err
	}

	return apps, nil
}
