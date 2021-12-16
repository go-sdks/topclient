package signer

import (
	"bytes"
	"crypto/hmac"

	// nolint: gosec
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"net/url"
	"sort"
	"strings"
	"time"
)

// SignMethod hmac，md5，hmac-sha256
type SignMethod string

const (
	SignMethodHMac       = "hmac"
	SignMethodMD5        = "md5"
	SignMethodHMacSha256 = "hmac-sha256"
)

type Signer interface {
	Sign(values url.Values, method SignMethod) (err error)
}

func NewSinger(appKey, appSecret string) Signer {
	return &signImpl{
		appKey:    appKey,
		appSecret: []byte(appSecret),
	}
}

type signImpl struct {
	appKey    string
	appSecret []byte
}

func (impl *signImpl) Sign(values url.Values, signMethod SignMethod) (err error) {
	fnCheckField := func(field string) error {
		if len(values[field]) == 0 {
			return fmt.Errorf("miss field: %s", field)
		}

		return nil
	}

	err = fnCheckField("method")
	if err != nil {
		return
	}

	values.Del("sign")

	fnSetNE := func(key, value string) {
		if _, ok := values[key]; ok {
			return
		}

		values.Set(key, value)
	}

	fnSetNE("v", "2.0")
	fnSetNE("sign_method", string(signMethod))
	fnSetNE("app_key", impl.appKey)
	fnSetNE("timestamp", time.Now().Format("2006-01-02 15:04:05"))

	unsigned := impl.getUnsignedText(values)
	s, err := impl.doSign(signMethod, unsigned)

	if err != nil {
		return
	}

	values.Set("sign", s)

	return
}

func (impl *signImpl) sortKeys(values url.Values) []string {
	sortedKeys := make([]string, 0, len(values))
	for k := range values {
		sortedKeys = append(sortedKeys, k)
	}

	sort.Strings(sortedKeys)

	return sortedKeys
}

func (impl *signImpl) getUnsignedText(values url.Values) string {
	sortedKeys := impl.sortKeys(values)
	buf := bytes.Buffer{}

	for _, k := range sortedKeys {
		buf.WriteString(k)
		buf.WriteString(strings.Join(values[k], ","))
	}

	return buf.String()
}

func (impl *signImpl) doSign(signMethod SignMethod, unsignedText string) (sign string, err error) {
	var h hash.Hash

	switch signMethod {
	case SignMethodHMac:
		h = hmac.New(md5.New, impl.appSecret)
		_, err = h.Write([]byte(unsignedText))
	case SignMethodHMacSha256:
		h = hmac.New(sha256.New, impl.appSecret)
		_, err = h.Write([]byte(unsignedText))
	case SignMethodMD5:
		// nolint: gosec
		h = md5.New()
		_, err = h.Write(impl.appSecret)

		if err != nil {
			return
		}

		_, err = h.Write([]byte(unsignedText))
		if err != nil {
			return
		}

		_, err = h.Write(impl.appSecret)
	default:
		err = errors.New("unsupported hash algorithm")
	}

	if err != nil {
		return
	}

	sum := h.Sum(nil)
	sign = strings.ToUpper(hex.EncodeToString(sum))

	return
}
