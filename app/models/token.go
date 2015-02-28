package models

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strconv"
	"time"

	"code.google.com/p/go-uuid/uuid"
)

const (
	TokenKey = "token"
	SeedKey  = "seed"
	LimitKey = "limit"

	TokenExpireDuration = 15 * time.Minute
)

type LimitedTimeTokenInfo struct {
	Token string
	Seed  string
	Limit string
}

func (tokenInfo *LimitedTimeTokenInfo) UrlValues() *url.Values {
	v := &url.Values{}
	v.Add(TokenKey, tokenInfo.Token)
	v.Add(SeedKey, tokenInfo.Seed)
	v.Add(LimitKey, tokenInfo.Limit)

	return v
}

func (tokenInfo *LimitedTimeTokenInfo) IsExpired() (bool, error) {
	limit, err := strconv.ParseInt(tokenInfo.Limit, 10, 64)
	if err != nil {
		return false, err
	}

	return limit < time.Now().Unix(), nil
}

func (tokenInfo *LimitedTimeTokenInfo) IsValid(key string) (bool, error) {
	expired, err := tokenInfo.IsExpired()
	if err != nil {
		return false, err
	}
	if expired {
		return false, nil
	}

	tokenValid, err := NewTokenBytes(tokenInfo.Seed, tokenInfo.Limit, key)
	if err != nil {
		return false, err
	}

	tokenDecoded, err := hex.DecodeString(tokenInfo.Token)
	if err != nil {
		return false, err
	}

	return hmac.Equal(tokenDecoded, tokenValid), nil
}

func NewTokenBytes(seed, limit, key string) ([]byte, error) {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(seed))
	tokenBytes := mac.Sum([]byte(limit))

	return tokenBytes, nil
}

func NewEncodedToken(seed, limit, key string) (string, error) {
	tokenBytes, err := NewTokenBytes(seed, limit, key)
	if err != nil {
		return "", err
	}

	token := hex.EncodeToString(tokenBytes)

	return token, nil
}

func NewLimitedTimeTokenInfo(token, seed, limit string) (*LimitedTimeTokenInfo, error) {
	return &LimitedTimeTokenInfo{
		Token: token,
		Seed:  seed,
		Limit: limit,
	}, nil
}

func NewLimitedTimeTokenInfoByKey(key string) (*LimitedTimeTokenInfo, error) {
	u := uuid.NewRandom()
	seed := u.String()

	limit := strconv.FormatInt(time.Now().Add(TokenExpireDuration).Unix(), 10)

	token, err := NewEncodedToken(seed, limit, key)
	if err != nil {
		return nil, err
	}

	return &LimitedTimeTokenInfo{
		Token: token,
		Seed:  seed,
		Limit: limit,
	}, nil
}
