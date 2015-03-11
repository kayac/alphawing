package storage

import (
	"os"

	"code.google.com/p/google-api-go-client/drive/v2"

	"github.com/kayac/alphawing/app/googleservice"
)

type GoogleDrive struct {
	Service *googleservice.GoogleService
	Parent  *drive.ParentReference
}

func (gd *GoogleDrive) Upload(file *os.File, filename string) (FileIdentifier, error) {
	driveFile := &drive.File{
		Title:   filename,
		Parents: []*drive.ParentReference{gd.Parent},
	}

	driveFile, err := gd.Service.FilesService.Insert(driveFile).Media(file).Do()
	if err != nil {
		return FileIdentifier{}, err
	}

	ident := FileIdentifier{
		FileId:   driveFile.Id,
		Filename: filename,
	}
	return ident, nil
}

func (gd *GoogleDrive) GetUrl(ident FileIdentifier) (string, error) {
	return "", nil
}

func (gd *GoogleDrive) Delete(ident FileIdentifier) error {
	return nil
}
