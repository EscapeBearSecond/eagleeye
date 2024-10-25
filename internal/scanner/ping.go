package scanner

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/EscapeBearSecond/falcon/internal/export"
	"github.com/EscapeBearSecond/falcon/internal/stage"
	"github.com/EscapeBearSecond/falcon/internal/util"
	"github.com/EscapeBearSecond/falcon/internal/util/log"
	"github.com/EscapeBearSecond/falcon/internal/util/privileges"
	"github.com/EscapeBearSecond/falcon/pkg/types"
	"github.com/panjf2000/ants/v2"
	"github.com/projectdiscovery/ratelimit"
	probing "github.com/prometheus-community/pro-bing"
	"github.com/samber/lo"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cast"
)

var _ Scanner[[]string, []string] = (*hostDiscoverer)(nil)

// hostDiscoverer 在线检测
type hostDiscoverer struct {
	name         string
	entryID      string
	timeout      time.Duration
	count        int
	targets      map[string][]string
	exporter     export.Exporter
	logger       *slog.Logger
	ratelimit    int
	rl           *ratelimit.Limiter
	concurrency  int
	pool         *ants.Pool
	m            sync.Mutex
	bar          *progressbar.ProgressBar
	callback     types.PingResultCallback
	silent       bool
	stageManager *stage.Manager
	completed    *atomic.Int64
}

// NewHostDiscoverer 实例化在线检测
func NewHostDiscoverer(cfg *HostDiscovererConfig) (Scanner[[]string, []string], error) {
	duration, err := time.ParseDuration(cfg.Timeout)
	if err != nil {
		return nil, fmt.Errorf("invalid host discovery timeout: %w", err)
	}

	pinger := &hostDiscoverer{
		count:        cfg.Count,
		timeout:      duration,
		concurrency:  cfg.Concurrency,
		ratelimit:    cfg.RateLimit,
		callback:     cfg.ResultCallback,
		silent:       cfg.Silent,
		entryID:      cfg.EntryID,
		stageManager: cfg.StageManager,
		completed:    &atomic.Int64{},
	}

	if pinger.silent {
		pinger.logger = log.Must(log.NewLogger(log.WithSilent(true)))
	} else {
		pinger.logger = log.Must(log.NewLogger(log.WithStdout()))
	}

	pinger.name = pingName

	switch cfg.Format {
	case "csv":
		exporter, err := export.NewCsvExporter(filepath.Join(cfg.Directory, cfg.EntryID, pinger.name), pingHeader...)
		if err != nil {
			return nil, err
		}
		pinger.exporter = exporter
	case "excel":
		exporter, err := export.NewExcelExporter(filepath.Join(cfg.Directory, cfg.EntryID, pinger.name), pingHeader...)
		if err != nil {
			return nil, err
		}
		pinger.exporter = exporter
	default:
		return nil, ErrHostOuputSupport
	}

	pool, err := ants.NewPool(pinger.concurrency)
	if err != nil {
		return nil, fmt.Errorf("create host discoverer routine pool failed: %w", err)
	}
	pinger.pool = pool

	rl := ratelimit.New(context.Background(), uint(pinger.ratelimit), 1*time.Second)
	pinger.rl = rl

	return pinger, nil
}

func (p *hostDiscoverer) Scan(c context.Context, o *Options[[]string]) ([]string, error) {
	p.logger.InfoContext(c, "Running host discovery")
	results, err := p.scan(c, o)
	if err != nil {
		return nil, err
	}
	p.logger.InfoContext(c, "Host discovery completed")

	p.doCallback(c)

	return results, nil
}

func (p *hostDiscoverer) doCallback(c context.Context) error {
	if p.callback != nil {
		results := make([]*types.PingResultItem, 0, len(p.targets))
		for ip, infos := range p.targets {
			result := &types.PingResultItem{
				EntryID: p.entryID,
				IP:      ip,
			}
			if ok := infos != nil; ok {
				result.Active = ok
				os, _ := lo.Find(infos, func(info string) bool {
					return strings.Split(info, ":")[0] == "os"
				})
				ttl, _ := lo.Find(infos, func(info string) bool {
					return strings.Split(info, ":")[0] == "ttl"
				})
				result.OS = strings.Split(os, ":")[1]
				result.TTL = cast.ToInt(strings.Split(ttl, ":")[1])
			}
			results = append(results, result)
		}
		err := p.callback(c, &types.PingResult{EntryID: p.entryID, Items: results})
		if err != nil {
			return fmt.Errorf("host discovery callback failed: %w", err)
		}
	}
	return nil
}

// Scan 扫描方法
func (p *hostDiscoverer) scan(c context.Context, o *Options[[]string]) ([]string, error) {
	defer p.exporter.Close()
	defer p.pool.Release()
	defer p.rl.Stop()

	p.targets = make(map[string][]string, len(o.Targets))

	total := int64(len(o.Targets))
	p.bar = util.NewProgressbar(pingName, total, p.silent)

	ok := make(chan struct{})
	defer close(ok)
	go p.progress(c, ok)

	wg := sync.WaitGroup{}
	for _, item := range o.Targets {
		t := item
		if util.IsHostPort(item) {
			t, _, _ = net.SplitHostPort(item)
		}

		select {
		case <-c.Done():
			return nil, context.Canceled
		default:
		}

		pinger, err := probing.NewPinger(t)
		if err != nil {
			return nil, err
		}

		var (
			addr  string
			infos []string
		)

		if runtime.GOOS == "windows" {
			pinger.SetPrivileged(true)
		} else {
			pinger.SetPrivileged(privileges.IsPrivileged)
		}
		pinger.Count = p.count
		pinger.Interval = 100 * time.Millisecond
		pinger.Timeout = p.timeout
		pinger.OnRecv = func(packet *probing.Packet) {
			addr = packet.Addr
			switch packet.TTL {
			case 128:
				infos = append(infos, "ttl:128")
				infos = append(infos, "os:windows")
			case 64:
				infos = append(infos, "ttl:64")
				infos = append(infos, "os:linux")
			default:
				infos = append(infos, fmt.Sprintf("ttl:%d", packet.TTL))
				infos = append(infos, "os:unknown")
			}
			pinger.Stop()
		}

		wg.Add(1)
		p.rl.Take()
		p.pool.Submit(func() {
			defer wg.Done()
			defer p.completed.Add(1)

			defer pinger.Stop()

			pinger.RunWithContext(c)

			select {
			case <-c.Done():
				return
			default:
			}

			if addr != "" {
				p.m.Lock()
				_, contained := p.targets[addr]
				if !contained {
					p.targets[addr] = infos
				}
				p.m.Unlock()

				os, _ := lo.Find(infos, func(info string) bool {
					return strings.Split(info, ":")[0] == "os"
				})
				ttl, _ := lo.Find(infos, func(info string) bool {
					return strings.Split(info, ":")[0] == "ttl"
				})

				if !contained {
					p.exporter.Export(c, []any{addr, "是", strings.Split(os, ":")[1], strings.Split(ttl, ":")[1]})
				}
			} else {
				p.m.Lock()
				_, contained := p.targets[t]
				if !contained {
					p.targets[t] = nil
				}
				p.m.Unlock()

				if !contained {
					p.exporter.Export(c, []any{t, "否", "", ""})
				}
			}
		})
	}

	wg.Wait()

	select {
	case <-c.Done():
		return nil, context.Canceled
	default:
	}

	return lo.Uniq(
		lo.Keys(
			lo.PickBy(p.targets,
				func(_ string, v []string) bool { return v != nil },
			),
		),
	), nil
}

func (p *hostDiscoverer) progress(c context.Context, ok <-chan struct{}) {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-c.Done():
			return
		case <-ok:
			p.bar.Finish()
			p.stageManager.Put(types.StageHostDiscovery, 1)
			return
		case <-ticker.C:
			p.bar.Set64(p.completed.Load())
			p.stageManager.Put(types.StageHostDiscovery, p.bar.State().CurrentPercent)
		}
	}
}
