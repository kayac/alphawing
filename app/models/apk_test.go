package models

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// e.g: TEST_APK=1 TEST_APK_PATH=/path/to go test apk_test.go apk.go -v
func TestParseApk(t *testing.T) {
	if os.Getenv("TEST_APK") == "" {
		// skip
		return
	}

	apkPath := os.Getenv("TEST_APK_PATH")
	if apkPath == "" {
		t.Fatal("require env {TEST_APK_PATH}", apkPath)
	}

	whichResultByte, err := exec.Command("which", "aapt").Output()
	if err != nil {
		t.Fatal(err)
	}
	aaptPath := strings.TrimRight(string(whichResultByte), "\n")

	file, err := os.OpenFile(apkPath, os.O_RDONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	apk, err := NewApk(file, aaptPath)
	if err != nil {
		t.Fatal(err)
	}

	log.Printf("%+v", apk)
}
