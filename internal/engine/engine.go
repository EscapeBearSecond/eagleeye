package engine

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/EscapeBearSecond/falcon/internal/global"
	"github.com/EscapeBearSecond/falcon/internal/job"
	"github.com/EscapeBearSecond/falcon/internal/meta"
	"github.com/EscapeBearSecond/falcon/internal/scanner"
	"github.com/EscapeBearSecond/falcon/internal/stage"
	"github.com/EscapeBearSecond/falcon/internal/target"
	"github.com/EscapeBearSecond/falcon/internal/util"
	"github.com/EscapeBearSecond/falcon/pkg/types"
	"github.com/common-nighthawk/go-figure"
	"github.com/projectdiscovery/nuclei/v3/pkg/protocols"
	"github.com/samber/lo"
)

type Engine struct {
	targets        []string
	excludeTargets []string

	portScanner    scanner.Scanner[[]string, []string] // 端口扫描器
	hostDiscoverer scanner.Scanner[[]string, []string] // 探活扫描器

	jobs []*job.Job

	disableBanner bool
	eOptions      *protocols.ExecutorOptions

	stageManager *stage.Manager
}

// New 实例化引擎
func New(opts ...Option) (*Engine, error) {
	e := &Engine{
		targets:        make([]string, 0),
		excludeTargets: make([]string, 0),
		jobs:           make([]*job.Job, 0, 3),
	}

	for _, o := range opts {
		o(e)
	}

	err := e.init()
	if err != nil {
		return nil, fmt.Errorf("init engine failed: %w", err)
	}

	return e, nil
}

func (e *Engine) init() error {
	if len(e.targets) == 0 {
		return types.ErrInvalidTargets
	}

	targets, err := target.ProcessAsync(e.targets, e.excludeTargets...)
	if err != nil {
		return fmt.Errorf("process targets failed: %w", err)
	}
	e.targets = targets

	e.eOptions = global.ExecutorOptions()

	// 丢弃错误
	// 如果browser为nil, 会跳过所有headless模板
	browser, _ := global.Browser()
	e.eOptions.Browser = browser

	names := make([]string, 0, len(e.jobs))
	// 遍历所有的job
	for _, j := range e.jobs {
		// 如果job的name不存在，则加入name
		if !lo.Contains(names, j.Name()) {
			names = append(names, j.Name())
		} else { //如果job的name已存在，则存在重名任务，给当前job添加随机后缀
			j.SetName(fmt.Sprintf("%s_%s", j.Name(), util.RandomStr(10)))
		}
		// 加载模版，使用filepath的walkdir方式
		err := j.LoadTemplates(e.eOptions)
		if err != nil {
			return fmt.Errorf("load job [%s] templates failed: %w", j.Name(), err)
		}
	}

	e.stageManager.Put(types.StagePreExecute, 0)

	return nil
}

// printBanner 打印banner信息
func printBanner() {
	figure.NewColorFigure("Eagleeye", "rectangles", "green", true).Print()
	fmt.Printf("\n\t\t\t\tv%s\n\n", meta.BuildVer)
}

// ExecuteWithContext 执行
func (e *Engine) ExecuteWithContext(c context.Context) error {
	defer e.stageManager.Put(types.StagePostExecute, 0)
	defer e.close()

	if !e.disableBanner {
		printBanner()
	}

	// 用于任务间隔的计时（当前置任务结束，重置计时器）
	timer := time.NewTimer(0)
	defer timer.Stop()

	if e.hostDiscoverer != nil {
		<-timer.C
		targets, err := e.hostDiscoverer.Scan(c, &scanner.Options[[]string]{Targets: e.targets})
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return fmt.Errorf("run host discovery failed: %w", err)
		}

		if len(targets) == 0 {
			return types.ErrNoActiveHost
		}

		e.targets = targets
		timer.Reset(5 * time.Second)

		debug.FreeOSMemory()
	}

	// 执行端口扫描
	if e.portScanner != nil {
		<-timer.C
		targets, err := e.portScanner.Scan(c, &scanner.Options[[]string]{Targets: e.targets})
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return fmt.Errorf("run port scanning failed: %w", err)
		}

		if len(targets) == 0 {
			return types.ErrNoExistPort
		}

		e.targets = targets
		timer.Reset(5 * time.Second)

		debug.FreeOSMemory()
	}

	for _, j := range e.jobs {
		select {
		case <-c.Done():
			return nil
		default:
		}

		<-timer.C

		err := j.ExecuteWithContext(c, &job.Options{Targets: e.targets})
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return fmt.Errorf("excute job [%s] failed: %w", j.Name(), err)
		}
		timer.Reset(5 * time.Second)

		debug.FreeOSMemory()
	}

	return nil
}

func (e *Engine) close() {
	e.eOptions.Output.Close()
	e.eOptions.Progress.Stop()
	e.eOptions.RateLimiter.Stop()

	if e.eOptions.Browser != nil {
		e.eOptions.Browser.Close()
	}
}
