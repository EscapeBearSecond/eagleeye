package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

func Sign(opts ...Option) (string, error) {
	var o options
	o.kvs = make([]string, 0)
	for _, option := range opts {
		option(&o)
	}

	if o.secret == "" {
		return "", errors.New("secret is empty")
	}

	hm := hmac.New(
		sha256.New,
		[]byte(o.secret))

	for _, kv := range o.kvs {
		hm.Write([]byte(kv))
	}

	hBytes := hm.Sum(nil)
	return hex.EncodeToString(hBytes), nil
}
