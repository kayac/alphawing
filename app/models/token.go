package models

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/revel/revel"

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

func (token LimitedTimeToken) IsValid(seed, limitStr string) (bool, error) {
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return false, err
	}

	if int64(limit) < time.Now().Unix() {
		return false, nil
	}

	tokenValid, err := NewTokenBytes(seed, int64(limit))
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

func GetSecret() (string, error) {
	secret, found := revel.Config.String("app.secret")
	if !found {
		return "", fmt.Errorf("undefined config: app.secret")
	}
	return secret, nil
}

func NewTokenBytes(seed string, limit int64) ([]byte, error) {
	secret, err := GetSecret()
	if err != nil {
		return nil, err
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(seed))
	tokenBytes := mac.Sum([]byte(strconv.Itoa(int(limit))))

	return tokenBytes, nil
}

func NewEncodedToken(seed string, limit int64) (LimitedTimeToken, error) {
	tokenBytes, err := NewTokenBytes(seed, limit)
	if err != nil {
		return "", err
	}

	token := hex.EncodeToString(tokenBytes)

	return LimitedTimeToken(token), nil
}

func NewLimitedTimeTokenInfo() (*LimitedTimeTokenInfo, error) {
	u := uuid.NewRandom()
	seed := u.String()
	limit := time.Now().Add(TokenExpireDuration).Unix()

	token, err := NewEncodedToken(seed, limit)
	if err != nil {
		return nil, err
	}

	return &LimitedTimeTokenInfo{
		Token: token,
		Seed:  seed,
		Limit: limit,
	}, nil
}
