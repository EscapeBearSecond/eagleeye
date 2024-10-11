package eagleeye

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	core "codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/internal/engine"
	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/internal/global"
	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/internal/job"
	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/internal/mapper/vuln"
	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/internal/scanner"
	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/internal/stage"
	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/pkg/types"
	"github.com/rs/xid"
	"github.com/samber/lo"
)

// EagleeyeEngine sdk引擎
type EagleeyeEngine struct {
	entries []*EagleeyeEntry
	m       sync.RWMutex
	dir     string
}

// NewEngine 创建sdk引擎
func NewEngine(options ...Option) (*EagleeyeEngine, error) {
	if err := global.Init(); err != nil {
		return nil, fmt.Errorf("create eagleeye engine failed: %w", err)
	}

	engine := &EagleeyeEngine{
		entries: make([]*EagleeyeEntry, 0),
		m:       sync.RWMutex{},
	}

	for _, option := range options {
		option.apply(engine)
	}

	if engine.dir == "" {
		engine.dir = "."
	}

	if engine.dir != "." {
		err := os.MkdirAll(engine.dir, 0755)
		if err != nil {
			return nil, fmt.Errorf("create results directory failed: %w", err)
		}
	}

	return engine, nil
}

// Entry 获取条目
func (e *EagleeyeEngine) Entry(entryID string) *EagleeyeEntry {
	e.m.RLock()
	defer e.m.RUnlock()
	entry, _ := lo.Find(e.entries, func(entry *EagleeyeEntry) bool {
		return entry.EntryID == entryID
	})
	return entry
}

func (e *EagleeyeEngine) deleteEntry(entry *EagleeyeEntry) {
	e.m.Lock()
	defer e.m.Unlock()
	index := lo.IndexOf(e.entries, entry)
	// 如果存在，删除
	if index != -1 {
		e.entries = slices.Delete(e.entries, index, index+1)
	}
}

func (e *EagleeyeEngine) insertEntry(entry *EagleeyeEntry) {
	e.m.Lock()
	defer e.m.Unlock()
	index := lo.IndexOf(e.entries, entry)
	// 如果不存在，添加
	if index == -1 {
		e.entries = append(e.entries, entry)
	}
}

// RemoveFiles 删除文件(通常用于程序重启后冗余文件清理)
func (e *EagleeyeEngine) RemoveFiles(entryIDs ...string) {
	for _, entryID := range entryIDs {
		os.RemoveAll(filepath.Join(e.dir, entryID))
	}
}

// Close 关闭引擎
func (e *EagleeyeEngine) Close() {
	e.m.Lock()
	defer e.m.Unlock()
	for _, entry := range e.entries {
		if entry.state.Load() == running {
			entry.stop()
		}
	}

	global.Release()
}

const (
	initial uint32 = 0
	running uint32 = 1
	stopped uint32 = 2
)

// NewEntry 条目
type EagleeyeEntry struct {
	EntryID      string
	core         *core.Engine
	e            *EagleeyeEngine
	c            context.Context
	cancel       context.CancelFunc
	state        *atomic.Uint32
	result       *types.EntryResult
	err          error
	stageManager *stage.Manager
}

// NewEntry 创建条目
func (e *EagleeyeEngine) NewEntry(options *types.Options, extraOpts ...ExtraOption) (*EagleeyeEntry, error) {
	o := &types.Options{}

	err := o.Merge(options)
	if err != nil {
		return nil, fmt.Errorf("merge config failed: %w", err)
	}

	var extras extraOptions
	for _, extraOpt := range extraOpts {
		extraOpt.apply(&extras)
	}

	// 生成条目ID
	entryID := extras.id
	if entryID == "" {
		entryID = xid.New().String()
	}

	if err := os.MkdirAll(
		filepath.Join(
			e.dir, entryID,
		), 0755); err != nil {
		return nil, fmt.Errorf("create entry directory failed: %w", err)
	}

	stageManager := stage.NewManager()

	coreOptions := []core.Option{
		core.WithTargets(o.Targets),
		core.WithDisableBanner(true),
		core.WithStageManager(stageManager),
	}

	if len(o.ExcludeTargets) > 0 {
		coreOptions = append(coreOptions, core.WithExcludeTargets(o.ExcludeTargets))
	}

	entryResult := &types.EntryResult{
		EntryID:        entryID,
		JobResults:     make([]*types.JobResult, 0, len(o.Jobs)),
		Targets:        o.Targets,
		ExcludeTargets: o.ExcludeTargets,
	}

	if o.PortScanning.Use {
		portScanner, err := scanner.NewPortScannerV3(&scanner.PortScannerConfig{
			Ports:       o.PortScanning.Ports,
			Timeout:     o.PortScanning.Timeout,
			Count:       o.PortScanning.Count,
			Format:      o.PortScanning.Format,
			RateLimit:   o.PortScanning.RateLimit,
			Concurrency: o.PortScanning.Concurrency,
			ResultCallback: func(ctx context.Context, pr *types.PortResult) error {
				entryResult.PortScanningResult = pr
				if o.PortScanning.ResultCallback != nil {
					return o.PortScanning.ResultCallback(ctx, pr)
				}
				return nil
			},
			EntryID:      entryID,
			Silent:       true,
			Directory:    e.dir,
			StageManager: stageManager,
		})
		if err != nil {
			return nil, err
		}
		coreOptions = append(coreOptions, core.WithPortScanner(portScanner))
	}

	if o.HostDiscovery.Use {
		hostDiscovery, err := scanner.NewHostDiscoverer(&scanner.HostDiscovererConfig{
			Timeout:     o.HostDiscovery.Timeout,
			Count:       o.HostDiscovery.Count,
			Format:      o.HostDiscovery.Format,
			RateLimit:   o.HostDiscovery.RateLimit,
			Concurrency: o.HostDiscovery.Concurrency,
			ResultCallback: func(ctx context.Context, pr *types.PingResult) error {
				entryResult.HostDiscoveryResult = pr
				if o.HostDiscovery.ResultCallback != nil {
					return o.HostDiscovery.ResultCallback(ctx, pr)
				}
				return nil
			},
			EntryID:      entryID,
			Silent:       true,
			Directory:    e.dir,
			StageManager: stageManager,
		})
		if err != nil {
			return nil, err
		}
		coreOptions = append(coreOptions, core.WithHostDiscoverer(hostDiscovery))
	}

	vm, err := vuln.New(o.Mapping.Vuln)
	if err != nil {
		return nil, err
	}

	for i := range o.Jobs {
		newJob, err := job.NewJob(
			job.WithIndex(i),
			job.WithName(o.Jobs[i].Name),
			job.WithKind(o.Jobs[i].Kind),
			job.WithRateLimit(o.Jobs[i].RateLimit),
			job.WithConcurrency(o.Jobs[i].Concurrency),
			job.WithExportFormat(o.Jobs[i].Format),
			job.WithTemplate(o.Jobs[i].Template),
			job.WithGetTemplates(o.Jobs[i].GetTemplates),
			job.WithTimeout(o.Jobs[i].Timeout),
			job.WithRetries(o.Jobs[i].Count),
			job.WithCallback(func(ctx context.Context, jr *types.JobResult) error {
				entryResult.JobResults = append(entryResult.JobResults, jr)
				if o.Jobs[i].ResultCallback != nil {
					return o.Jobs[i].ResultCallback(ctx, jr)
				}
				return nil
			}),
			job.WithEntryID(entryID),
			job.WithSilent(true),
			job.WithVulnMapper(vm),
			job.WithDirectory(e.dir),
			job.WithStageManager(stageManager),
		)
		if err != nil {
			return nil, err
		}

		coreOptions = append(coreOptions, core.WithJobs(newJob))
	}

	//实例化poc引擎
	engine, err := core.New(coreOptions...)
	if err != nil {
		return nil, err
	}

	c, cancel := context.WithCancel(context.Background())
	entry := &EagleeyeEntry{
		EntryID:      entryID,
		core:         engine,
		e:            e,
		c:            c,
		cancel:       cancel,
		state:        new(atomic.Uint32),
		result:       entryResult,
		stageManager: stageManager,
	}

	return entry, nil
}

func (entry *EagleeyeEntry) Stage() types.Stage {
	return entry.stageManager.Get()
}

// Run 运行条目
func (entry *EagleeyeEntry) Run(c context.Context) error {
	// 如果状态交换失败，说明已经运行了
	// 只在状态为 0:未开始 时才可调用成功，已停止的状态为 2:已停止
	if !entry.state.CompareAndSwap(initial, running) {
		return types.ErrAlreadyRunningOrStopped
	}

	runE := make(chan error)
	go func() {
		entry.result.StartTime = time.Now()
		runE <- entry.core.ExecuteWithContext(entry.c)
		entry.result.EndTime = time.Now()
	}()

	// 当entry运行后，添加entry到引擎上下文
	entry.e.insertEntry(entry)
	// entry运行结束，走引擎上下文移除entry
	defer entry.e.deleteEntry(entry)

	var err error
	// 等待函数返回或者上下文被终止
	select {
	case <-c.Done(): // 如果外部上下文被取消
		// 停止entry运行
		entry.stop()
		// 等待函数返回
		<-runE
		err = types.ErrHasBeenStopped
	case err = <-runE: // 如果函数返回
		// 停止entry运行
		if !entry.stop() {
			// 如果执行失败，则已被停止
			err = types.ErrHasBeenStopped
		}
	case <-entry.c.Done(): // 如果entry上下文被取消
		// 等待函数返回
		<-runE
		err = types.ErrHasBeenStopped
	}

	// 如果存在错误，则删除entry对应目录
	if err != nil {
		defer entry.e.RemoveFiles(entry.EntryID)
	}

	entry.err = err
	return err
}

// Result 获取条目结果
func (entry *EagleeyeEntry) Result() *types.EntryResult {
	if entry.err != nil {
		return nil
	}
	return entry.result
}

// Stop 外部停止
func (entry *EagleeyeEntry) Stop() error {
	if !entry.stop() {
		return types.ErrStoppedOrNotRunning
	}

	return nil
}

// stop 内部停止
func (entry *EagleeyeEntry) stop() bool {
	if !entry.state.CompareAndSwap(running, stopped) {
		return false
	}

	entry.cancel()
	return true
}

func NewID() string {
	return xid.New().String()
}
