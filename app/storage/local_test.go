package storage

import (
	"crypto/rand"
	"encoding/binary"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestSatisfyLocal(t *testing.T) {
	l := Local{}

	_, ok := interface{}(l).(Storage)

	if !ok {
		t.Error("storage.Local is not satisfy of storage.Storage")
	}
}

func randStringForTestFile() string {
	var n uint64
	binary.Read(rand.Reader, binary.LittleEndian, &n)
	return strconv.FormatUint(n, 36)
}

func TestUploadLocal(t *testing.T) {
	tempParentDir := os.TempDir()
	tempDir := filepath.Join(tempParentDir, randStringForTestFile())
	err := os.Mkdir(tempDir, os.ModeTemporary|0755)
	if err != nil {
		t.Fatal("cannot create tempdir: ", err)
	}
	defer os.Remove(tempDir)

	srcPath := filepath.Join(tempDir, randStringForTestFile())
	src, err := os.Create(srcPath)
	if err != nil {
		t.Fatal("cannot create source file: ", err)
	}

	data := randStringForTestFile()
	_, err = io.WriteString(src, data)
	if err != nil {
		t.Fatal("cannot write to source file: ", err)
	}
	src.Close()
	src, err = os.Open(srcPath)
	if err != nil {
		t.Fatal("cannot reopen source file: ", err)
	}

	dstDir := filepath.Join(tempDir, randStringForTestFile())
	if err != nil {
		t.Fatal("cannot create destination dir: ", err)
	}

	err = os.Mkdir(dstDir, os.ModeTemporary|0755)
	l := Local{Dir: dstDir}

	dstBasename := randStringForTestFile()
	fileId, err := l.Upload(src, dstBasename)
	if err != nil {
		t.Fatal("unexpected error: ", err)
	}

	dst, err := os.Open(filepath.Join(dstDir, fileId+"_"+dstBasename))
	if err != nil {
		t.Fatal("cannot open destination file: ", err)
	}
	defer dst.Close()

	r, storageFile, err := l.DownloadFile(fileId)
	if err != nil {
		t.Fatal("unexpected error: ", err)
	}
	defer r.(io.Closer).Close()

	gotBytesData, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal("cannot read by download file reader: ", err)
	}

	gotData := string(gotBytesData)
	if gotData != data {
		t.Fatalf("unmatch read data: got %s, expected %s.", gotData, data)
	}

	if storageFile.Filename != dstBasename {
		t.Fatalf("unmatch filename: got %s, expected %s.", storageFile.Filename, dstBasename)
	}

	newBasename := randStringForTestFile()
	err = l.ChangeFilename(fileId, newBasename)
	if err != nil {
		t.Fatal("unexpected error: ", err)
	}

}
