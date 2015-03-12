package googledrive

import (
	"github.com/kayac/alphawing/app/googleservice"
	"github.com/kayac/alphawing/app/storage"
)

var defaultRole = "reader"

type GoogleDrive struct {
	Service *googleservice.GoogleService
}

func (gd *GoogleDrive) CreateGroup(name string) (*storage.FileIdentifier, error) {
	ident := &storage.FileIdentifier{}

	file, err := gd.Service.CreateFolder(name)
	if err != nil {
		return ident, err
	}

	ident.FileId = file.Id
	return ident, nil
}

func (gd *GoogleDrive) AddUser(ident *storage.FileIdentifier, email string) error {
	perm := gd.Service.CreateUserPermission(email, defaultRole)
	_, err := gd.Service.InsertPermission(ident.FileId, perm)
	return err
}

func (gd *GoogleDrive) DeleteUser(ident *storage.FileIdentifier, email string) error {
	permId, err := gd.Service.GetPermissionId(ident.FileId, email)
	if err != nil {
		return err
	}

	return gd.Service.DeletePermission(ident.FileId, permId)
}

func (gd *GoogleDrive) GetUserList(ident *storage.FileIdentifier) ([]string, error) {
	permList, err := gd.Service.GetPermissionList(ident.FileId)
	if err != nil {
		return err
	}

	var permIds []string
	for _, perm := range permList.Items {
		permIds = append(permIds, perm.Id)
	}

	return permIds, nil
}
