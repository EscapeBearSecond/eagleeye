package log

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

// NewLogger 配置输出目的地（支持stdout/stderr/file）
func NewLogger(opts ...Option) (*slog.Logger, error) {
	o := options{
		writers: make([]io.Writer, 0, 3),
	}
	for _, opt := range opts {
		opt(&o)
	}

	if o.filename != "" {
		file, err := os.Create(o.filename)
		if err != nil {
			return nil, fmt.Errorf("create log file failed: %w", err)
		}
		o.writers = append(o.writers, file)
	}

	var writer io.Writer
	if len(o.writers) == 0 {
		writer = os.Stdout
	} else {
		writer = io.MultiWriter(o.writers...)
	}

	if o.silent {
		writer = io.Discard
	}

	slogOptions := &slog.HandlerOptions{
		AddSource: o.addSource,
	}

	var slogHandler slog.Handler
	if !o.json {
		slogHandler = slog.NewTextHandler(writer, slogOptions)
	} else {
		slogHandler = slog.NewJSONHandler(writer, slogOptions)
	}

	return slog.New(slogHandler), nil
}

// Must 忽略error
//
// 警告：请在确定没有error时使用
func Must(l *slog.Logger, err error) *slog.Logger {
	return l
}
