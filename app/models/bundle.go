package models

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/coopernurse/gorp"
	"github.com/kayac/alphawing/app/permission"
	"github.com/kayac/alphawing/app/storage"
)

type BundlePlatformType int

const (
	BundlePlatformTypeAndroid BundlePlatformType = 1 + iota
	BundlePlatformTypeIOS
)

func (platformType BundlePlatformType) Extention() BundleFileExtension {
	var ext BundleFileExtension
	if platformType == BundlePlatformTypeAndroid {
		ext = BundleFileExtensionAndroid
	} else if platformType == BundlePlatformTypeIOS {
		ext = BundleFileExtensionIOS
	}
	return ext
}

func (platformType BundlePlatformType) String() string {
	var str string
	if platformType == BundlePlatformTypeAndroid {
		str = "android"
	} else if platformType == BundlePlatformTypeIOS {
		str = "ios"
	}
	return str
}

type BundleFileExtension string

const (
	BundleFileExtensionAndroid BundleFileExtension = ".apk"
	BundleFileExtensionIOS     BundleFileExtension = ".ipa"
)

func (ext BundleFileExtension) IsValid() bool {
	var ok bool
	if ext == BundleFileExtensionAndroid {
		ok = true
	} else if ext == BundleFileExtensionIOS {
		ok = true
	}
	return ok
}

func (ext BundleFileExtension) PlatformType() BundlePlatformType {
	var platformType BundlePlatformType
	if ext == BundleFileExtensionAndroid {
		platformType = BundlePlatformTypeAndroid
	} else if ext == BundleFileExtensionIOS {
		platformType = BundlePlatformTypeIOS
	}
	return platformType
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

	BundleInfo *BundleInfo           `db:"-"`
	File       *os.File              `db:"-"`
	FileName   string                `db:"-"`
	Storage    storage.Storage       `db:"-"`
	Permission permission.Permission `db:"-"`
}

type BundleJsonResponse struct {
	FileId       string `json:"file_id"`
	Version      string `json:"version"`
	Revision     int    `json:"revision"`
	InstallUrl   string `json:"install_url"`
	QrCodeUrl    string `json:"qr_code_url"`
	PlatformType string `json:"platform_type"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type Bundles []*Bundle

func (bundles Bundles) JsonResponse(ub UriBuilder) ([]*BundleJsonResponse, error) {
	bundlesJsonResponse := []*BundleJsonResponse{}

	for _, bundle := range bundles {
		bundleJsonResponse, err := bundle.JsonResponse(ub)
		if err != nil {
			return nil, err
		}

		bundlesJsonResponse = append(bundlesJsonResponse, bundleJsonResponse)
	}

	return bundlesJsonResponse, nil
}

type BundlesJsonResponse struct {
	TotalCount int                   `json:"total_count"`
	Page       int                   `json:"page"`
	Limit      int                   `json:"limit"`
	Bundles    []*BundleJsonResponse `json:"bundles"`
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
		FileId:       bundle.FileId,
		Version:      bundle.BundleVersion,
		Revision:     bundle.Revision,
		InstallUrl:   installUrl.String(),
		QrCodeUrl:    qrCodeUrl.String(),
		PlatformType: bundle.PlatformType.String(),
		CreatedAt:    bundle.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    bundle.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (bundle *Bundle) Plist(txn gorp.SqlExecutor, ipaUrl *url.URL) (*Plist, error) {
	app, err := bundle.App(txn)
	if err != nil {
		return nil, err
	}

	return NewPlist(app.Title, bundle.BundleVersion, ipaUrl.String()), nil
}

func (bundle *Bundle) PlistReader(txn gorp.SqlExecutor, ipaUrl *url.URL) (io.Reader, error) {
	p, err := bundle.Plist(txn, ipaUrl)
	if err != nil {
		return nil, err
	}

	return p.Reader()
}

func (bundle *Bundle) BuildFileName() string {
	return fmt.Sprintf(
		"app_%d_ver_%s_rev_%d%s",
		bundle.AppId,
		bundle.BundleInfo.Version,
		bundle.Revision,
		bundle.PlatformType.Extention(),
	)
}

func (bundle *Bundle) IsApk() bool {
	var ok bool
	if bundle.PlatformType == BundlePlatformTypeAndroid {
		ok = true
	}
	return ok
}

func (bundle *Bundle) IsIpa() bool {
	var ok bool
	if bundle.PlatformType == BundlePlatformTypeIOS {
		ok = true
	}
	return ok
}

func (bundle *Bundle) App(txn gorp.SqlExecutor) (*App, error) {
	app, err := txn.Get(App{}, bundle.AppId)
	if err != nil {
		return nil, err
	}
	return app.(*App), nil
}

func (bundle *Bundle) PreInsert(s gorp.SqlExecutor) error {
	bundle.BundleVersion = bundle.BundleInfo.Version
	bundle.CreatedAt = time.Now()
	bundle.UpdatedAt = bundle.CreatedAt
	return nil
}

func (bundle *Bundle) PreUpdate(s gorp.SqlExecutor) error {
	bundle.UpdatedAt = time.Now()
	return nil
}

func (bundle *Bundle) Save(txn gorp.SqlExecutor) error {
	return txn.Insert(bundle)
}

func (bundle *Bundle) Update(txn gorp.SqlExecutor) error {
	current, err := GetBundle(txn, bundle.Id)
	if err != nil {
		return err
	}

	current.Description = bundle.Description
	if bundle.FileId != "" {
		current.FileId = bundle.FileId
	}

	_, err = txn.Update(current)
	return err
}

func (bundle *Bundle) DeleteFromDB(txn gorp.SqlExecutor) error {
	_, err := txn.Delete(bundle)
	return err
}

func (bundle *Bundle) DeleteFromGoogleDrive() error {
	if bundle.FileId == "" {
		return nil
	}
	return bundle.Storage.Delete(bundle.FileId)
}

func (bundle *Bundle) Delete(txn gorp.SqlExecutor) error {
	if err := bundle.DeleteFromGoogleDrive(); err != nil {
		code, _, _ := ParseGoogleApiError(err)
		if code != http.StatusNotFound {
			return err
		}
	}
	return bundle.DeleteFromDB(txn)
}

func (bundle *Bundle) DownloadFile() (io.Reader, storage.StorageFile, error) {
	return bundle.Storage.DownloadFile(bundle.FileId)
}

func CreateBundle(txn gorp.SqlExecutor, bundle *Bundle) error {
	return txn.Insert(bundle)
}

func GetBundle(txn gorp.SqlExecutor, id int) (*Bundle, error) {
	var bundle Bundle
	if err := txn.SelectOne(&bundle, "SELECT * FROM bundle WHERE id = ?", id); err != nil {
		return nil, err
	}
	return &bundle, nil
}

func GetBundleByFileId(txn gorp.SqlExecutor, fileId string) (*Bundle, error) {
	var bundle Bundle
	if err := txn.SelectOne(&bundle, "SELECT * FROM bundle WHERE file_id = ?", fileId); err != nil {
		return nil, err
	}
	return &bundle, nil
}
