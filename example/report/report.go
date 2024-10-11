package main

import (
	"context"
	"log"
	"os"

	"github.com/EscapeBearSecond/eagleeye/pkg/report"
	eagleeye "github.com/EscapeBearSecond/eagleeye/pkg/sdk"
	"github.com/EscapeBearSecond/eagleeye/pkg/types"
)

func main() {
	stdlog := log.New(os.Stderr, "", log.LstdFlags)

	options := &types.Options{
		Targets: []string{
			"192.168.1.1-192.168.1.255",
		},
		ExcludeTargets: []string{
			"192.168.1.108",
		},
		HostDiscovery: types.HostDiscoveryOptions{
			Use:         true,
			Timeout:     "5s",
			Count:       1,
			Format:      "csv",
			RateLimit:   1000,
			Concurrency: 1000,
		},
		PortScanning: types.PortScanningOptions{
			Use:         true,
			Timeout:     "5s",
			Count:       1,
			Format:      "csv",
			Ports:       "http",
			RateLimit:   1000,
			Concurrency: 1000,
		},
		Jobs: []types.JobOptions{
			{
				Name:        "漏洞扫描",
				Kind:        "vul-scan",
				Template:    "./templates/漏洞扫描",
				Format:      "csv",
				Count:       1,
				Timeout:     "5s",
				RateLimit:   2000,
				Concurrency: 2000,
			},
			{
				Name:        "资产扫描",
				Kind:        "asset-scan",
				Template:    "./templates/资产扫描",
				Format:      "csv",
				Count:       1,
				Timeout:     "5s",
				RateLimit:   2000,
				Concurrency: 2000,
			},
		},
	}

	engine, err := eagleeye.NewEngine(eagleeye.WithDirectory("./results"))
	if err != nil {
		stdlog.Fatalln(err)
	}
	defer engine.Close()

	entry, err := engine.NewEntry(options)
	if err != nil {
		stdlog.Fatalln("error:", err)
	}

	err = entry.Run(context.Background())
	if err != nil {
		stdlog.Fatalln("error:", err)
	}

	ret := entry.Result()

	err = report.Generate(
		report.WithJobIndexes(0),
		report.WithEntryResult(ret),
		report.WithCustomer("示例客户"),
		report.WithReporter("示例报告人"))
	if err != nil {
		stdlog.Fatalln("error:", err)
	}
}
