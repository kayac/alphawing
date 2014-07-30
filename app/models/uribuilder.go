package models

import (
	"net/url"
)

type UriBuilder interface {
	UriFor(string) (*url.URL, error)
}
