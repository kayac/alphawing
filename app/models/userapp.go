package models

import "github.com/coopernurse/gorp"

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

func GetUserAppsByUserId(txn gorp.SqlExecutor, userId int) ([]*UserApp, error) {
	var userApps []*UserApp
	if _, err := txn.Select(&userApps, "SELECT * FROM user_app WHERE user_id = ?", userId); err != nil {
		return nil, err
	}
	return userApps, nil
}
