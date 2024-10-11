//go:build linux
// +build linux

package shaker

type event struct {
	Fd  int
	Err error
}
