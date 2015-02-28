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
	TokenKey            = "token"
	SeedKey             = "seed"
	LimitKey            = "limit"
	TokenExpireDuration = 15 * time.Minute
)

type LimitedTimeTokenInfo struct {
	Token string
	Seed  string
	Limit int64
}

func (tokenInfo *LimitedTimeTokenInfo) UrlValues() *url.Values {
	v := &url.Values{}
	v.Add(TokenKey, string(tokenInfo.Token))
	v.Add(SeedKey, tokenInfo.Seed)
	v.Add(LimitKey, strconv.Itoa(int(tokenInfo.Limit)))

	return v
}

func (tokenInfo *LimitedTimeTokenInfo) IsValid(key string) (bool, error) {
	if tokenInfo.Limit < time.Now().Unix() {
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

func NewTokenBytes(seed string, limit int64, key string) ([]byte, error) {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(seed))
	tokenBytes := mac.Sum([]byte(strconv.Itoa(int(limit))))

	return tokenBytes, nil
}

func NewEncodedToken(seed string, limit int64, key string) (string, error) {
	tokenBytes, err := NewTokenBytes(seed, limit, key)
	if err != nil {
		return "", err
	}

	token := hex.EncodeToString(tokenBytes)

	return token, nil
}

func NewLimitedTimeTokenInfo(token, seed, limitStr string) (*LimitedTimeTokenInfo, error) {
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return nil, err
	}

	return &LimitedTimeTokenInfo{
		Token: token,
		Seed:  seed,
		Limit: int64(limit),
	}, nil
}

func NewLimitedTimeTokenInfoByKey(key string) (*LimitedTimeTokenInfo, error) {
	u := uuid.NewRandom()
	seed := u.String()
	limit := time.Now().Add(TokenExpireDuration).Unix()

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
