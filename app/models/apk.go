package models

import (
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var aaptPath string

type Apk struct {
	Name    string
	NameJa  string
	Version string
}

type ApkParseError struct {
	Offset int64
}

func (e *ApkParseError) Error() string {
	return "Can't parse apk file."
}

func NewApk(file *os.File, path string) (*Apk, error) {
	if file == nil {
		return nil, errors.New("File is required.")
	}
	apk := &Apk{}
	if aaptPath == "" {
		if err := apk.setAaptPath(path); err != nil {
			return nil, err
		}
	}
	if err := apk.setApkInfo(file); err != nil {
		return nil, err
	}
	return apk, nil
}

func (apk *Apk) aapt(args ...string) (string, error) {
	cmd := exec.Command(aaptPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), err
}

func (apk *Apk) setAaptPath(path string) error {
	_, err := exec.LookPath(path)
	if err != nil {
		return err
	}
	aaptPath = path
	return nil
}

func (apk *Apk) setApkInfo(file *os.File) error {
	dump, _ := apk.aapt("d", "badging", file.Name())
	hash, err := apk.parseDumpBadging(dump)
	if err != nil {
		return err
	}
	apk.Name = hash["application-label"]
	apk.NameJa = hash["application-label-ja"]
	apk.Version = hash["versionName"]
	return nil
}

func (apk *Apk) parseDumpBadging(dump string) (map[string]string, error) {
	hash := make(map[string]string)
	var err error
	reg_package, err := regexp.Compile(`^package:`)
	if err != nil {
		return nil, err
	}
	reg_package_version_name, err := regexp.Compile(`versionName='(.+?)'`)
	if err != nil {
		return nil, err
	}
	reg_label, err := regexp.Compile(`^application-label:'(.+?)'`)
	if err != nil {
		return nil, err
	}
	reg_label_ja, err := regexp.Compile(`^application-label-ja:'(.+?)'`)
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(dump, "\n") {
		if reg_package.MatchString(line) {
			ret := reg_package_version_name.FindStringSubmatch(line)
			hash["versionName"] = ret[1]
		}
		if reg_label.MatchString(line) {
			ret := reg_label.FindStringSubmatch(line)
			hash["application-label"] = ret[1]
		}
		if reg_label_ja.MatchString(line) {
			ret := reg_label_ja.FindStringSubmatch(line)
			hash["application-label-ja"] = ret[1]
		}
	}
	return hash, err
}
