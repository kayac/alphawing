package models

import (
	"time"

	"github.com/coopernurse/gorp"
)

type Authority struct {
	Id           int       `db:"id"`
	AppId        int       `db:"app_id"`
	PermissionId string    `db:"permission_id"`
	Email        string    `db:"email"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func (authority *Authority) PreInsert(s gorp.SqlExecutor) error {
	authority.CreatedAt = time.Now()
	authority.UpdatedAt = authority.CreatedAt
	return nil
}

func (authority *Authority) PreUpdate(s gorp.SqlExecutor) error {
	authority.UpdatedAt = time.Now()
	return nil
}

func (authority *Authority) Save(txn gorp.SqlExecutor) error {
	return txn.Insert(authority)
}

func (authority *Authority) DeleteFromDB(txn gorp.SqlExecutor) error {
	_, err := txn.Delete(authority)
	return err
}

func GetAuthority(txn gorp.SqlExecutor, id int) (*Authority, error) {
	authority, err := txn.Get(Authority{}, id)
	if err != nil {
		return nil, err
	}
	return authority.(*Authority), nil
}

func IsExistAuthorityForEmail(txn gorp.SqlExecutor, email string) (bool, error) {
	count, err := txn.SelectInt("SELECT COUNT(id) FROM authority WHERE email = ?", email)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}
