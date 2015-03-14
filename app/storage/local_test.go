package storage

import (
	"crypto/rand"
	"encoding/binary"
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

	src, err := os.Create(filepath.Join(tempDir, randStringForTestFile()))
	if err != nil {
		t.Fatal("cannot create source file: ", err)
	}
	defer src.Close()

	dstDir := filepath.Join(tempDir, randStringForTestFile())
	if err != nil {
		t.Fatal("cannot create destination dir: ", err)
	}

	err = os.Mkdir(dstDir, os.ModeTemporary|0755)
	l := Local{Dir: dstDir}

	dstBasename := randStringForTestFile()
	filename, err := l.Upload(src, dstBasename)
	if err != nil {
		t.Fatal("unexpected error: ", err)
	}
	if filename != dstBasename {
		t.Fatal("unexpected fileId: ", filename)
	}

	dst, err := os.Open(filepath.Join(dstDir, dstBasename))
	if err != nil {
		t.Fatal("cannot open destination file: ", err)
	}
	defer dst.Close()
}
