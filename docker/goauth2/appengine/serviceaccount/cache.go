// Copyright 2013 The goauth2 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build appengine

package serviceaccount

import (
	"time"

	"appengine"
	"appengine/memcache"

	"code.google.com/p/goauth2/oauth"
)

// cache implements TokenCache using memcache to store AccessToken
// for the application service account.
type cache struct {
	Context appengine.Context
	Key     string
}

func (m cache) Token() (*oauth.Token, error) {
	tok := new(oauth.Token)
	_, err := memcache.Gob.Get(m.Context, m.Key, tok)
	if err != nil {
		return nil, err
	}
	return tok, nil
}

func (m cache) PutToken(tok *oauth.Token) error {
	return memcache.Gob.Set(m.Context, &memcache.Item{
		Key: m.Key,
		Object: oauth.Token{
			AccessToken: tok.AccessToken,
			Expiry:      tok.Expiry,
		},
		Expiration: tok.Expiry.Sub(time.Now()),
	})
}
