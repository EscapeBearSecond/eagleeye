package util

import (
	"fmt"
	"io"
	"os"

	"github.com/schollz/progressbar/v3"
)

func NewProgressbar(name string, size int64, silent ...bool) *progressbar.ProgressBar {
	var writer io.Writer = os.Stdout
	if len(silent) > 0 && silent[0] {
		writer = io.Discard
	}

	options := []progressbar.Option{
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetElapsedTime(true),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetDescription(fmt.Sprintf("[%s]", name)),
		progressbar.OptionSetItsString("req"),
		progressbar.OptionShowIts(),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "#",
			AltSaucerHead: "$",
			SaucerHead:    "$",
			SaucerPadding: "·",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(writer, "\n")
		}),
		progressbar.OptionSetWriter(writer),
	}

	return progressbar.NewOptions64(size, options...)
}
