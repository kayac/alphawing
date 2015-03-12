package storage

import (
	"fmt"
	"os"

	"code.google.com/p/google-api-go-client/drive/v2"

	"github.com/kayac/alphawing/app/googleservice"
)

type GoogleDrive struct {
	Service *googleservice.GoogleService
	Parent  *drive.ParentReference
}

func (gd *GoogleDrive) GetUrl(ident FileIdentifier) (string, error) {
	file, err := gd.Service.FilesService.Get(ident.FileId).Do()
	if err != nil {
		return "", err
	}
	return file.DownloadUrl, nil
}

func (gd *GoogleDrive) GetFileList(viewerEmail string) ([]string, error) {
	var fileIds []string

	q := fmt.Sprintf("'%s' in owners and sharedWithMe = true", viewerEmail)
	fileList, err := gd.Service.FilesService.List().Q(q).Do()
	if err != nil {
		return fileIds, err
	}

	for _, file := range fileList.Items {
		fileIds = append(fileIds, file.Id)
	}

	return fileIds, nil
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

func (gd *GoogleDrive) ChangeFilename(ident FileIdentifier, filename string) error {
	file, err := gd.Service.FilesService.Get(ident.FileId).Do()
	if err != nil {
		return err
	}

	file.Title = filename
	_, err = gd.Service.FilesService.Update(fileId, file).Do()

	return err
}

func (gd *GoogleDrive) Delete(ident FileIdentifier) error {
	return gd.Service.FilesService.Delete(ident.FileId).Do()
}

func (gd *GoogleDrive) DeleteAll() error {
	fileList, err := gd.Service.FilesService.List().Do()
	if err != nil {
		return err
	}

	for _, file := range fileList.Items {
		err = gd.Delete(FileIdentifier{FileId: file.Id})
		if err != nil {
			return err
		}
	}

	return nil
}
