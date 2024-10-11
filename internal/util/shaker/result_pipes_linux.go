//go:build linux
// +build linux

package shaker

import "sync"

type resultPipes interface {
	popResultPipe(int) (chan error, bool)
	deregisterResultPipe(int)
	registerResultPipe(int, chan error)
}

type resultPipesMU struct {
	l             sync.Mutex
	fdResultPipes map[int]chan error
}

func newResultPipesMU() *resultPipesMU {
	return &resultPipesMU{fdResultPipes: make(map[int]chan error)}
}

func (r *resultPipesMU) popResultPipe(fd int) (chan error, bool) {
	r.l.Lock()
	p, exists := r.fdResultPipes[fd]
	if exists {
		delete(r.fdResultPipes, fd)
	}
	r.l.Unlock()
	return p, exists
}

func (r *resultPipesMU) deregisterResultPipe(fd int) {
	r.l.Lock()
	delete(r.fdResultPipes, fd)
	r.l.Unlock()
}

func (r *resultPipesMU) registerResultPipe(fd int, pipe chan error) {
	r.l.Lock()
	r.fdResultPipes[fd] = pipe
	r.l.Unlock()
}

type resultPipesSyncMap struct {
	sync.Map
}

func newResultPipesSyncMap() *resultPipesSyncMap {
	return &resultPipesSyncMap{}
}

func (r *resultPipesSyncMap) popResultPipe(fd int) (chan error, bool) {
	p, exist := r.Load(fd)
	if exist {
		r.Delete(fd)
	}
	if p != nil {
		return p.(chan error), exist
	}
	return nil, exist
}

func (r *resultPipesSyncMap) deregisterResultPipe(fd int) {
	r.Delete(fd)
}

func (r *resultPipesSyncMap) registerResultPipe(fd int, pipe chan error) {
	r.Store(fd, pipe)
}
