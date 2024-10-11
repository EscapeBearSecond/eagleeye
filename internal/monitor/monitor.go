package monitor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

// Monitor 监控器
type Monitor struct {
	cpu bool
	mem bool
	// network    bool
	// etherNum   int
	startImmed bool
	interval   string
	duration   time.Duration
	draw       bool
	drawName   string
	c          context.Context
	status     *atomic.Uint32
	proc       *process.Process
	logger     *slog.Logger
	// pc         *gopacket.PacketSource
	metricsCPU []float64
	metricsMEM []float64
	stop       chan struct{}
}

// NewMonitor 实例化监控器
func NewMonitor(options ...Option) (*Monitor, error) {
	m := &Monitor{
		status:     &atomic.Uint32{},
		metricsCPU: make([]float64, 0),
		metricsMEM: make([]float64, 0),
		stop:       make(chan struct{}),
	}
	for _, opts := range options {
		opts.apply(m)
	}

	if !m.cpu && !m.mem {
		return nil, errors.New("no monitor metrics")
	}

	duration, err := time.ParseDuration(m.interval)
	if err != nil {
		return nil, fmt.Errorf("invalid monitor interval: %w", err)
	}
	m.duration = duration

	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return nil, fmt.Errorf("create process monitor failed: %w", err)
	}
	m.proc = proc

	// 目前不做网络监控
	// if m.network {
	// 	_, err := os.Stat("./interfaces.csv")
	// 	if os.IsNotExist(err) {
	// 		return nil, fmt.Errorf("please run iface command first: %w", err)
	// 	}

	// 	f, err := os.Open("./interfaces.csv")
	// 	if err != nil {
	// 		return nil, fmt.Errorf("open interfaces.csv failed: %w", err)
	// 	}

	// 	reader := csv.NewReader(f)
	// 	contents, err := reader.ReadAll()
	// 	if err != nil {
	// 		return nil, fmt.Errorf("read interfaces.csv failed: %w", err)
	// 	}
	// 	deviceName := contents[m.etherNum][1]

	// 	err = m.initNetworkTraffic(deviceName)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("initialize network traffice failed: %w", err)
	// 	}
	// }

	if m.startImmed {
		go m.Start(m.c)
	}
	return m, nil
}

// Start 开始监控
func (m *Monitor) Start(c context.Context) {
	if m.Started() {
		return
	}
	m.status.Store(1)

	// if m.network {
	// 	go m.handleNetworkTraffic(c)
	// }

	ticker := time.NewTicker(m.duration)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			args := make([]any, 0, 4)

			if m.cpu {
				cpuPercent, err := m.proc.CPUPercentWithContext(c)
				if err != nil {
					if m.logger != nil {
						m.logger.ErrorContext(c, "Monitor Collect Failed")
					}
					m.metricsCPU = append(m.metricsCPU, 0)
					continue
				}
				m.metricsCPU = append(m.metricsCPU, cpuPercent)
				args = append(args, "CPU使用率", fmt.Sprintf("%.2f%%", cpuPercent))
			}

			if m.mem {
				memInfo, err := m.proc.MemoryInfoWithContext(c)
				if err != nil {
					if m.logger != nil {
						m.logger.ErrorContext(c, "Monitor Collect Failed")
					}
					m.metricsMEM = append(m.metricsMEM, 0)
					continue
				}
				rssMB := float64(memInfo.RSS / 1024 / 1024)
				m.metricsMEM = append(m.metricsMEM, rssMB)
				args = append(args, "内存使用量", fmt.Sprintf("%.2fMB", rssMB))
			}

			if m.logger != nil {
				m.logger.InfoContext(c, "Monitor CPU/MEM Success", args...)
			}
		case <-c.Done():
			return
		case <-m.stop:
			m.status.Store(0)
			return
		}
	}
}

// Stop 关闭监控器，释放资源，并绘制指标
func (m *Monitor) Stop() {
	close(m.stop)
	// 模拟简单条件变量
	for m.status.Load() != 0 {
		// 暂停当前goroutine，使其他goroutine可以被调度
		runtime.Gosched()
	}
	m.drawMetrics(m.c)
}

// drawMetrics 绘制指标数据
func (m *Monitor) drawMetrics(c context.Context) {
	xValues := make([]float64, 0, len(m.metricsCPU))
	for i := range len(m.metricsCPU) {
		xValues = append(xValues, float64(i*int(m.duration/time.Second)))
	}

	graph := chart.Chart{
		Title: "CPU/MEM Usage",
		Background: chart.Style{
			Padding: chart.Box{
				Left:  35,
				Right: 15,
			},
			FillColor: drawing.ColorFromHex("EFEFEF"),
		},
		XAxis: chart.XAxis{
			TickPosition: chart.TickPositionUnderTick,
			ValueFormatter: func(v interface{}) string {
				return fmt.Sprintf("%.0fs", v.(float64))
			},
		},
		YAxis: chart.YAxis{
			Name: "CPU/%",
			ValueFormatter: func(v interface{}) string {
				return fmt.Sprintf("%.0f", v.(float64))
			},
		},
		YAxisSecondary: chart.YAxis{
			Name: "MEM/MB",
			ValueFormatter: func(v interface{}) string {
				return fmt.Sprintf("%.0f", v.(float64))
			},
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Name: "CPU",
				Style: chart.Style{
					StrokeWidth: 2,
					StrokeColor: drawing.ColorBlue,
					FillColor:   drawing.ColorBlue.WithAlpha(64),
				},
				YAxis:   chart.YAxisPrimary,
				XValues: xValues,
				YValues: m.metricsCPU,
			},
			chart.ContinuousSeries{
				Name: "MEM",
				Style: chart.Style{
					StrokeWidth: 2,
					StrokeColor: drawing.ColorGreen,
					FillColor:   drawing.ColorGreen.WithAlpha(64),
				},
				YAxis:   chart.YAxisSecondary,
				XValues: xValues,
				YValues: m.metricsMEM,
			},
		},
	}

	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}

	f, err := os.Create(m.drawName)
	if err != nil {
		if m.logger != nil {
			m.logger.ErrorContext(c, "Draw Monitor Metrics Failed")
		}
		return
	}
	defer f.Close()

	err = graph.Render(chart.PNG, f)
	if err != nil {
		if m.logger != nil {
			m.logger.ErrorContext(c, "Draw Monitor Metrics Failed")
		}
		return
	}
}

// Started 获取当前监控器是否启动
func (m *Monitor) Started() bool {
	return m.status.Load() == 1
}
