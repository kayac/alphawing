package storage

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"code.google.com/p/google-api-go-client/drive/v2"

	"github.com/kayac/alphawing/app/googleservice"
)

type GoogleDrive struct {
	Service *googleservice.GoogleService
	Parent  *drive.ParentReference
}

func NewGoogleDrive(service *googleservice.GoogleService, parentFileId string) (GoogleDrive, error) {
	parent := &drive.ParentReference{
		Id: parentFileId,
	}
	gd := GoogleDrive{
		Service: service,
		Parent:  parent,
	}

	return gd, nil
}

func (gd GoogleDrive) GetUrl(fileId string) (string, error) {
	file, err := gd.Service.FilesService.Get(fileId).Do()
	if err != nil {
		return "", err
	}
	return file.DownloadUrl, nil
}

func (gd GoogleDrive) DownloadFile(fileId string) (*http.Response, StorageFile, error) {
	file, err := gd.Service.FilesService.Get(fileId).Do()
	if err != nil {
		return &http.Response{}, StorageFile{}, err
	}

	modtime, err := time.Parse(time.RFC3339, file.ModifiedDate)
	if err != nil {
		return &http.Response{}, StorageFile{}, err
	}
	storageFile := StorageFile{
		Modtime:  modtime,
		Filename: file.OriginalFilename,
	}

	resp, err := gd.Service.Client.Get(file.DownloadUrl)
	if err != nil {
		return &http.Response{}, StorageFile{}, err
	}

	return resp, storageFile, nil
}

func (gd GoogleDrive) GetFileList(viewerEmail string) ([]string, error) {
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

func (gd GoogleDrive) Upload(file *os.File, filename string) (string, error) {
	driveFile := &drive.File{
		Title:   filename,
		Parents: []*drive.ParentReference{gd.Parent},
	}

	driveFile, err := gd.Service.FilesService.Insert(driveFile).Media(file).Do()
	if err != nil {
		return "", err
	}
	return driveFile.Id, nil
}

func (gd GoogleDrive) ChangeFilename(fileId string, filename string) error {
	file, err := gd.Service.FilesService.Get(fileId).Do()
	if err != nil {
		return err
	}

	file.Title = filename
	_, err = gd.Service.FilesService.Update(fileId, file).Do()

	return err
}

func (gd GoogleDrive) Delete(fileId string) error {
	return gd.Service.FilesService.Delete(fileId).Do()
}

func (gd GoogleDrive) DeleteAll() error {
	fileList, err := gd.Service.FilesService.List().Do()
	if err != nil {
		return err
	}

	for _, file := range fileList.Items {
		err = gd.Delete(file.Id)
		if err != nil {
			return err
		}
	}

	return nil
}
