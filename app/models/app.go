package models

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/coopernurse/gorp"
	"github.com/pborman/uuid"
	"google.golang.org/api/drive/v2"
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

func (app *App) Bundles(txn gorp.SqlExecutor) ([]*Bundle, error) {
	var bundles []*Bundle
	_, err := txn.Select(&bundles, "SELECT * FROM bundle WHERE app_id = ? ORDER BY id DESC", app.Id)
	if err != nil {
		return nil, err
	}
	return bundles, nil
}

func (app *App) BundlesByPlatformType(txn gorp.SqlExecutor, platformType BundlePlatformType) ([]*Bundle, error) {
	var bundles []*Bundle
	_, err := txn.Select(&bundles, "SELECT * FROM bundle WHERE app_id = ? AND platform_type = ? ORDER BY id DESC", app.Id, platformType)
	if err != nil {
		return nil, err
	}
	return bundles, nil
}

func (app *App) BundlesWithPager(txn gorp.SqlExecutor, page, limit int) (Bundles, int, error) {
	if page < 1 {
		page = 1
	}

	count, err := txn.SelectInt("SELECT COUNT(*) FROM bundle WHERE app_id = ?", app.Id)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if int(count) <= offset {
		// 空であることが明らかなのでそのまま返す
		return Bundles([]*Bundle{}), int(count), nil
	}

	var bundles []*Bundle
	_, err = txn.Select(&bundles, "SELECT * FROM bundle WHERE app_id = ? ORDER BY id DESC LIMIT ? OFFSET ?", app.Id, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return Bundles(bundles), int(count), nil
}

func (app *App) Authorities(txn gorp.SqlExecutor) ([]*Authority, error) {
	var authorities []*Authority
	_, err := txn.Select(&authorities, "SELECT * FROM authority WHERE app_id = ? ORDER BY id ASC", app.Id)
	if err != nil {
		return nil, err
	}
	return authorities, nil
}

func (app *App) GetMaxRevisionByBundleVersion(txn gorp.SqlExecutor, bundleVersion string) (int, error) {
	revision, err := txn.SelectInt(
		"SELECT IFNULL(MAX(revision), 0) FROM bundle WHERE app_id = ? AND bundle_version = ?",
		app.Id,
		bundleVersion,
	)
	return int(revision), err
}

func (app *App) UserApps(txn gorp.SqlExecutor) ([]*UserApp, error) {
	var userApps []*UserApp
	_, err := txn.Select(&userApps, "SELECT * FROM user_app WHERE app_id = ? ORDER BY id ASC", app.Id)
	if err != nil {
		return nil, err
	}
	return userApps, nil
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

func (app *App) Save(txn gorp.SqlExecutor) error {
	return txn.Insert(app)
}

func (app *App) RefreshToken(txn gorp.SqlExecutor) error {
	current, err := GetApp(txn, app.Id)
	if err != nil {
		return err
	}

	current.ApiToken = NewToken()

	_, err = txn.Update(current)
	return err
}

func (app *App) Update(txn gorp.SqlExecutor) error {
	current, err := GetApp(txn, app.Id)
	if err != nil {
		return err
	}

	current.Title = app.Title
	current.Description = app.Description

	_, err = txn.Update(current)
	return err
}

func (app *App) DeleteFromDB(txn gorp.SqlExecutor) error {
	_, err := txn.Delete(app)
	return err
}

func (app *App) DeleteFromStorage(s *GoogleService) error {
	return s.DeleteBucket(app.FileId)
}

func (app *App) Delete(txn gorp.SqlExecutor, s *GoogleService) error {
	if err := app.DeleteBundles(txn); err != nil {
		return err
	}
	if err := app.DeleteAuthorities(txn); err != nil {
		return err
	}
	if err := app.DeleteUserApps(txn); err != nil {
		return err
	}
	if err := app.DeleteFromDB(txn); err != nil {
		return err
	}
	return app.DeleteFromStorage(s)
}

func (app *App) DeleteBundles(txn gorp.SqlExecutor) error {
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

func (app *App) DeleteAuthority(txn gorp.SqlExecutor, s *GoogleService, authority *Authority) error {
	if err := authority.DeleteFromDB(txn); err != nil {
		return err
	}

	return s.DeletePermission(app.FileId, authority.PermissionId)
}

func (app *App) DeleteAuthorities(txn gorp.SqlExecutor) error {
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

func (app *App) DeleteUserApps(txn gorp.SqlExecutor) error {
	userApps, err := app.UserApps(txn)
	if err != nil {
		return err
	}

	args := make([]interface{}, len(userApps))
	for i, userApp := range userApps {
		args[i] = userApp
	}

	_, err = txn.Delete(args...)
	return err
}

func (app *App) HasAuthorityForEmail(txn gorp.SqlExecutor, email string) (bool, error) {
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

func (app *App) CreateBundle(dbm *gorp.DbMap, s *GoogleService, bundle *Bundle) error {
	bundle.AppId = app.Id

	bundleInfo, err := NewBundleInfo(bundle.File, bundle.PlatformType)
	if err != nil {
		return err
	}
	if len(bundleInfo.Version) == 0 {
		return &BundleParseError{}
	}
	bundle.BundleInfo = bundleInfo

	// increment revision number & save application information
	err = Transact(dbm, func(txn gorp.SqlExecutor) error {
		maxRevision, err := app.GetMaxRevisionByBundleVersion(txn, bundleInfo.Version)
		if err != nil {
			return err
		}
		bundle.Revision = maxRevision + 1
		bundle.FileName = bundle.BuildFileName()
		return bundle.Save(txn)
	})
	if err != nil {
		panic(err)
	}

	// upload file
	file, err := s.InsertFile(bundle.File, bundle.FileName, app.FileId)
	if err != nil {
		return err
	}

	// update FileId
	bundle.FileId = app.FileId + "/" + file.Name
	return Transact(dbm, func(txn gorp.SqlExecutor) error {
		return bundle.Update(txn)
	})
}

func (app *App) CreateAuthority(txn gorp.SqlExecutor, s *GoogleService, authority *Authority) error {
	authority.AppId = app.Id

	permission := s.CreateUserPermission(authority.Email, "READER")
	permissionInserted, err := s.InsertPermission(app.FileId, permission)
	if err != nil {
		return err
	}
	authority.PermissionId = permissionInserted.Entity

	return authority.Save(txn)
}

func CreateApp(txn gorp.SqlExecutor, s *GoogleService, app *App) error {
	bucket, err := s.CreateBucket()
	if err != nil {
		return err
	}
	app.FileId = bucket.Name

	return app.Save(txn)
}

func GetApp(txn gorp.SqlExecutor, id int) (*App, error) {
	var app App
	if err := txn.SelectOne(&app, "SELECT * FROM app WHERE id = ?", id); err != nil {
		return nil, err
	}
	return &app, nil
}

func GetAppByApiToken(txn gorp.SqlExecutor, apiToken string) (*App, error) {
	var app App
	if err := txn.SelectOne(&app, "SELECT * FROM app where api_token = ?", apiToken); err != nil {
		return nil, err
	}
	return &app, nil
}

func GetApps(txn gorp.SqlExecutor, fileIds []string) ([]*App, error) {
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

func GetAppsByIds(txn gorp.SqlExecutor, ids []int) ([]*App, error) {
	if len(ids) <= 0 {
		return []*App{}, nil
	}

	args := make([]interface{}, len(ids))
	quarks := make([]string, len(ids))
	for i, id := range ids {
		args[i] = id
		quarks[i] = "?"
	}

	var apps []*App
	_, err := txn.Select(&apps, fmt.Sprintf("SELECT * FROM app WHERE id in (%s) ORDER BY id DESC", strings.Join(quarks, ",")), args...)
	if err != nil {
		return nil, err
	}

	return apps, nil
}
