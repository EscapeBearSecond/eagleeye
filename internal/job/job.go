package job

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/EscapeBearSecond/falcon/internal/global"
	"github.com/EscapeBearSecond/falcon/internal/mapper"
	"github.com/EscapeBearSecond/falcon/internal/mapper/vuln"
	"github.com/EscapeBearSecond/falcon/internal/stage"
	ptarget "github.com/EscapeBearSecond/falcon/internal/target"
	"github.com/EscapeBearSecond/falcon/internal/tpl"
	"github.com/EscapeBearSecond/falcon/internal/util"
	"github.com/EscapeBearSecond/falcon/internal/util/log"
	"github.com/EscapeBearSecond/falcon/pkg/types"
	"github.com/panjf2000/ants/v2"
	"github.com/projectdiscovery/nuclei/v3/pkg/catalog/disk"
	"github.com/projectdiscovery/nuclei/v3/pkg/output"
	"github.com/projectdiscovery/nuclei/v3/pkg/protocols"
	"github.com/projectdiscovery/nuclei/v3/pkg/protocols/common/contextargs"
	"github.com/projectdiscovery/nuclei/v3/pkg/scan"
	"github.com/projectdiscovery/ratelimit"
	"github.com/schollz/progressbar/v3"
)

type Job struct {
	name string
	kind string

	wg   sync.WaitGroup
	pool *ants.PoolWithFunc //goroutine池

	template     string
	getTemplates types.GetTemplates

	c       int
	limit   int
	timeout string
	retries int

	duration time.Duration

	ratelimit *ratelimit.Limiter // 限流器

	bar *progressbar.ProgressBar // 进度条

	pocs []*tpl.POC

	format string // 结果输出接口
	exp    exporter

	logger    *slog.Logger
	silent    bool
	outLogger *slog.Logger

	m         sync.Mutex
	cbResults []*types.JobResultItem
	callback  types.JobResultCallback
	entryID   string

	enableHeadless     bool
	skipHeadlessSize   int
	skipHeadlessReason string

	vulnMapper *vuln.Mapper
	directory  string

	stageManager *stage.Manager
	index        int

	completed *atomic.Int64
}

func (j *Job) Name() string {
	return j.name
}

func (j *Job) SetName(name string) {
	j.name = name
}

// New 实例化任务（一组模版的扫描行为）
func NewJob(opts ...Option) (*Job, error) {
	j := &Job{
		wg: sync.WaitGroup{},
	}

	for _, o := range opts {
		o(j)
	}

	err := j.init()
	if err != nil {
		return nil, fmt.Errorf("init job [%s] failed: %w", j.name, err)
	}

	return j, nil
}

func (j *Job) init() error {
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)

	if j.template == "" && j.getTemplates == nil {
		return fmt.Errorf("job [%s] templates are empty", j.name)
	}

	if j.callback != nil {
		j.cbResults = make([]*types.JobResultItem, 0)
	}

	if j.name == "" {
		j.name = util.RandomStr(10)
	}

	if j.silent {
		j.logger = log.Must(log.NewLogger(log.WithSilent(true)))
	} else {
		j.logger = log.Must(log.NewLogger(log.WithStdout()))
	}

	if j.limit > 0 {
		j.ratelimit = ratelimit.New(context.Background(), uint(j.limit), 1*time.Second)
	} else {
		j.ratelimit = ratelimit.NewUnlimited(context.Background())
	}

	duration, err := time.ParseDuration(j.timeout)
	if err != nil {
		return fmt.Errorf("invalid job [%s] timeout format: %w", j.name, err)
	}
	j.duration = duration

	exporter, err := newExporter(exportFormat(j.format), filepath.Join(j.directory, j.entryID, j.name))
	if err != nil {
		return fmt.Errorf("create job [%s] exporter failed: %w", j.name, err)
	}
	j.exp = exporter

	// 实例化对应函数的goroutine池
	{
		pool, err := ants.NewPoolWithFunc(j.c, func(i interface{}) {
			task := i.(*task)
			j.executePOCForTarget(task.c, task.poc, task.input)
		})
		if err != nil {
			return fmt.Errorf("create job [%s] routine pool failed: %w", j.name, err)
		}
		j.pool = pool
	}

	j.completed = &atomic.Int64{}

	return nil
}

// LoadTemplates 加载模板
func (j *Job) LoadTemplates(eOptions *protocols.ExecutorOptions) error {
	var result *tpl.Result
	var err error

	// 为execute options配置模版地址
	eOptions.Catalog = disk.NewCatalog(j.template)
	// 如果模版地址不为空，则使用文件路径中的模板
	if j.template != "" {
		result, err = tpl.LoadWithFileWalk(j.template, eOptions,
			tpl.WithEnableHeadless(j.enableHeadless))
	} else { // 否则使用getTemplates获取原始字符串模板（由于不可都为空）
		result, err = tpl.LoadWithFunc(j.getTemplates, eOptions,
			tpl.WithEnableHeadless(j.enableHeadless))
	}
	if err != nil {
		return fmt.Errorf("job [%s] templates load failed: %w", j.name, err)
	}

	if len(result.Pocs) == 0 {
		return fmt.Errorf("job [%s] templates load failed: %w", j.name, types.ErrInvalidTemplates)
	}

	j.skipHeadlessSize = result.SkipHeadlessSize
	j.skipHeadlessReason = result.SkipHeadlessReason
	j.pocs = result.Pocs
	return nil
}

// ExecuteWithContext 执行
func (j *Job) ExecuteWithContext(c context.Context, o *Options) error {
	j.logger.InfoContext(c, "Execute job start")

	err := j.executeWithContext(c, o)
	if err != nil {
		return err
	}
	j.logger.InfoContext(c, "Execute job end")

	j.doCallback(c)

	return nil
}

func (j *Job) doCallback(c context.Context) error {
	if j.callback != nil {
		err := j.callback(c, &types.JobResult{
			Name:    j.name,
			Kind:    j.kind,
			EntryID: j.entryID,
			Items:   j.cbResults,
		})
		if err != nil {
			return fmt.Errorf("job [%s] callback failed: %w", j.name, err)
		}
	}
	return nil
}

func (j *Job) executeWithContext(c context.Context, o *Options) error {
	defer j.close()

	if j.skipHeadlessSize > 0 {
		if len(j.pocs) == 0 {
			j.logger.InfoContext(c, "No Remaining Templates",
				"skip_headless_size", j.skipHeadlessSize,
				"skip_headless_reason", j.skipHeadlessReason,
			)
			return nil
		}
		j.logger.InfoContext(c, "Skip Headless Templates",
			"skip_headless_size", j.skipHeadlessSize,
			"skip_headless_reason", j.skipHeadlessReason,
		)
	}

	if len(j.pocs) == 0 {
		return fmt.Errorf("job [%s] pocs are empty", j.name)
	}

	total := int64(len(j.pocs)) * int64(len(o.Targets))
	// 进度条
	j.bar = util.NewProgressbar(j.name, total, j.silent)

	ok := make(chan struct{})
	defer close(ok)
	go j.progress(c, ok)

	for _, poc := range j.pocs {
		select {
		case <-c.Done():
			return context.Canceled
		default:
		}

		pocTimeout := j.duration
		if len(poc.RequestsJavascript) > 0 {
			pocTimeout = j.duration * 5
		}

		ports := poc.GetPorts()

		for _, target := range o.Targets {
			select {
			case <-c.Done():
				return context.Canceled
			default:
			}

			if ptarget.ShouldSkip(target, ports...) {
				j.completed.Add(1)
				continue
			}

			j.wg.Add(1)
			j.ratelimit.Take()

			j.pool.Invoke(
				newTask(
					context.WithValue(c, pocTimeoutKey, pocTimeout),
					poc,
					target,
				),
			)
		}
	}

	j.wg.Wait()

	return nil
}

func (j *Job) progress(c context.Context, ok <-chan struct{}) {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-c.Done():
			return
		case <-ok:
			j.bar.Finish()
			j.stageManager.Put(types.StageJob, 1, j.stageEntries()...)
			return
		case <-ticker.C:
			j.bar.Set64(j.completed.Load())
			j.stageManager.Put(types.StageJob, j.bar.State().CurrentPercent, j.stageEntries()...)
		}
	}
}

func (j *Job) stageEntries() []stage.Entry {
	entries := []stage.Entry{
		stage.NewEntry(types.StageEntryJobKind, j.kind),
		stage.NewEntry(types.StageEntryJobIndex, j.index),
	}
	return entries
}

// executePOCForTarget 工作单元（单poc 单地址）
func (j *Job) executePOCForTarget(c context.Context, poc *tpl.POC, input string) {
	defer func() {
		if r := recover(); r != nil {
			if j.outLogger != nil {
				j.outLogger.ErrorContext(c, "Execute Task Panic",
					"job_name", j.name,
					"template_id", poc.ID,
					"type", poc.Type().String(),
					"target", input,
					"reason", r,
				)
			}
		}
	}()

	defer j.wg.Done()
	defer j.completed.Add(1)

	inputs := make([]string, 0, 2)
	// 执行http预处理
	if len(poc.RequestsHTTP) > 0 || len(poc.RequestsHeadless) > 0 {
		schemes := make([]string, 0)

		// 如果input包含:80和:443的端口，则使用对应的scheme
		if util.IsHostPort(input) {
			_, port, _ := net.SplitHostPort(input)
			switch port {
			case "80":
				schemes = append(schemes, "http")
			case "443":
				schemes = append(schemes, "https")
			}
		}

		// 如果input不包含:80和:443的端口，则尝试metadata中的schemes
		if len(schemes) == 0 {
			// 处理metadata里的schemes
			schemes = append(schemes, poc.GetSchemes()...)
		}

		// 如果metadata中没有schemes，默认放入http和https
		if len(schemes) == 0 {
			schemes = append(schemes, "http", "https")
		}

		// 格式化url输入
		for _, scheme := range schemes {
			formedURL := fmt.Sprintf("%s://%s", scheme, input)
			inputs = append(inputs, formedURL)
		}
	} else {
		inputs = append(inputs, input)
	}

	timeout := c.Value(pocTimeoutKey).(time.Duration)

	var results []*output.ResultEvent
	var err error

retry:
	for range j.retries {
		select {
		case <-c.Done():
			return
		default:
		}
		for _, input := range inputs {
			select {
			case <-c.Done():
				return
			default:
			}

			results, err = func() ([]*output.ResultEvent, error) {
				cc, cancel := context.WithTimeout(c, timeout)
				defer cancel()

				var ctxErrors []error

				scanContext := scan.NewScanContext(cc,
					contextargs.NewWithInput(cc, input))
				scanContext.OnResult = func(event *output.InternalWrappedEvent) {}
				scanContext.OnError = func(err error) {
					ctxErrors = append(ctxErrors, err)
				}
				results, err := poc.Executer.ExecuteWithResults(
					scanContext,
				)
				// 1.判断execute是否有错，有错则返回错误
				if err != nil {
					return nil, err
				}
				// 2.判断execute结果是否为空，不为空则返回结果
				if len(results) != 0 {
					return results, nil
				}
				// 3.如果既没有错结果也为空，返回上下文错误
				if len(ctxErrors) != 0 {
					return nil, util.JoinErrors(ctxErrors)
				}
				// 4.如果上下文错误也为空，返回未知错误
				return nil, errors.New("unknown internal error")
			}()

			if err == nil {
				break retry
			}
		}
	}

	select {
	case <-c.Done():
		return
	default:
	}

	if err != nil {
		if j.outLogger != nil {
			j.outLogger.ErrorContext(c, "Execute Task Failed",
				"job_name", j.name,
				"template_id", poc.ID,
				"type", poc.Type().String(),
				"target", input,
				"reason", err.Error(),
			)
		}
		return
	}

	if j.outLogger != nil {
		j.outLogger.InfoContext(c, "Execute Task Success",
			"job_name", j.name,
			"template_id", poc.ID,
			"type", poc.Type().String(),
			"target", input,
		)
	}

	exists := make(map[string]bool)
	for _, result := range results {
		if result != nil {
			if exists[result.TemplateID] {
				continue
			}
			exists[result.TemplateID] = true

			if j.outLogger != nil {
				j.outLogger.InfoContext(c, "Execute Task Result Exist",
					"job_name", j.name,
					"template_id", result.TemplateID,
					"type", result.Type,
					"target", input,
				)
			}

			dests := []mapper.Dest{}
			// mappings
			{
				vulnMappings, err := j.vulnMapper.Get(result.TemplateID).By(result.ExtractedResults...)
				if err != nil {
					if j.outLogger != nil {
						j.outLogger.WarnContext(c, "Get Vulnerability Mappings Failed",
							"job_name", j.name,
							"template_id", result.TemplateID,
							"version", result.ExtractedResults[0],
							"reason", err.Error(),
						)
					}
					continue
				}

				dests = append(dests, vulnMappings...)
			}

			if global.UseSyncPool() {
				j.handleResultUseSyncPool(c, result, dests)
			} else {
				j.handleResult(c, result, dests)
			}
		} else {
			if j.outLogger != nil {
				j.outLogger.WarnContext(c, "Execute Task Result Empty",
					"job_name", j.name,
					"template_id", poc.ID,
					"type", poc.Type().String(),
					"target", input,
				)
			}
		}
	}
}

func (j *Job) handleResult(c context.Context, result *output.ResultEvent, dests []mapper.Dest) {
	results := make([]*types.JobResultItem, 0, len(dests)+1)
	if len(dests) != 0 {
		for _, dest := range dests {
			results = append(results, dest.Assign(j.exp.
				GetResult().
				WithEntryID(j.entryID).
				Fill(result)))
		}
	} else {
		results = append(results, j.exp.
			GetResult().
			WithEntryID(j.entryID).
			Fill(result))
	}

	if j.callback != nil {
		j.m.Lock()
		j.cbResults = append(j.cbResults, results...)
		j.m.Unlock()
	}

	if j.exp != nil {
		for _, r := range results {
			j.exp.Export(c, r)
		}
	}
}

func (j *Job) handleResultUseSyncPool(c context.Context, result *output.ResultEvent, dests []mapper.Dest) {
	if j.callback != nil {
		j.pushCbResult(c, result, dests)
	}

	if j.exp != nil {
		j.export(c, result, dests)
	}
}

func (j *Job) export(c context.Context, result *output.ResultEvent, dests []mapper.Dest) {
	results := make([]*types.JobResultItem, 0, len(dests)+1)
	if len(dests) != 0 {
		for _, dest := range dests {
			results = append(results, dest.Assign(j.exp.
				GetResult().
				Fill(result)))
		}
	} else {
		results = append(results, j.exp.
			GetResult().
			Fill(result))
	}

	for _, r := range results {
		j.exp.Export(c, r)
	}
}

func (j *Job) pushCbResult(_ context.Context, result *output.ResultEvent, dests []mapper.Dest) {
	results := make([]*types.JobResultItem, 0, len(dests)+1)
	if len(dests) != 0 {
		for _, dest := range dests {
			results = append(results, dest.Assign(
				types.NewJobResultItem().
					WithEntryID(j.entryID).
					Fill(result)))
		}
	} else {
		results = append(results,
			types.NewJobResultItem().
				WithEntryID(j.entryID).
				Fill(result))
	}

	j.m.Lock()
	j.cbResults = append(j.cbResults, results...)
	j.m.Unlock()
}

// close 关闭或停止相关对象
func (j *Job) close() {
	j.pool.Release()
	for j.pool.Running() != 0 {
		runtime.Gosched()
	}

	if j.exp != nil {
		j.exp.Stop()
	}

	if j.ratelimit != nil {
		j.ratelimit.Stop()
	}
}
