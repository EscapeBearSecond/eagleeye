package types

import (
	"fmt"
	"os"

	"github.com/jinzhu/copier"
	"gopkg.in/yaml.v3"
)

// DefaultOptions 默认选项
func DefaultOptions(jobSize ...int) *Options {
	o := &Options{
		Monitor: MonitorOptions{
			Interval: "5s",
		},
		PortScanning: PortScanningOptions{
			Timeout:     "1s",
			Count:       1,
			Format:      "csv",
			Ports:       "http",
			RateLimit:   150,
			Concurrency: 150,
		},
		HostDiscovery: HostDiscoveryOptions{
			Timeout:     "1s",
			Count:       1,
			Format:      "csv",
			RateLimit:   150,
			Concurrency: 150,
		},
	}
	if len(jobSize) != 0 && jobSize[0] != 0 {
		for range jobSize[0] {
			o.Jobs = append(o.Jobs, JobOptions{
				Format:      "csv",
				Count:       1,
				Timeout:     "1s",
				RateLimit:   150,
				Concurrency: 150,
			})
		}
	}
	return o
}

// Options 选项
type Options struct {
	Targets        []string             `yaml:"targets" json:"targets"`                 //目标
	ExcludeTargets []string             `yaml:"exclude_targets" json:"exclude_targets"` //排除目标
	OutLog         bool                 `yaml:"out_log" json:"-"`                       //输出运行日志
	Monitor        MonitorOptions       `yaml:"monitor" json:"-"`                       //监控
	Mapping        Mapping              `yaml:"mapping" json:"mapping"`                 //映射
	PortScanning   PortScanningOptions  `yaml:"port_scanning" json:"port_scanning"`     //端口扫描
	HostDiscovery  HostDiscoveryOptions `yaml:"host_discovery" json:"host_discovery"`   //在线检测
	Jobs           []JobOptions         `yaml:"jobs" json:"jobs"`                       //任务
}

type Mapping struct {
	Vuln string `yaml:"vuln" json:"vuln"`
}

// JobOptions 任务选项
type JobOptions struct {
	Name           string            `yaml:"name" json:"name"`               //任务名称
	Kind           string            `yaml:"kind" json:"kind"`               //任务类型(由于name可能是中文，kind只作为英文比较字段)
	Template       string            `yaml:"template" json:"template"`       //模板，支持单文件和目录
	GetTemplates   GetTemplates      `yaml:"-" json:"-"`                     //获取模板
	Format         string            `yaml:"format" json:"format"`           //输出格式
	Count          int               `yaml:"count" json:"count"`             //循环次数
	Timeout        string            `yaml:"timeout" json:"timeout"`         //超时时间
	RateLimit      int               `yaml:"rate_limit" json:"rate_limit"`   //限流
	Concurrency    int               `yaml:"concurrency" json:"concurrency"` //并发数
	Headless       bool              `yaml:"headless" json:"headless"`       //是否启用浏览器
	ResultCallback JobResultCallback `yaml:"-" json:"-"`                     //结果回调
}

// MonitorOptions 监控选项(sdk模式不生效)
type MonitorOptions struct {
	Use      bool   `yaml:"use" json:"-"`      //开启指标监控
	Interval string `yaml:"interval" json:"-"` //监控周期(s)
	// EtherNum int    `yaml:"ether_num"` //监控网卡编号
}

// PortScanningOptions 端口扫描选项
type PortScanningOptions struct {
	Use            bool               `yaml:"use" json:"use"`                 //开启端口扫描
	Timeout        string             `yaml:"timeout" json:"timeout"`         //超时时间(0.5s, 1m)
	Count          int                `yaml:"count" json:"count"`             //轮次
	Format         string             `yaml:"format" json:"format"`           //导出结果格式(csv,excel)
	Ports          string             `yaml:"ports" json:"ports"`             //扫描端口
	RateLimit      int                `yaml:"rate_limit" json:"rate_limit"`   //限流
	Concurrency    int                `yaml:"concurrency" json:"concurrency"` //并发数
	ResultCallback PortResultCallback `yaml:"-" json:"-"`                     //结果回调
}

// HostDiscoveryOptions 在线检测选项
type HostDiscoveryOptions struct {
	Use            bool               `yaml:"use" json:"use"`                 //开启设备发现(探活)
	Timeout        string             `yaml:"timeout" json:"timeout"`         //超时时间(0.5s, 1m)
	Count          int                `yaml:"count" json:"count"`             //轮次
	Format         string             `yaml:"format" json:"format"`           //导出结果格式(csv,excel)
	RateLimit      int                `yaml:"rate_limit" json:"rate_limit"`   //限流
	Concurrency    int                `yaml:"concurrency" json:"concurrency"` //并发数
	ResultCallback PingResultCallback `yaml:"-" json:"-"`                     //结果回调
}

// Parse 解析配置
func (o *Options) Parse(cfgFile string) error {
	f, err := os.Open(cfgFile)
	if err != nil {
		return fmt.Errorf("open config file failed: %w", err)
	}

	decoder := yaml.NewDecoder(f)
	var yamlCfg Options
	err = decoder.Decode(&yamlCfg)
	if err != nil {
		return fmt.Errorf("invalid config file format: %w", err)
	}

	return o.Merge(&yamlCfg)
}

// Merge 合并配置
func (o *Options) Merge(options *Options) error {
	defaultOptions := DefaultOptions(len(options.Jobs))
	// 配置默认值
	copier.CopyWithOption(o, defaultOptions, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	})

	// 合并yaml配置文件中的值
	copier.CopyWithOption(o, options, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	})

	return nil
}
