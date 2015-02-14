package models

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"github.com/DHowett/go-plist"
	"github.com/shogo82148/androidbinary"
	"io/ioutil"
	"os"
	"strings"
)

// an AppInfo is information of an application package(apk file, ipa file, etc.)
type AppInfo struct {
	Version      string
	PlatformType BundlePlatformType
}

type androidManifest struct {
	XMLName     xml.Name `xml:"manifest"`
	VersionName string   `xml:"http://schemas.android.com/apk/res/android versionName,attr"`
}

type iosInfo struct {
	CFBundleVersion string `plist:"CFBundleVersion"`
}

type AppParseError struct {
	Offset int64
}

func (e *AppParseError) Error() string {
	return "cannot parse application package file"
}

func NewAppInfo(file *os.File, platformType BundlePlatformType) (*AppInfo, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	reader, err := zip.NewReader(file, stat.Size())
	if err != nil {
		return nil, err
	}

	// search system files
	var xmlFile *zip.File   // apk system file
	var plistFile *zip.File // ipa system file
	for _, f := range reader.File {
		switch {
		case f.Name == "AndroidManifest.xml":
			xmlFile = f
		case strings.HasSuffix(f.Name, "/Info.plist"):
			plistFile = f
		}
	}

	// parse an apk file
	if platformType == BundlePlatformTypeAndroid {
		appInfo, err := parseApkFile(xmlFile)
		return appInfo, err
	}

	// parse an ipa file
	if platformType == BundlePlatformTypeIOS {
		appInfo, err := parseIpaFile(plistFile)
		return appInfo, err
	}

	return nil, errors.New("unknown platform")
}

func parseApkFile(xmlFile *zip.File) (*AppInfo, error) {
	if xmlFile == nil {
		return nil, errors.New("AndroidManifest.xml is not found")
	}

	manifest, err := parseAndroidManifest(xmlFile)
	if err != nil {
		return nil, err
	}

	appInfo := &AppInfo{}
	appInfo.Version = manifest.VersionName
	appInfo.PlatformType = BundlePlatformTypeAndroid

	return appInfo, nil
}

func parseAndroidManifest(xmlFile *zip.File) (*androidManifest, error) {
	rc, err := xmlFile.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	buf, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	xmlContent, err := androidbinary.NewXMLFile(bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}

	decoder := xml.NewDecoder(xmlContent.Reader())
	manifest := &androidManifest{}
	if err := decoder.Decode(manifest); err != nil {
		return nil, err
	}

	return manifest, nil
}

func parseIpaFile(plistFile *zip.File) (*AppInfo, error) {
	if plistFile == nil {
		return nil, errors.New("info.plist is not found")
	}

	rc, err := plistFile.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	buf, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	info := &iosInfo{}
	_, err = plist.Unmarshal(buf, info)
	if err != nil {
		return nil, err
	}

	appInfo := &AppInfo{}
	appInfo.Version = info.CFBundleVersion
	appInfo.PlatformType = BundlePlatformTypeIOS

	return appInfo, nil
}
