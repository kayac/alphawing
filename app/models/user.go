package models

import (
	"database/sql"
	"time"

	"github.com/coopernurse/gorp"
	"github.com/revel/revel"
)

type User struct {
	Id        int       `db:"id"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (user *User) PreInsert(s gorp.SqlExecutor) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt
	return nil
}

func (user *User) PreUpdate(s gorp.SqlExecutor) error {
	user.UpdatedAt = time.Now()
	return nil
}

func (user *User) Validate(v *revel.Validation) {
	v.Required(user.Email)
}

func (user *User) Save(txn gorp.SqlExecutor) error {
	return txn.Insert(user)
}

func (user *User) Update(txn gorp.SqlExecutor) error {
	_, err := txn.Update(user)
	return err
}

func (user *User) Delete(txn gorp.SqlExecutor) error {
	_, err := txn.Delete(user)
	return err
}

func (user *User) GetViewableApps(txn gorp.SqlExecutor) ([]*App, error) {
	userApps, err := GetUserAppsByUserId(txn, user.Id)
	if err != nil {
		return nil, err
	}

	if len(userApps) <= 0 {
		return []*App{}, nil
	}

	appIds := make([]int, len(userApps))
	for i, userApp := range userApps {
		appIds[i] = userApp.AppId
	}

	apps, err := GetAppsByIds(txn, appIds)
	if err != nil {
		return nil, err
	}

	return apps, nil
}

func CreateUser(txn gorp.SqlExecutor, user *User) error {
	return txn.Insert(user)
}

func FindOrCreateUser(txn gorp.SqlExecutor, email string) (*User, error) {
	user, err := GetUserFromEmail(txn, email)
	if err == sql.ErrNoRows {
		user = &User{
			Email: email,
		}
		err = user.Save(txn)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUser(txn gorp.SqlExecutor, id int) (*User, error) {
	user, err := txn.Get(User{}, id)
	if err != nil {
		return nil, err
	}
	return user.(*User), nil
}

func GetUserFromEmail(txn gorp.SqlExecutor, email string) (*User, error) {
	var user User
	err := txn.SelectOne(&user, "SELECT * FROM user WHERE email = ?", email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
