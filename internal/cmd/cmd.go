package cmd

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/EscapeBearSecond/falcon/internal/engine"
	"github.com/EscapeBearSecond/falcon/internal/flag"
	"github.com/EscapeBearSecond/falcon/internal/global"
	"github.com/EscapeBearSecond/falcon/internal/job"
	"github.com/EscapeBearSecond/falcon/internal/mapper/vuln"
	"github.com/EscapeBearSecond/falcon/internal/meta"
	"github.com/EscapeBearSecond/falcon/internal/monitor"
	"github.com/EscapeBearSecond/falcon/internal/scanner"
	"github.com/EscapeBearSecond/falcon/internal/util/log"
	"github.com/EscapeBearSecond/falcon/pkg/types"
	"github.com/spf13/cobra"
)

var (
	cfgFile string //配置文件

	o types.Options //配置对象
)

// rootCmd 主命令
var rootCmd = cobra.Command{
	Use:           "eagleeye",
	Short:         "scanner",
	Long:          "eagleeye scanner",
	Version:       meta.BuildVer,
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if cfgFile != "" {
			err := o.Parse(cfgFile)
			if err != nil {
				return err
			}
		}

		if len(o.Targets) == 0 || o.Targets[0] == "" {
			return errors.New("targets are required")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmd.Context()

		if err := global.Init(); err != nil {
			return err
		}
		defer global.Release()

		options := []engine.Option{
			engine.WithTargets(o.Targets),
		}

		if len(o.ExcludeTargets) > 0 {
			options = append(options, engine.WithExcludeTargets(o.ExcludeTargets))
		}

		if o.PortScanning.Use {
			portScanner, err := scanner.NewPortScannerV3(&scanner.PortScannerConfig{
				Ports:       o.PortScanning.Ports,
				Timeout:     o.PortScanning.Timeout,
				Count:       o.PortScanning.Count,
				Format:      o.PortScanning.Format,
				RateLimit:   o.PortScanning.RateLimit,
				Concurrency: o.PortScanning.Concurrency,
				Directory:   ".",
			})
			if err != nil {
				return err
			}
			options = append(options, engine.WithPortScanner(portScanner))
		}

		if o.HostDiscovery.Use {
			hostDiscoverer, err := scanner.NewHostDiscoverer(&scanner.HostDiscovererConfig{
				Timeout:     o.HostDiscovery.Timeout,
				Count:       o.HostDiscovery.Count,
				Format:      o.HostDiscovery.Format,
				RateLimit:   o.HostDiscovery.RateLimit,
				Concurrency: o.HostDiscovery.Concurrency,
				Directory:   ".",
			})
			if err != nil {
				return err
			}
			options = append(options, engine.WithHostDiscoverer(hostDiscoverer))
		}

		var logger *slog.Logger
		if o.OutLog {
			var err error
			logger, err = log.NewLogger(log.WithFile("log.out"))
			if err != nil {
				return err
			}
		}

		vm, err := vuln.New(o.Mapping.Vuln)
		if err != nil {
			return err
		}
		for i, j := range o.Jobs {
			newJob, err := job.NewJob(
				job.WithIndex(i),
				job.WithName(j.Name),
				job.WithRateLimit(j.RateLimit),
				job.WithConcurrency(j.Concurrency),
				job.WithExportFormat(j.Format),
				job.WithOutLogger(logger),
				job.WithTemplate(j.Template),
				job.WithTimeout(j.Timeout),
				job.WithRetries(j.Count),
				job.WithEnableHeadless(j.Headless),
				job.WithVulnMapper(vm),
				job.WithDirectory("."),
			)
			if err != nil {
				return err
			}

			options = append(options, engine.WithJobs(newJob))
		}

		//实例化poc引擎
		e, err := engine.New(options...)
		if err != nil {
			return err
		}

		//实例化监控器，并立刻执行
		if o.Monitor.Use {
			logger, err := log.NewLogger(log.WithFile("monitor.out"))
			if err != nil {
				return err
			}
			monitorOpts := []monitor.Option{
				monitor.WithCPU(),
				monitor.WithMemory(),
				monitor.WithImmediately(c),
				monitor.WithInterval(o.Monitor.Interval),
				monitor.WithLogger(logger),
				monitor.WithDraw("monitor.png"),
			}

			mo, err := monitor.NewMonitor(monitorOpts...)
			if err != nil {
				return err
			}
			defer mo.Stop()
		}

		//执行引擎
		return e.ExecuteWithContext(c)
	},
}

func init() {
	defaultOptions := types.DefaultOptions()

	rootCmd.Flags().StringVar(&cfgFile, "cfg", "", "config file")

	{
		rootCmd.Flags().StringSliceVarP(&o.Targets, "targets", "u", nil, "目标地址/文件")
		rootCmd.Flags().StringSliceVar(&o.ExcludeTargets, "ue", nil, "排除目标地址/文件")
	}

	rootCmd.Flags().BoolVarP(&o.OutLog, "out_log", "l", false, "任务执行日志")

	{
		rootCmd.Flags().StringVarP(&o.Mapping.Vuln, "vuln", "z", "", "漏洞映射文件")
	}

	//进程网卡监控
	{
		rootCmd.Flags().BoolVarP(&o.Monitor.Use, "monitor", "m", false, "监控日志")
		rootCmd.Flags().StringVar(&o.Monitor.Interval, "mi", defaultOptions.Monitor.Interval, "监控频率")
		// rootCmd.Flags().IntVar(&cfg.Monitor.EtherNum, "me", 0, "网卡编号")
	}

	//设备在线监测
	{
		rootCmd.Flags().BoolVarP(&o.HostDiscovery.Use, "discovery", "d", false, "设备探活")
		rootCmd.Flags().StringVar(&o.HostDiscovery.Timeout, "de", defaultOptions.HostDiscovery.Timeout, "探活超时时间")
		rootCmd.Flags().IntVar(&o.HostDiscovery.Count, "dn", defaultOptions.HostDiscovery.Count, "探活轮次")
		rootCmd.Flags().StringVar(&o.HostDiscovery.Format, "da", defaultOptions.HostDiscovery.Format, "探活输出格式")
		rootCmd.Flags().IntVar(&o.HostDiscovery.RateLimit, "dr", defaultOptions.HostDiscovery.RateLimit, "探活频率")
		rootCmd.Flags().IntVar(&o.HostDiscovery.Concurrency, "dc", defaultOptions.HostDiscovery.Concurrency, "探活并发数")
	}

	//端口扫描
	{
		rootCmd.Flags().BoolVarP(&o.PortScanning.Use, "port_scanning", "p", false, "端口扫描")
		rootCmd.Flags().StringVar(&o.PortScanning.Timeout, "pe", defaultOptions.PortScanning.Timeout, "端口扫描超时时间")
		rootCmd.Flags().IntVar(&o.PortScanning.Count, "pn", defaultOptions.PortScanning.Count, "端口扫描轮次")
		rootCmd.Flags().StringVar(&o.PortScanning.Format, "pa", defaultOptions.PortScanning.Format, "端口扫描输出格式")
		rootCmd.Flags().StringVar(&o.PortScanning.Ports, "pp", defaultOptions.PortScanning.Ports, "端口扫描端口")
		rootCmd.Flags().IntVar(&o.PortScanning.RateLimit, "pr", defaultOptions.PortScanning.RateLimit, "端口扫描频率")
		rootCmd.Flags().IntVar(&o.PortScanning.Concurrency, "pc", defaultOptions.PortScanning.Concurrency, "端口扫描并发数")
	}

	rootCmd.Flags().VarP(flag.NewJobFlag(&o.Jobs), "job", "j", "任务配置")

	rootCmd.AddCommand(&metaCmd, &ifaceCmd, &mockServerCmd, &apiserverCmd, &mmh3Cmd, &licenseCmd, &templateCmd, &reportCmd)
}

func Execute() {
	c, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-quit:
			cancel()
		case <-c.Done():
		}
	}()

	cobra.CheckErr(rootCmd.ExecuteContext(c))
}
