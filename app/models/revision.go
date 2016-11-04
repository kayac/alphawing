package models

import (
	"github.com/coopernurse/gorp"
)

type Revision struct {
	Id            int                `db:"id"`
	AppId         int                `db:"app_id"`
	PlatformType  BundlePlatformType `db:"platform_type"`
	BundleVersion string             `db:"bundle_version"`
	MaxRevision   int                `db:"max_revision"`
}

func GetMaxRevision(txn gorp.SqlExecutor, appId int, platformType BundlePlatformType, bundleVersion string) (*Revision, error) {
	var revision Revision
	err := txn.SelectOne(
		&revision,
		"SELECT * FROM revision WHERE app_id = ? AND platform_type = ? AND bundle_version = ?",
		appId,
		platformType,
		bundleVersion,
	)
	if err != nil {
		return nil, err
	}
	return &revision, nil
}

func (r *Revision) Save(txn gorp.SqlExecutor) error {
	return txn.Insert(r)
}

func (r *Revision) Update(txn gorp.SqlExecutor) error {
	_, err := txn.Update(r)
	return err
}
