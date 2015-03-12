package permission

import "github.com/kayac/alphawing/app/storage"

type Permission interface {
	CreateGroup(name string) (ident storage.FileIdentifier, err error)
	AddUser(email string) error
	UpdateUser(email string) error
	DeleteUser(email string) error
	GetUserList(name string) (emails []string, err error)
}
