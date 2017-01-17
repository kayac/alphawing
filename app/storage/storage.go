package storage

import (
	"io"
	"os"
	"time"
)

type Storage interface {
	GetUrl(fileId string) (url string, err error)
	DownloadFile(fileId string) (r io.Reader, file StorageFile, err error)
	GetFileList(viewerEmail string) (fileIds []string, err error)

	Upload(file *os.File, filename string) (fileId string, err error)
	ChangeFilename(fileId string, filename string) error

	Delete(fileId string) error
	DeleteAll() error
}

type StorageFile struct {
	Modtime  time.Time
	Filename string
}
