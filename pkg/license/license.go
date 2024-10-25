package license

import (
	"io"

	"github.com/EscapeBearSecond/falcon/internal/license"
)

type Watcher license.Watcher
type License license.License

// Watch 监听license
func Watch(path string) (*Watcher, error) {
	w, err := license.Watch(path)
	if err != nil {
		return nil, err
	}
	return (*Watcher)(w), nil
}

// Stop 停止监听
func (w *Watcher) Stop() {
	(*license.Watcher)(w).Stop()
}

func VerifyFromReader(reader io.Reader) error {
	l, err := license.ParseFromReader(reader)
	if err != nil {
		return err
	}

	return l.Verify()
}

// Verify 验证证书
func Verify() error {
	return license.Verify()
}

// L 获取当前license（拷贝副本）
func L() *License {
	return (*License)(license.L())
}

func UseSecret(str string) {
	license.UseSecret(str)
}
