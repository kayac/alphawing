package googleservice

import "code.google.com/p/google-api-go-client/drive/v2"

type GoogleService struct {
	FilesService *drive.FilesService
}
