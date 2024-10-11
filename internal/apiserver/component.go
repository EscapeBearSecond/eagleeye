package apiserver

import (
	"log/slog"

	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/internal/util/log"
	eagleeye "codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/pkg/sdk"
)

var (
	Eagleeye *eagleeye.EagleeyeEngine
	Logger   *slog.Logger
)

func InitComponent() error {
	var err error
	Eagleeye, err = eagleeye.NewEngine()
	if err != nil {
		return err
	}

	Logger = log.Must(
		log.NewLogger(log.WithStdout(),
			log.WithJSON(true),
			log.WithAddSource(true),
		))

	return nil
}

func ReleaseComponent() {
	Eagleeye.Close()
}
