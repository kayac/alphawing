package models

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strconv"
	"time"

	"github.com/pborman/uuid"
)

const (
	SignatureExpireDuration      = 15 * time.Minute
	SignaturePermittedHttpMethod = "GET"
)

type ParamToSign struct {
	Method string
	Host   string
	Path   string
	Token  string
	Limit  string
}

func (param *ParamToSign) String() string {
	return param.Method + "\n" +
		param.Host + "\n" +
		param.Path + "\n" +
		param.Token + "\n" +
		param.Limit
}

type LimitedTimeSignatureInfo struct {
	Signature   string
	ParamToSign *ParamToSign
}

func (signatureInfo *LimitedTimeSignatureInfo) signatureBytes(key string) []byte {
	hash := hmac.New(sha256.New, []byte(key))
	hash.Write([]byte(signatureInfo.ParamToSign.String()))

	return hash.Sum(nil)
}

func (signatureInfo *LimitedTimeSignatureInfo) encodedSigature(key string) string {
	signatureBytes := signatureInfo.signatureBytes(key)

	return hex.EncodeToString(signatureBytes)
}

func (signatureInfo *LimitedTimeSignatureInfo) RefreshSignature(key string) {
	signatureInfo.Signature = signatureInfo.encodedSigature(key)

	return
}

func (signatureInfo *LimitedTimeSignatureInfo) UrlValues() *url.Values {
	v := &url.Values{}
	v.Add("signature", signatureInfo.Signature)
	v.Add("token", signatureInfo.ParamToSign.Token)
	v.Add("limit", signatureInfo.ParamToSign.Limit)

	return v
}

func (signatureInfo *LimitedTimeSignatureInfo) IsExpired() (bool, error) {
	limit, err := strconv.ParseInt(signatureInfo.ParamToSign.Limit, 10, 64)
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

	signatureValid := signatureInfo.signatureBytes(key)

	signature, err := hex.DecodeString(signatureInfo.Signature)
	if err != nil {
		return false, err
	}

	return hmac.Equal(signature, signatureValid), nil
}

func NewLimitedTimeSignatureInfo(host, path string) *LimitedTimeSignatureInfo {
	return &LimitedTimeSignatureInfo{
		ParamToSign: &ParamToSign{
			Method: SignaturePermittedHttpMethod,
			Host:   host,
			Path:   path,
			Token:  uuid.NewRandom().String(),
			Limit:  strconv.FormatInt(time.Now().Add(SignatureExpireDuration).Unix(), 10),
		},
	}
}
