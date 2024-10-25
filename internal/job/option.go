package job

import (
	"log/slog"

	"github.com/EscapeBearSecond/falcon/internal/mapper/vuln"
	"github.com/EscapeBearSecond/falcon/internal/stage"
	"github.com/EscapeBearSecond/falcon/pkg/types"
)

type Option func(*Job)

// WithName 配置任务名称
func WithName(name string) Option {
	return func(j *Job) {
		j.name = name
	}
}

func WithIndex(index int) Option {
	return func(j *Job) {
		j.index = index
	}
}

// WithKind 配置任务类型
func WithKind(kind string) Option {
	return func(j *Job) {
		j.kind = kind
	}
}

// WithTemplate 配置任务模版
func WithTemplate(tpl string) Option {
	return func(j *Job) {
		j.template = tpl
	}
}

// WithGetTemplates 配置任务模版获取函数
func WithGetTemplates(getTemplates types.GetTemplates) Option {
	return func(j *Job) {
		j.getTemplates = getTemplates
	}
}

// WithConcurrency 配置任务并发数
func WithConcurrency(c int) Option {
	return func(j *Job) {
		j.c = c
	}
}

// WithRateLimit 配置任务频率
func WithRateLimit(limit int) Option {
	return func(j *Job) {
		j.limit = limit
	}
}

// WithExportFormat 配置任务输出格式
func WithExportFormat(format string) Option {
	return func(j *Job) {
		j.format = format
	}
}

// WithOutLogger 配置日志文件输出
func WithOutLogger(logger *slog.Logger) Option {
	return func(j *Job) {
		j.outLogger = logger
	}
}

// WithTimeout 配置超时时间
func WithTimeout(timeout string) Option {
	return func(j *Job) {
		j.timeout = timeout
	}
}

// WithRetries 配置重试次数
func WithRetries(count int) Option {
	return func(j *Job) {
		j.retries = count
	}
}

// WithCallback 配置回调
func WithCallback(callback types.JobResultCallback) Option {
	return func(j *Job) {
		j.callback = callback
	}
}

// WithEntryID 配置条目ID
func WithEntryID(id string) Option {
	return func(j *Job) {
		j.entryID = id
	}
}

// WithSilent 配置禁用进度(包括进度条和日志输出)
func WithSilent(silent bool) Option {
	return func(j *Job) {
		j.silent = silent
	}
}

func WithEnableHeadless(headless bool) Option {
	return func(j *Job) {
		j.enableHeadless = headless
	}
}

func WithVulnMapper(vm *vuln.Mapper) Option {
	return func(j *Job) {
		j.vulnMapper = vm
	}
}

func WithDirectory(dir string) Option {
	return func(j *Job) {
		j.directory = dir
	}
}

func WithStageManager(manager *stage.Manager) Option {
	return func(j *Job) {
		j.stageManager = manager
	}
}

type Options struct {
	Targets []string
}
