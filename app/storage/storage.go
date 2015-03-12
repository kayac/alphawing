package storage

import (
	"os"
)

type Storage interface {
	GetUrl(identifier FileIdentifier) (url string, err error)
	GetFile(identifier FileIdentifier) (file *os.File, err error)
	GetFileList(viewerEmail string) (fileIds []string, err error)

	Upload(file *os.File, filename string) (ident FileIdentifier, err error)
	ChangeFilename(identifier FileIdentifier, filename string) error

	Delete(identifier FileIdentifier) error
	DeleteAll() error
}

type FileIdentifier struct {
	Filename string
	FileId   string
}
