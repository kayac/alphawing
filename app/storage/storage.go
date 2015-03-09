package storage

import (
	"os"
)

type Storage interface {
	Upload(file os.File, filename string) (url string, err error)
	GetUrl(filename string) (url string, err error)
	Delete(filename string) error
}

type NotFoundError struct{}

func (e NotFoundError) Error() string {
	return "Not found file in storage"
}
