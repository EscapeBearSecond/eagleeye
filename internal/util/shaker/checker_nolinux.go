//go:build !linux
// +build !linux

package shaker

import (
	"context"
	"net"
	"time"
)

type Checker struct {
	zeroLinger bool
	isReady    chan struct{}
}

func NewChecker() *Checker {
	return NewCheckerZeroLinger(true)
}

func NewCheckerZeroLinger(zeroLinger bool) *Checker {
	isReady := make(chan struct{})
	close(isReady)
	return &Checker{zeroLinger: zeroLinger, isReady: isReady}
}

func (c *Checker) CheckingLoop(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (c *Checker) CheckAddr(addr string, timeout time.Duration) error {
	return c.CheckAddrZeroLinger(addr, timeout, c.zeroLinger)
}

func (c *Checker) CheckAddrZeroLinger(addr string, timeout time.Duration, zeroLinger bool) error {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if conn != nil {
		if zeroLinger {

			conn.(*net.TCPConn).SetLinger(0)
		}
		conn.Close()
	}
	if opErr, ok := err.(*net.OpError); ok {
		if opErr.Timeout() {
			return ErrTimeout
		}
	}
	return err
}

func (c *Checker) IsReady() bool { return true }

func (c *Checker) WaitReady() <-chan struct{} {
	return c.isReady
}

func (c *Checker) Close() error { return nil }
