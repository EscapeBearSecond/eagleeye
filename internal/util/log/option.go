package log

import (
	"io"
	"os"
)

type options struct {
	writers   []io.Writer
	filename  string
	addSource bool
	json      bool
	silent    bool
}

type Option func(*options)

func WithStdout() Option {
	return func(o *options) {
		o.writers = append(o.writers, os.Stdout)
	}
}

func WithStderr() Option {
	return func(o *options) {
		o.writers = append(o.writers, os.Stderr)
	}
}

func WithSilent(silent bool) Option {
	return func(o *options) {
		o.silent = silent
	}
}

func WithFile(filename string) Option {
	return func(o *options) {
		o.filename = filename
	}
}

func WithAddSource(enable bool) Option {
	return func(o *options) {
		o.addSource = enable
	}
}

func WithJSON(enable bool) Option {
	return func(o *options) {
		o.json = enable
	}
}
