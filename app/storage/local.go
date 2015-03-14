package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Local struct {
	Dir string
}

func generateFileId(filename string) string {
	now := time.Now()
	hash := sha256.Sum256(append([]byte(filename), []byte(now.Format(time.RFC3339))...))
	return hex.EncodeToString(hash[0:5])
}

func (l Local) searchFile(fileId string) (string, error) {
	files, err := ioutil.ReadDir(l.Dir)
	if err != nil {
		return "", err
	}
	var targetBasename string
	for _, file := range files {
		if strings.HasPrefix(file.Name(), fileId) {
			targetBasename = file.Name()
			break
		}
	}

	if targetBasename == "" {
		return "", errors.New("file not exists by fileId")
	}
	return targetBasename, nil
}

// not implemented yet
func (l Local) GetUrl(fileId string) (string, error) {
	return "", nil
}

func (l Local) DownloadFile(fileId string) (io.Reader, StorageFile, error) {
	targetBasename, err := l.searchFile(fileId)
	if err != nil {
		return nil, StorageFile{}, err
	}

	target, err := os.Open(filepath.Join(l.Dir, targetBasename))
	if err != nil {
		return nil, StorageFile{}, err
	}

	fileInfo, err := target.Stat()
	if err != nil {
		return nil, StorageFile{}, err
	}

	filename := strings.Replace(targetBasename, fileId+"_", "", 1)
	storageFile := StorageFile{
		Modtime:  fileInfo.ModTime(),
		Filename: filename,
	}

	return target, storageFile, nil
}

func (l Local) GetFileList(viewerEmail string) ([]string, error) {
	return nil, nil
}

func (l Local) Upload(src *os.File, filename string) (string, error) {
	fileId := generateFileId(filename)
	basename := fileId + "_" + filename
	dstPath := filepath.Join(l.Dir, basename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}

	if _, err = io.Copy(dst, src); err != nil {
		dst.Close()
		return "", err
	}

	return fileId, dst.Close()
}

func (l Local) ChangeFilename(fileId string, filename string) error {
	srcBasename, err := l.searchFile(fileId)
	if err != nil {
		return err
	}
	srcPath := filepath.Join(l.Dir, srcBasename)
	dstPath := filepath.Join(l.Dir, fileId+"_"+filename)

	return os.Rename(srcPath, dstPath)
}

func (l Local) Delete(fileId string) error {
	return nil
}

func (l Local) DeleteAll() error {
	return nil
}
