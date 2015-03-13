package permission

import "github.com/kayac/alphawing/app/googleservice"

var defaultRole = "reader"

type GoogleDrive struct {
	Service *googleservice.GoogleService
}

func (gd *GoogleDrive) CreateGroup(name string) (string, error) {
	file, err := gd.Service.CreateFolder(name)
	if err != nil {
		return "", err
	}

	return fild.Id, nil
}

func (gd *GoogleDrive) AddUser(groupId string, email string) (string, error) {
	perm := gd.Service.CreateUserPermission(email, defaultRole)
	insertedPerm, err := gd.Service.InsertPermission(groupId, perm)
	if err != nil {
		return "", err
	}
	return insertedPerm.Id, err
}

func (gd *GoogleDrive) DeleteUser(groupId string, permId string) error {
	return gd.Service.DeletePermission(groupId, permId)
}

func (gd *GoogleDrive) GetUserList(groupId string) ([]string, error) {
	permList, err := gd.Service.GetPermissionList(groupId)
	if err != nil {
		return err
	}

	var permIds []string
	for _, perm := range permList.Items {
		permIds = append(permIds, perm.Id)
	}

	return permIds, nil
}

func (gd *GoogleDrive) DeleteGroup(groupId string) error {
	return gd.Service.FilesService.Delete(groupId).Do()
}
