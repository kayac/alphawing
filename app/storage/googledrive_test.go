package storage

import "testing"

func TestSatisfyGoogleDrive(t *testing.T) {
	g := GoogleDrive{}

	_, ok := interface{}(g).(Storage)

	if !ok {
		t.Error("storage.GoogleDrive is not satisfy of storage.Storage")
	}
}
