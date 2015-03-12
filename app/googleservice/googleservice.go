package googleservice

import (
	"errors"

	"code.google.com/p/google-api-go-client/drive/v2"
)

type GoogleService struct {
	FilesService      *drive.FilesService
	PermissionService *drive.PermissionService
}

func (s *GoogleService) CreateFolder(folderName string) (*drive.File, error) {
	driveFolder := &drive.File{
		Title:    folderName,
		MimeType: "application/vnd.google-apps.folder",
	}
	return s.FilesService.Insert(driveFolder).Do()
}

func (s *GoogleService) CreateUserPermission(email string, role string) *drive.Permission {
	return &drive.Permission{
		Role:  role,
		Type:  "user",
		Value: email,
	}
}

func (s *GoogleService) InsertPermission(fileId string, permission *drive.Permission) (*drive.Permission, error) {
	return s.PermissionsService.Insert(fileId, permission).Do()
}

func (s *GoogleService) GetPermissionList(fileId string) (*drive.PermissionList, error) {
	return s.PermissionsService.List(fileId).Do()
}

func (s *GoogleService) DeletePermission(fileId string, permissionId string) error {
	return s.PermissionsService.Delete(fileId, permissionId).Do()
}

func (s *GoogleService) GetPermissionId(fileId string, email string) (string, error) {
	permList, err := s.GetPermissionList(fileId)
	if err != nil {
		return "", err
	}

	var permId string
	for _, perm := range permList.Items {
		if perm.EmailAddress == email {
			permId = perm.Id
			break
		}
	}

	if !permId {
		return "", errors.New("Not found email in folder")
	}

	return permId, nil
}
