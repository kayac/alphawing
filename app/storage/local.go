package storage

import (
	"io"
	"os"
	"path/filepath"
)

type Local struct {
	Dir string
}

// not implemented yet
func (l Local) GetUrl(fileId string) (string, error) {
	return "", nil
}

func (l Local) DownloadFile(fileId string) (io.Reader, StorageFile, error) {
	return nil, StorageFile{}, nil
}

func (l Local) GetFileList(viewerEmail string) ([]string, error) {
	return nil, nil
}

func (l Local) Upload(src *os.File, filename string) (string, error) {
	dstPath := filepath.Join(l.Dir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return "", err
	}

	return filename, dst.Close()
}

func (l Local) ChangeFilename(fileId string, filename string) error {
	return nil
}

func (l Local) Delete(fileId string) error {
	return nil
}

func (l Local) DeleteAll() error {
	return nil
}
