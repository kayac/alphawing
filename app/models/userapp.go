package models

import (
	"database/sql"

	"github.com/coopernurse/gorp"
)

type UserApp struct {
	Id     int `db:"id"`
	UserId int `db:"user_id"`
	AppId  int `db:"app_id"`
}

func (ua *UserApp) Save(txn gorp.SqlExecutor) error {
	return txn.Insert(ua)
}

func (ua *UserApp) Delete(txn gorp.SqlExecutor) error {
	_, err := txn.Delete(ua)
	return err
}

func TryCreateUserApp(txn gorp.SqlExecutor, userId, appId int) (*UserApp, error) {
	ua, err := GetUserApp(txn, userId, appId)
	if err == sql.ErrNoRows {
		ua = &UserApp{
			UserId: userId,
			AppId:  appId,
		}
		if err := ua.Save(txn); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return ua, nil
}

func TryDeleteUserApp(txn gorp.SqlExecutor, userId, appId int) error {
	ua, err := GetUserApp(txn, userId, appId)
	if err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return err
	}
	return ua.Delete(txn)
}

func GetUserApp(txn gorp.SqlExecutor, userId, appId int) (*UserApp, error) {
	var ua UserApp
	err := txn.SelectOne(&ua, "SELECT * FROM user_app WHERE user_id = ? AND app_id = ?", userId, appId)
	if err != nil {
		return nil, err
	}
	return &ua, nil
}

func GetUserAppsByUserId(txn gorp.SqlExecutor, userId int) ([]*UserApp, error) {
	var userApps []*UserApp
	if _, err := txn.Select(&userApps, "SELECT * FROM user_app WHERE user_id = ?", userId); err != nil {
		return nil, err
	}
	return userApps, nil
}
