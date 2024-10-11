package global

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/projectdiscovery/goflags"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/gologger/levels"
	"github.com/projectdiscovery/nuclei/v3/pkg/protocols"
	"github.com/projectdiscovery/nuclei/v3/pkg/protocols/common/protocolinit"
	"github.com/projectdiscovery/nuclei/v3/pkg/protocols/common/protocolstate"
	"github.com/projectdiscovery/nuclei/v3/pkg/protocols/headless/engine"
	"github.com/projectdiscovery/nuclei/v3/pkg/templates"
	"github.com/projectdiscovery/nuclei/v3/pkg/types"
	"github.com/projectdiscovery/ratelimit"
	"github.com/projectdiscovery/useragent"
)

var (
	useSyncPool bool

	timeout  int
	tOptions *types.Options //nuclei *types.Options
)

// Init 初始化nuclei相关对象
func Init() error {
	// golang 运行时环境配置
	{
		// 默认开启tls rsa加密
		os.Setenv("GODEBUG", "tlsrsakex=1")
		// 配置go最大processor数为核心数2倍
		runtime.GOMAXPROCS(runtime.NumCPU() * 2)

		// GC调整为50%
		debug.SetGCPercent(50)
		// 最大线程数设置为20000
		debug.SetMaxThreads(20000)
	}

	log.SetOutput(io.Discard)

	gologger.DefaultLogger.SetMaxLevel(levels.LevelFatal)
	gologger.DefaultLogger.SetWriter(&nopWriter{})

	useSyncPool = false

	// 最大超时10s
	timeout = 10

	tOptions = types.DefaultOptions()
	// 不限制频率，业务侧控制
	tOptions.RateLimit = math.MaxInt32
	// 默认全局超时，业务侧控制
	tOptions.Timeout = timeout
	// 不重试，业务侧控制
	tOptions.Retries = 1
	// 使用系统dns
	tOptions.SystemResolvers = true

	// 配置javascript runtime池大小（根据文档，似乎120最合适，就先这样吧）
	tOptions.JsConcurrency = 120

	// 针对headless的页面超时
	tOptions.PageTimeout = 3 * timeout
	// 设置为true，默认编译headless模板
	tOptions.Headless = true
	// 使用已安装的chrome(避免下载)
	tOptions.UseInstalledChrome = true
	// experimental(不确定是否有效)
	tOptions.HeadlessOptionalArguments = goflags.StringSlice{
		"rod-leakless=true",
	}

	// 设置为true，默认编译code模板
	tOptions.EnableCodeTemplates = true

	// 允许本地文件访问（目前主要用于文件字典）
	tOptions.AllowLocalFileAccess = true

	// 初始化nuclei全局配置
	err := protocolinit.Init(tOptions)
	if err != nil {
		return fmt.Errorf("global init failed: %w", err)
	}

	tOptions.GetTimeouts()

	return nil
}

// TypesOptions 获取*types.Options
func TypesOptions() *types.Options {
	return tOptions
}

// ExecutorOptions 获取*protocols.ExecutorOptions（非单例）
func ExecutorOptions() *protocols.ExecutorOptions {

	output := &fakeWriter{}
	progress := &fakeProgress{}

	// interactshClient, err := interactsh.New(interactsh.DefaultOptions(output, nil, progress))
	// if err != nil {
	// 	return fmt.Errorf("create interactsh client failed: %w", err)
	// }

	eOptions := &protocols.ExecutorOptions{
		Parser:  templates.NewParser(),
		Options: tOptions,
		// 不限制频率，保持与types.Options一致
		RateLimiter: ratelimit.New(context.Background(), math.MaxInt32, 1*time.Second),
		Progress:    progress,
		Output:      output,
		DoNotCache:  true,
		// Browser:     browser,
		// Interactsh:  interactshClient,
	}

	return eOptions
}

func Browser() (*engine.Browser, error) {
	// 初始化浏览器
	browser, err := engine.New(tOptions)

	if err != nil {
		return nil, fmt.Errorf("create headless browser failed: %w", err)
	}

	//设置浏览器随机ua
	browser.SetUserAgent(useragent.PickRandom().Raw)
	return browser, nil
}

// GlobalTimeout 获取*types.Options的timeout
func MaxTimeout() time.Duration {
	return time.Duration(timeout) * time.Second
}

func UseSyncPool() bool {
	return useSyncPool
}

// Release 释放对象
func Release() {
	protocolstate.Close()
}
