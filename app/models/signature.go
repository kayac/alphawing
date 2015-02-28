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
	SignatureKey = "signature"
	TokenKey     = "token"
	LimitKey     = "limit"

	TokenExpireDuration = 15 * time.Minute
)

type LimitedTimeSignatureInfo struct {
	Signature string
	Token     string
	Limit     string
}

func (signatureInfo *LimitedTimeSignatureInfo) UrlValues() *url.Values {
	v := &url.Values{}
	v.Add(SignatureKey, signatureInfo.Signature)
	v.Add(TokenKey, signatureInfo.Token)
	v.Add(LimitKey, signatureInfo.Limit)

	return v
}

func (signatureInfo *LimitedTimeSignatureInfo) IsExpired() (bool, error) {
	limit, err := strconv.ParseInt(signatureInfo.Limit, 10, 64)
	if err != nil {
		return false, err
	}

	return limit < time.Now().Unix(), nil
}

func (signatureInfo *LimitedTimeSignatureInfo) IsValid(key string) (bool, error) {
	expired, err := signatureInfo.IsExpired()
	if err != nil {
		return false, err
	}
	if expired {
		return false, nil
	}

	signatureValid, err := NewSignatureBytes(signatureInfo.Token, signatureInfo.Limit, key)
	if err != nil {
		return false, err
	}

	signature, err := hex.DecodeString(signatureInfo.Signature)
	if err != nil {
		return false, err
	}

	return hmac.Equal(signature, signatureValid), nil
}

func NewSignatureBytes(token, limit, key string) ([]byte, error) {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(token))

	return mac.Sum([]byte(limit)), nil
}

func NewEncodedSigature(token, limit, key string) (string, error) {
	signatureBytes, err := NewSignatureBytes(token, limit, key)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(signatureBytes), nil
}

func NewLimitedTimeSignatureInfo(signature, token, limit string) (*LimitedTimeSignatureInfo, error) {
	return &LimitedTimeSignatureInfo{
		Signature: signature,
		Token:     token,
		Limit:     limit,
	}, nil
}

func NewLimitedTimeSignatureInfoByKey(key string) (*LimitedTimeSignatureInfo, error) {
	token := uuid.NewRandom().String()
	limit := strconv.FormatInt(time.Now().Add(TokenExpireDuration).Unix(), 10)

	signature, err := NewEncodedSigature(token, limit, key)
	if err != nil {
		return nil, err
	}

	return &LimitedTimeSignatureInfo{
		Signature: signature,
		Token:     token,
		Limit:     limit,
	}, nil
}
