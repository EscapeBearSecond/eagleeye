package apiserver

import (
	"log/slog"

	"github.com/EscapeBearSecond/falcon/internal/util/log"
	eagleeye "github.com/EscapeBearSecond/falcon/pkg/sdk"
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
