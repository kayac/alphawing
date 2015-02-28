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

type LimitedTimeToken string

func (token LimitedTimeToken) Decode() ([]byte, error) {
	return hex.DecodeString(string(token))
}

func (token LimitedTimeToken) IsValid(seed, limitStr, key string) (bool, error) {
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return false, err
	}

	if int64(limit) < time.Now().Unix() {
		return false, nil
	}

	tokenValid, err := NewTokenBytes(seed, int64(limit), key)
	if err != nil {
		return false, err
	}

	tokenDecoded, err := token.Decode()
	if err != nil {
		return false, err
	}

	if hmac.Equal(tokenDecoded, tokenValid) {
		return true, nil
	}
	return false, nil
}

type LimitedTimeTokenInfo struct {
	Token LimitedTimeToken
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

func NewTokenBytes(seed string, limit int64, key string) ([]byte, error) {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(seed))
	tokenBytes := mac.Sum([]byte(strconv.Itoa(int(limit))))

	return tokenBytes, nil
}

func NewEncodedToken(seed string, limit int64, key string) (LimitedTimeToken, error) {
	tokenBytes, err := NewTokenBytes(seed, limit, key)
	if err != nil {
		return "", err
	}

	token := hex.EncodeToString(tokenBytes)

	return LimitedTimeToken(token), nil
}

func NewLimitedTimeTokenInfo(key string) (*LimitedTimeTokenInfo, error) {
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
