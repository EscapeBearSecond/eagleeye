package scanner

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/internal/export"
	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/internal/stage"
	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/internal/util"
	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/internal/util/log"
	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/internal/util/shaker"
	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/pkg/types"
	"github.com/panjf2000/ants/v2"
	"github.com/projectdiscovery/ratelimit"
	"github.com/samber/lo"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cast"
)

var _ Scanner[[]string, []string] = (*portScannerV3)(nil)

// portScanner 端口扫描器
type portScannerV3 struct {
	name         string
	entryID      string
	exporter     export.Exporter
	logger       *slog.Logger
	retries      int
	bar          *progressbar.ProgressBar
	callback     types.PortResultCallback
	silent       bool
	stageManager *stage.Manager

	ports      string
	portsSlice []int

	rl   *ratelimit.Limiter
	pool *ants.Pool

	portSize  int64
	total     int64
	completed *atomic.Int64
	c         context.Context
	m         sync.Mutex
	targets   map[string]struct{}
	timeout   time.Duration

	checker *shaker.Checker
}

// NewPortScanner 实例化扫描器
func NewPortScannerV3(config *PortScannerConfig) (Scanner[[]string, []string], error) {
	duration, err := time.ParseDuration(config.Timeout)
	if err != nil {
		return nil, fmt.Errorf("invalid port scanner timeout: %w", err)
	}

	scanner := &portScannerV3{
		name:         portName,
		entryID:      config.EntryID,
		retries:      config.Count,
		callback:     config.ResultCallback,
		silent:       config.Silent,
		stageManager: config.StageManager,
		timeout:      duration,
		completed:    &atomic.Int64{},
		rl:           ratelimit.New(context.Background(), uint(config.RateLimit), 1*time.Second),
	}

	if scanner.silent {
		scanner.logger = log.Must(log.NewLogger(log.WithSilent(true)))
	} else {
		scanner.logger = log.Must(log.NewLogger(log.WithStdout()))
	}

	switch config.Ports {
	case "top100":
		scanner.ports = top100
	case "top1000":
		scanner.ports = top1000
	case "http":
		scanner.ports = httpPort
	default:
		scanner.ports = config.Ports
	}

	switch config.Format {
	case "csv":
		exporter, err := export.NewCsvExporter(filepath.Join(config.Directory, scanner.entryID, scanner.name), portHeader...)
		if err != nil {
			return nil, err
		}
		scanner.exporter = exporter
	case "excel":
		exporter, err := export.NewExcelExporter(filepath.Join(config.Directory, scanner.entryID, scanner.name), portHeader...)
		if err != nil {
			return nil, err
		}
		scanner.exporter = exporter
	default:
		return nil, ErrPortOuputSupport
	}

	pool, err := ants.NewPool(config.Concurrency)
	if err != nil {
		return nil, fmt.Errorf("create port scanner routine pool failed: %w", err)
	}
	scanner.pool = pool

	scanner.checker = shaker.NewChecker()

	ports, _ := util.ParsePortsList(scanner.ports)
	scanner.portsSlice = ports
	scanner.portSize = int64(len(ports))

	return scanner, nil
}

// Scan 扫描任务
func (sc *portScannerV3) Scan(c context.Context, o *Options[[]string]) ([]string, error) {

	sc.logger.InfoContext(c, "Running port scan")

	// 开始扫描
	results, err := sc.scan(c, o)
	if err != nil {
		return nil, err
	}

	sc.logger.InfoContext(c, "Port scan complete")

	// 执行回调
	sc.doCallback(c)

	return results, nil
}

func (sc *portScannerV3) doCallback(c context.Context) error {
	if sc.callback != nil {
		results := make([]*types.PortResultItem, 0, len(sc.targets))
		for target := range sc.targets {
			ip, port, _ := net.SplitHostPort(target)
			results = append(results, &types.PortResultItem{
				EntryID:  sc.entryID,
				IP:       ip,
				Port:     cast.ToInt(port),
				HostPort: target,
			})
		}
		err := sc.callback(c, &types.PortResult{EntryID: sc.entryID, Items: results})
		if err != nil {
			return fmt.Errorf("port scanning callback failed: %w", err)
		}
	}
	return nil
}

// scan 核心scan方法
func (sc *portScannerV3) scan(c context.Context, o *Options[[]string]) ([]string, error) {
	defer sc.exporter.Close()
	defer sc.pool.Release()
	defer sc.rl.Stop()

	sc.c = c
	sc.targets = make(map[string]struct{}, 0)

	scanTargets := make([]string, 0, len(o.Targets))
	// 添加扫描目标
	for _, target := range o.Targets {
		if util.IsHostPort(target) {
			if _, contained := sc.targets[target]; !contained {
				sc.targets[target] = struct{}{}
			}
			continue
		}
		scanTargets = append(scanTargets, target)
	}

	// 构建进度条
	sc.total = sc.portSize * int64(len(scanTargets))
	sc.bar = util.NewProgressbar(sc.name, int64(sc.total), sc.silent)

	checkingLoopErr := make(chan error, 1)
	cc, stopChecker := context.WithCancel(c)
	defer stopChecker()
	go func() {
		checkingLoopErr <- sc.checker.CheckingLoop(cc)
		close(checkingLoopErr)
	}()

	ok := make(chan struct{})
	defer close(ok)
	go sc.progress(c, ok)

	select {
	case err := <-checkingLoopErr:
		return nil, fmt.Errorf("port scanner checking loop failed: %w", err)
	case <-sc.checker.WaitReady():
	}

	wg := sync.WaitGroup{}
	for _, target := range scanTargets {

		select {
		case <-c.Done():
			return nil, context.Canceled
		default:
		}

		for _, port := range sc.portsSlice {

			select {
			case <-c.Done():
				return nil, context.Canceled
			default:
			}

			addr := fmt.Sprintf("%s:%d", target, port)

			wg.Add(1)
			sc.rl.Take()
			sc.pool.Submit(func() {
				defer wg.Done()
				defer sc.completed.Add(1)

				var err error
				for range sc.retries {

					select {
					case <-c.Done():
						return
					default:
					}

					err = sc.checker.CheckAddr(addr, sc.timeout)
					if err == nil {
						break
					}
				}

				select {
				case <-c.Done():
					return
				default:
				}

				if err == nil {
					sc.m.Lock()
					_, contained := sc.targets[addr]
					if !contained {
						sc.targets[addr] = struct{}{}
					}
					sc.m.Unlock()

					if !contained {
						sc.exporter.Export(c, []any{target, port})
					}
				}
			})
		}
	}

	wg.Wait()

	select {
	case <-c.Done():
		return nil, context.Canceled
	default:
	}

	return lo.Uniq(lo.Keys(sc.targets)), nil
}

func (sc *portScannerV3) progress(c context.Context, ok <-chan struct{}) {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-c.Done():
			return
		case <-ok:
			sc.bar.Finish()
			sc.stageManager.Put(types.StagePortScanning, 1)
			return
		case <-ticker.C:
			sc.bar.Set64(sc.completed.Load())
			sc.stageManager.Put(types.StagePortScanning, sc.bar.State().CurrentPercent)
		}
	}
}
