package permission

import "github.com/kayac/alphawing/app/storage"

type Permission interface {
	CreateGroup(name string) (ident storage.FileIdentifier, err error)
	AddUser(ident storage.FileIdentifier, email string) error
	DeleteUser(ident storage.FileIdentifier, email string) error
	GetUserList(ident storage.FileIdentifier, name string) (emails []string, err error)
}
