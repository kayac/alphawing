package googledrive

import (
	"github.com/kayac/alphawing/app/googleservice"
	"github.com/kayac/alphawing/app/storage"
)

type GoogleDrive struct {
	Service *googleservice.GoogleService
}

func (gd *GoogleDrive) CreateGroup(name string) (storage.FileIdentifier, error) {
	return storage.FileIdentifier, nil
}

func (gd *GoogleDrive) AddUser(email string) error {
	return nil
}

func (gd *GoogleDrive) DeleteUser(email string) error {
	return nil
}

func (gd *GoogleDrive) GetUserList(name string) ([]string, error) {
	return []string{}, nil
}
