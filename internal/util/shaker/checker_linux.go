//go:build linux
// +build linux

package shaker

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

type Checker struct {
	pipePool
	resultPipes
	pollerLock sync.Mutex
	_pollerFd  int32
	zeroLinger bool
	isReady    chan struct{}
}

func NewChecker() *Checker {
	return NewCheckerZeroLinger(true)
}

func NewCheckerZeroLinger(zeroLinger bool) *Checker {
	return &Checker{
		pipePool:    newPipePoolSyncPool(),
		resultPipes: newResultPipesSyncMap(),
		_pollerFd:   -1,
		zeroLinger:  zeroLinger,
		isReady:     make(chan struct{}),
	}
}

func (c *Checker) CheckingLoop(ctx context.Context) error {
	pollerFd, err := c.createPoller()
	if err != nil {
		return errors.Wrap(err, "error creating poller")
	}
	defer c.closePoller()

	c.setReady()
	defer c.resetReady()

	return c.pollingLoop(ctx, pollerFd)
}

func (c *Checker) createPoller() (int, error) {
	c.pollerLock.Lock()
	defer c.pollerLock.Unlock()

	if c.pollerFD() > 0 {
		return -1, ErrCheckerAlreadyStarted
	}

	pollerFd, err := createPoller()
	if err != nil {
		return -1, err
	}
	c.setPollerFD(pollerFd)

	return pollerFd, nil
}

func (c *Checker) closePoller() error {
	c.pollerLock.Lock()
	defer c.pollerLock.Unlock()
	var err error
	if c.pollerFD() > 0 {
		err = unix.Close(c.pollerFD())
	}
	c.setPollerFD(-1)
	return err
}

func (c *Checker) setReady() {
	close(c.isReady)
}

func (c *Checker) resetReady() {
	c.isReady = make(chan struct{})
}

const pollerTimeout = time.Second

func (c *Checker) pollingLoop(ctx context.Context, pollerFd int) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			evts, err := pollEvents(pollerFd, pollerTimeout)
			if err != nil {

				return errors.Wrap(err, "error during polling loop")
			}

			c.handlePollerEvents(evts)
		}
	}
}

func (c *Checker) handlePollerEvents(evts []event) {
	for _, e := range evts {
		if pipe, exists := c.resultPipes.popResultPipe(e.Fd); exists {
			pipe <- e.Err
		}
	}
}

func (c *Checker) pollerFD() int {
	return int(atomic.LoadInt32(&c._pollerFd))
}

func (c *Checker) setPollerFD(fd int) {
	atomic.StoreInt32(&c._pollerFd, int32(fd))
}

func (c *Checker) CheckAddr(addr string, timeout time.Duration) (err error) {
	return c.CheckAddrZeroLinger(addr, timeout, c.zeroLinger)
}

func (c *Checker) CheckAddrZeroLinger(addr string, timeout time.Duration, zeroLinger bool) error {
	deadline := time.Now().Add(timeout)

	rAddr, family, err := parseSockAddr(addr)
	if err != nil {
		return err
	}
	fd, err := createSocketZeroLinger(family, zeroLinger)
	if err != nil {
		return err
	}
	defer unix.Close(fd)

	if success, cErr := connect(fd, rAddr); cErr != nil {
		return &connectError{cErr}
	} else if success {

		return nil
	}

	return c.waitConnectResult(fd, deadline.Sub(time.Now()))
}

func (c *Checker) waitConnectResult(fd int, timeout time.Duration) error {

	resultPipe := c.getPipe()
	defer func() {
		c.resultPipes.deregisterResultPipe(fd)
		c.putBackPipe(resultPipe)
	}()

	c.resultPipes.registerResultPipe(fd, resultPipe)

	if err := registerEvents(c.pollerFD(), fd); err != nil {
		return err
	}

	return c.waitPipeTimeout(resultPipe, timeout)
}

func (c *Checker) waitPipeTimeout(pipe chan error, timeout time.Duration) error {
	select {
	case ret := <-pipe:
		return ret
	case <-time.After(timeout):
		return ErrTimeout
	}
}

func (c *Checker) WaitReady() <-chan struct{} {
	return c.isReady
}

func (c *Checker) IsReady() bool {
	return c.pollerFD() > 0
}

func (c *Checker) PollerFd() int {
	return c.pollerFD()
}
