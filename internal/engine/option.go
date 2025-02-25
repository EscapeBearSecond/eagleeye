package engine

import (
	"github.com/EscapeBearSecond/falcon/internal/job"
	"github.com/EscapeBearSecond/falcon/internal/scanner"
	"github.com/EscapeBearSecond/falcon/internal/stage"
)

type Option func(*Engine)

// WithTargets 配置目标
func WithTargets(targets []string) Option {
	return func(e *Engine) {
		e.targets = append(e.targets, targets...)
	}
}

// WithPortScanner 配置端口扫描
func WithPortScanner(sc scanner.Scanner[[]string, []string]) Option {
	return func(e *Engine) {
		e.portScanner = sc
	}
}

// WithHostDiscoverer 配置在线监测
func WithHostDiscoverer(sc scanner.Scanner[[]string, []string]) Option {
	return func(e *Engine) {
		e.hostDiscoverer = sc
	}
}

// WithJobs 配置任务
func WithJobs(jobs ...*job.Job) Option {
	return func(e *Engine) {
		e.jobs = append(e.jobs, jobs...)
	}
}

// WithDisableBanner 禁用banner
func WithDisableBanner(disable bool) Option {
	return func(e *Engine) {
		e.disableBanner = disable
	}
}

func WithExcludeTargets(targets []string) Option {
	return func(e *Engine) {
		e.excludeTargets = append(e.excludeTargets, targets...)
	}
}

func WithStageManager(manager *stage.Manager) Option {
	return func(e *Engine) {
		e.stageManager = manager
	}
}
