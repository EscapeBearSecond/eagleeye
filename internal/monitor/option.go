package monitor

import (
	"context"
	"log/slog"
)

// Option 监控option
type Option interface {
	apply(*Monitor)
}

// WithMemory 开启内存监控
func WithMemory() Option {
	return &memory{}
}

type memory struct{}

func (mem memory) apply(m *Monitor) {
	m.mem = true
}

// WithCPU 开启cpu监控
func WithCPU() Option {
	return &cpu{}
}

type cpu struct{}

func (cpu cpu) apply(m *Monitor) {
	m.cpu = true
}

// WithImmediately 立即执行
func WithImmediately(c context.Context) Option {
	return &immediately{c: c}
}

type immediately struct {
	c context.Context
}

func (i immediately) apply(m *Monitor) {
	m.startImmed = true
	m.c = i.c
}

// WithInterval 监控周期
func WithInterval(i string) Option {
	return &interval{interval: i}
}

type interval struct {
	interval string
}

func (i interval) apply(m *Monitor) {
	m.interval = i.interval
}

// WithLogger 日志输出
func WithLogger(l *slog.Logger) Option {
	return &logger{logger: l}
}

type logger struct {
	logger *slog.Logger
}

func (i logger) apply(m *Monitor) {
	m.logger = i.logger
}

// WithNetwork 开启网卡监控
// func WithNetwork(etherNum int) Option {
// 	return &network{network: true, etherNum: etherNum}
// }

// type network struct {
// 	network  bool
// 	etherNum int
// }

// func (n network) apply(m *Monitor) {
// 	m.network = n.network
// 	m.etherNum = n.etherNum
// }

// WithNetwork 开启网卡监控
func WithDraw(name string) Option {
	return &draw{name: name}
}

type draw struct {
	name string
}

func (d draw) apply(m *Monitor) {
	m.draw = true
	m.drawName = d.name
}
