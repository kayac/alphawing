package models

import (
	"fmt"
	"os"
	"time"

	"github.com/coopernurse/gorp"
)

type BundlePlatformType int

const (
	BundlePlatformTypeAndroid BundlePlatformType = 1 + iota
	BundlePlatformTypeIOS
)

type BundleFileExtension string

const (
	BundleFileExtensionAndroid = ".apk"
	BundleFileExtensionIOS     = ".ipa"
)

func (ext BundleFileExtension) IsValid() (ok bool) {
	ok = false
	if ext == BundleFileExtensionAndroid {
		ok = true
	} else if ext == BundleFileExtensionIOS {
		ok = true
	}
	return
}

type Bundle struct {
	Id            int                `db:"id"`
	AppId         int                `db:"app_id"`
	FileId        string             `db:"file_id"`
	PlatformType  BundlePlatformType `db:"platform_type"`
	BundleVersion string             `db:"bundle_version"`
	Revision      int                `db:"revision"`
	Description   string             `db:"description"`
	CreatedAt     time.Time          `db:"created_at"`
	UpdatedAt     time.Time          `db:"updated_at"`

	AppInfo  *AppInfo `db:"-"`
	File     *os.File `db:"-"`
	FileName string   `db:"-"`
}

type BundleJsonResponse struct {
	FileId     string `json:"file_id"`
	Version    string `json:"version"`
	Revision   int    `json:"revision"`
	InstallUrl string `json:"install_url"`
	QrCodeUrl  string `json:"qr_code_url"`
}

func (bundle *Bundle) JsonResponse(ub UriBuilder) (*BundleJsonResponse, error) {
	installUrl, err := ub.UriFor(fmt.Sprintf("bundle/%d/download", bundle.Id))
	if err != nil {
		return nil, err
	}
	qrCodeUrl, err := ub.UriFor(fmt.Sprintf("bundle/%d", bundle.Id))
	if err != nil {
		return nil, err
	}

	return &BundleJsonResponse{
		FileId:     bundle.FileId,
		Version:    bundle.BundleVersion,
		Revision:   bundle.Revision,
		InstallUrl: installUrl.String(),
		QrCodeUrl:  qrCodeUrl.String(),
	}, nil
}

func (bundle *Bundle) App(txn *gorp.Transaction) (*App, error) {
	app, err := txn.Get(App{}, bundle.AppId)
	if err != nil {
		return nil, err
	}
	return app.(*App), nil
}

func (bundle *Bundle) PreInsert(s gorp.SqlExecutor) error {
	bundle.BundleVersion = bundle.AppInfo.Version
	bundle.CreatedAt = time.Now()
	bundle.UpdatedAt = bundle.CreatedAt
	return nil
}

func (bundle *Bundle) PreUpdate(s gorp.SqlExecutor) error {
	bundle.UpdatedAt = time.Now()
	return nil
}

func (bundle *Bundle) Save(txn *gorp.Transaction) error {
	return txn.Insert(bundle)
}

func (bundle *Bundle) Update(txn *gorp.Transaction) error {
	current, err := GetBundle(txn, bundle.Id)
	if err != nil {
		return err
	}

	current.Description = bundle.Description

	_, err = txn.Update(current)
	return err
}

func (bundle *Bundle) DeleteFromDB(txn *gorp.Transaction) error {
	_, err := txn.Delete(bundle)
	return err
}

func (bundle *Bundle) DeleteFromGoogleDrive(s *GoogleService) error {
	return s.DeleteFile(bundle.FileId)
}

func (bundle *Bundle) Delete(txn *gorp.Transaction, s *GoogleService) error {
	if err := bundle.DeleteFromDB(txn); err != nil {
		return err
	}
	if err := bundle.DeleteFromGoogleDrive(s); err != nil {
		return err
	}
	return nil
}

func CreateBundle(txn *gorp.Transaction, bundle *Bundle) error {
	return txn.Insert(bundle)
}

func GetBundle(txn *gorp.Transaction, id int) (*Bundle, error) {
	bundle, err := txn.Get(Bundle{}, id)
	if err != nil {
		return nil, err
	}
	return bundle.(*Bundle), nil
}
