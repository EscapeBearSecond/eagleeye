//go:build linux
// +build linux

package shaker

import "sync"

type pipePool interface {
	getPipe() chan error
	putBackPipe(chan error)
}

type pipePoolDummy struct{}

func newPipePoolDummy() *pipePoolDummy {
	return &pipePoolDummy{}
}

func (*pipePoolDummy) getPipe() chan error {
	return make(chan error, 1)
}

func (*pipePoolDummy) putBackPipe(pipe chan error) {}

type pipePoolSyncPool struct {
	pool sync.Pool
}

func newPipePoolSyncPool() *pipePoolSyncPool {
	return &pipePoolSyncPool{sync.Pool{
		New: func() interface{} {
			return make(chan error, 1)
		}},
	}
}

func (p *pipePoolSyncPool) getPipe() chan error {
	return p.pool.Get().(chan error)
}

func (p *pipePoolSyncPool) putBackPipe(pipe chan error) {
	p.cleanPipe(pipe)
	p.pool.Put(pipe)
}

func (p *pipePoolSyncPool) cleanPipe(pipe chan error) {
	select {
	case <-pipe:
	default:
	}
}
