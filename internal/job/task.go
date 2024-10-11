package job

import (
	"context"

	"github.com/EscapeBearSecond/eagleeye/internal/tpl"
)

type contextKey string

const pocTimeoutKey contextKey = "poc_timeout"

// task 最小任务单元
type task struct {
	c     context.Context
	poc   *tpl.POC
	input string
}

// 实例化task
func newTask(c context.Context, poc *tpl.POC, input string) *task {
	return &task{
		c:     c,
		poc:   poc,
		input: input,
	}
}
