package models

import "github.com/DHowett/go-plist"

const (
	AssetKind                = "software-package"
	MetadataBundleIdentifier = "com.example.test"
	MetadataKind             = "software"
)

type Plist struct {
	Items []*Item `plist:"items"`
}

type Item struct {
	Assets   []*Asset  `plist:"assets"`
	Metadata *Metadata `plist:"metadata"`
}

type Asset struct {
	Kind string `plist:"kind"`
	Url  string `plist:"url"`
}

type Metadata struct {
	BundleIdentifier string `plist:"bundle-identifier"`
	BundleVersion    string `plist:"bundle-version"`
	Kind             string `plist:"kind"`
	Title            string `plist:"title"`
}

func NewPlist(title, version, ipaUrl string) *Plist {
	return &Plist{
		Items: []*Item{
			&Item{
				Assets: []*Asset{
					&Asset{
						Kind: AssetKind,
						Url:  ipaUrl,
					},
				},
				Metadata: &Metadata{
					BundleIdentifier: MetadataBundleIdentifier,
					BundleVersion:    version,
					Kind:             MetadataKind,
					Title:            title,
				},
			},
		},
	}
}

func (p *Plist) Marshall() ([]byte, error) {
	return plist.MarshalIndent(p, plist.XMLFormat, "\t")
}
