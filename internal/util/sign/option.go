package sign

import "fmt"

type options struct {
	secret string
	kvs    []string
}

type Option func(*options)

func Secret(secret string) Option {
	return func(o *options) {
		o.secret = secret
	}
}

func KeyValue(key, value string) Option {
	return func(o *options) {
		o.kvs = append(o.kvs, fmt.Sprintf("%s=%s", key, value))
	}
}
