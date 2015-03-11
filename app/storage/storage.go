package storage

import (
	"os"
)

type Storage interface {
	Upload(file *os.File, filename string) (ident FileIdentifier, err error)
	GetUrl(identifier FileIdentifier) (url string, err error)
	Delete(identifier FileIdentifier) error
}

type FileIdentifier struct {
	Filename string
	FileId   string
}
