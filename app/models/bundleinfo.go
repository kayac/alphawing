package models

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"os"
	"strings"

	"github.com/DHowett/go-plist"
	"github.com/shogo82148/androidbinary"
)

// a BundleInfo is information of an application package(apk file, ipa file, etc.)
type BundleInfo struct {
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

type BundleParseError struct {
	Offset int64
}

func (e *BundleParseError) Error() string {
	return "cannot parse application package file"
}

func NewBundleInfo(file *os.File, platformType BundlePlatformType) (*BundleInfo, error) {
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
		bundleInfo, err := parseApkFile(xmlFile)
		return bundleInfo, err
	}

	// parse an ipa file
	if platformType == BundlePlatformTypeIOS {
		bundleInfo, err := parseIpaFile(plistFile)
		return bundleInfo, err
	}

	return nil, errors.New("unknown platform")
}

func parseApkFile(xmlFile *zip.File) (*BundleInfo, error) {
	if xmlFile == nil {
		return nil, errors.New("AndroidManifest.xml is not found")
	}

	manifest, err := parseAndroidManifest(xmlFile)
	if err != nil {
		return nil, err
	}

	bundleInfo := &BundleInfo{}
	bundleInfo.Version = manifest.VersionName
	bundleInfo.PlatformType = BundlePlatformTypeAndroid

	return bundleInfo, nil
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

func parseIpaFile(plistFile *zip.File) (*BundleInfo, error) {
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

	bundleInfo := &BundleInfo{}
	bundleInfo.Version = info.CFBundleVersion
	bundleInfo.PlatformType = BundlePlatformTypeIOS

	return bundleInfo, nil
}
