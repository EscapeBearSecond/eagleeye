package report

import (
	"bytes"
	"fmt"
	"math"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"slices"

	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/internal/util"
	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/pkg/types"
	"github.com/golang/freetype/truetype"
	"github.com/nguyenthenguyen/docx"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"github.com/wcharczuk/go-chart/v2"
)

func Generate(opts ...Option) error {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	if o.directory == "" {
		o.directory = "reports"
	}

	reportCfg := prepareConfig(&o)

	fileNames, err := generateChart(reportCfg)
	if err != nil {
		return err
	}
	defer removeFiles(fileNames...)

	return generateDocx(reportCfg)
}

func removeFiles(names ...string) {
	for _, name := range names {
		os.Remove(name)
	}
}

var chineseNum = []string{"一", "二", "三", "四", "五", "六", "七", "八", "九", "十"}

func catalogLtHunderd(num int) string {
	var nums []int
	for {
		firstNum := num / 10
		if firstNum == 0 {
			nums = append(nums, num)
			break
		}
		nums = append(nums, firstNum, 10)
		secondNum := num % 10
		if secondNum == 0 {
			break
		}
		num = secondNum
	}
	if len(nums) > 1 && nums[0] == 1 {
		nums = nums[1:]
	}

	return strings.Join(lo.Map(nums, func(num int, _ int) string { return chineseNum[num-1] }), "")
}

const (
	// placeholderCustomer = "##公司名称##"
	placeholderCustomer = ""
	// placeholderReporter = "##报告人##"
	placeholderReporter = ""
	// placeholderTargets  = "##扫描范围##"
	placeholderTargets = ""
	// placeholderExcludeTargets  = "##忽略范围##"
	placeholderExcludeTargets = ""
	// placeholderStart    = "##开始时间##"
	placeholderStart = ""
	// placeholderEnd      = "##结束时间##"
	placeholderEnd = ""
	// placeholderElaspsed = "##耗时##"
	placeholderElaspsed = ""
)

func prepareConfig(o *options) *config {
	reportCfg := config{
		ID:        o.result.EntryID,
		Directory: o.directory,
		Meta: &meta{
			Customer:        lo.If(o.customer != "", o.customer).Else(placeholderCustomer),
			Reporter:        lo.If(o.reporter != "", o.reporter).Else(placeholderReporter),
			ReportTime:      time.Now().Format("2006-01-02 15:04:05"),
			Addresses:       lo.If(len(o.result.Targets) != 0, o.result.Targets).Else([]string{placeholderTargets}),
			IgnoreAddresses: lo.If(len(o.result.ExcludeTargets) != 0, o.result.ExcludeTargets).Else([]string{placeholderExcludeTargets}),
			StartTime:       lo.If(!o.result.StartTime.IsZero(), o.result.StartTime.Format("2006-01-02 15:04:05")).Else(placeholderStart),
			EndTime:         lo.If(!o.result.EndTime.IsZero(), o.result.EndTime.Format("2006-01-02 15:04:05")).Else(placeholderEnd),
			ElaspsedTime: lo.If(!o.result.StartTime.IsZero() && !o.result.EndTime.IsZero() && o.result.EndTime.Sub(o.result.StartTime).Minutes() >= 0,
				fmt.Sprintf("%.2f", o.result.EndTime.Sub(o.result.StartTime).Minutes())).
				Else(placeholderElaspsed),
		},
		Discovery: &discovery{},
		PortScanning: &portScanning{
			Ports:   []*portScanningPorts{},
			IPPorts: []*portScanningIPPorts{},
		},
		Vulnerability: &vulnerability{
			Ports:   []*vulnerabilityPort{},
			Assets:  []*vulnerabilityAsset{},
			Details: []*vulnerabilityDetail{},
		},
	}

	// 服务发现
	if o.result.HostDiscoveryResult != nil && len(o.result.HostDiscoveryResult.Items) != 0 {
		reportCfg.Discovery.State = true

		reportCfg.Discovery.TotalCount = len(o.result.HostDiscoveryResult.Items)
		reportCfg.Discovery.Count = lo.CountBy(o.result.HostDiscoveryResult.Items, func(item *types.PingResultItem) bool {
			return item.Active
		})
		reportCfg.Discovery.CountPer = fmt.Sprintf("%.2f%%", float64(reportCfg.Discovery.Count)/float64(reportCfg.Discovery.TotalCount)*100)
		reportCfg.Discovery.UnusedCount = reportCfg.Discovery.TotalCount - reportCfg.Discovery.Count
		reportCfg.Discovery.UnusedCountPer = fmt.Sprintf("%.2f%%", float64(reportCfg.Discovery.UnusedCount)/float64(reportCfg.Discovery.TotalCount)*100)
	}

	// 端口扫描
	if o.result.PortScanningResult != nil && len(o.result.PortScanningResult.Items) != 0 {
		reportCfg.PortScanning.State = true

		reportCfg.PortScanning.TotalCount = len(lo.UniqBy(o.result.PortScanningResult.Items, func(item *types.PortResultItem) string {
			return item.IP
		}))
		portIPs := lo.ToPairs(lo.CountValuesBy(o.result.PortScanningResult.Items, func(item *types.PortResultItem) int {
			return item.Port
		}))
		slices.SortFunc(portIPs, func(a, b lo.Entry[int, int]) int {
			return b.Value - a.Value
		})
		for i, item := range portIPs {
			if i == 10 {
				break
			}
			reportCfg.PortScanning.Ports = append(reportCfg.PortScanning.Ports, &portScanningPorts{
				Port:  strconv.Itoa(item.Key),
				Count: item.Value,
			})
		}
		ipPorts := lo.ToPairs(lo.MapValues(lo.GroupBy(o.result.PortScanningResult.Items, func(item *types.PortResultItem) string {
			return item.IP
		}), func(items []*types.PortResultItem, _ string) []string {
			return lo.Map(items, func(item *types.PortResultItem, _ int) string {
				return cast.ToString(item.Port)
			})
		}))
		slices.SortFunc(ipPorts, func(a, b lo.Entry[string, []string]) int {
			return len(b.Value) - len(a.Value)
		})
		for i, item := range ipPorts {
			if i == 10 {
				break
			}
			reportCfg.PortScanning.IPPorts = append(reportCfg.PortScanning.IPPorts, &portScanningIPPorts{
				IP:    item.Key,
				Ports: item.Value,
			})
		}
	}

	if len(o.jobIndexes) != 0 && lo.EveryBy(o.jobIndexes, func(item int) bool {
		return item < len(o.result.JobResults)
	}) {
		var jobResultItems []*types.JobResultItem
		for _, idx := range o.jobIndexes {
			jobResultItems = append(jobResultItems, o.result.JobResults[idx].Items...)
		}

		if len(jobResultItems) != 0 {
			reportCfg.Vulnerability.State = true

			reportCfg.Vulnerability.TotalCount = len(lo.UniqBy(jobResultItems, func(item *types.JobResultItem) string {
				if util.IsHostPort(item.Host) {
					host, _, _ := net.SplitHostPort(item.Host)
					return host
				}
				return item.Host
			}))
			reportCfg.Vulnerability.TypeCount = len(lo.UniqBy(jobResultItems, func(item *types.JobResultItem) string {
				return item.TemplateName
			}))
			severityCountMap := lo.CountValuesBy(jobResultItems, func(item *types.JobResultItem) string { return item.Severity })
			reportCfg.Vulnerability.CriticalCount = severityCountMap["critical"]
			reportCfg.Vulnerability.HighCount = severityCountMap["high"]
			reportCfg.Vulnerability.MediumCount = severityCountMap["medium"]
			reportCfg.Vulnerability.LowCount = severityCountMap["low"]

			assetVulns := lo.ToPairs(lo.GroupBy(jobResultItems, func(item *types.JobResultItem) string {
				if util.IsHostPort(item.Host) {
					host, _, _ := net.SplitHostPort(item.Host)
					return host
				}
				return item.Host
			}))
			slices.SortFunc(assetVulns, func(a, b lo.Entry[string, []*types.JobResultItem]) int {
				return len(b.Value) - len(a.Value)
			})
			var count int
			for i, item := range assetVulns {
				if i == 10 {
					break
				}
				count += len(item.Value)
				scm := lo.CountValuesBy(item.Value, func(item *types.JobResultItem) string { return item.Severity })
				reportCfg.Vulnerability.Assets = append(reportCfg.Vulnerability.Assets, &vulnerabilityAsset{
					IP:            item.Key,
					Count:         len(item.Value),
					CriticalCount: scm["critical"],
					HighCount:     scm["high"],
					MediumCount:   scm["medium"],
					LowCount:      scm["low"],
				})
			}
			reportCfg.Vulnerability.Top10Count = count

			portVulns := lo.ToPairs(lo.CountValuesBy(jobResultItems, func(item *types.JobResultItem) string { return item.Port }))
			slices.SortFunc(portVulns, func(a, b lo.Entry[string, int]) int {
				return b.Value - a.Value
			})
			for i, item := range portVulns {
				if i == 10 {
					break
				}
				reportCfg.Vulnerability.Ports = append(reportCfg.Vulnerability.Ports, &vulnerabilityPort{
					Port:  item.Key,
					Count: item.Value,
					Percent: fmt.Sprintf("%.2f%%", float64(item.Value)/
						float64(reportCfg.Vulnerability.CriticalCount+reportCfg.Vulnerability.HighCount+reportCfg.Vulnerability.MediumCount+reportCfg.Vulnerability.LowCount)*100),
				})
			}

			vulns := lo.ToPairs(lo.GroupBy(jobResultItems, func(item *types.JobResultItem) string { return item.TemplateName }))
			slices.SortFunc(vulns, func(a, b lo.Entry[string, []*types.JobResultItem]) int {
				return len(b.Value) - len(a.Value)
			})
			for i, item := range vulns {
				if i == 10 {
					break
				}
				templateIDCount := lo.CountBy(jobResultItems, func(jri *types.JobResultItem) bool {
					return jri.TemplateID == item.Value[0].TemplateID
				})
				reportCfg.Vulnerability.Details = append(reportCfg.Vulnerability.Details, &vulnerabilityDetail{
					No:          lo.If(templateIDCount > 0, "-").Else(item.Value[0].TemplateID),
					Name:        item.Value[0].TemplateName,
					AssetCount:  len(item.Value),
					Severity:    item.Value[0].Severity,
					Description: item.Value[0].Description,
					Recommend:   item.Value[0].Remediation,
				})
			}
		}
	}

	return &reportCfg
}

func generateDocx(reportCfg *config) error {
	doc, err := docx.ReadDocxFromFS("report_template.zip", reportTemplateFS)
	if err != nil {
		return fmt.Errorf("read report template failed: %w", err)
	}
	defer doc.Close()

	e := doc.Editable()

	t, err := template.New("report").Funcs(template.FuncMap{
		"join": strings.Join,
		"add": func(a, b int) int {
			return a + b
		},
		"catalog": catalogLtHunderd,
	}).Parse(e.GetContent())
	if err != nil {
		return fmt.Errorf("parse report template failed: %w", err)
	}

	buf := &bytes.Buffer{}
	err = t.Execute(buf, &reportCfg)
	if err != nil {
		return fmt.Errorf("execute report template failed: %w", err)
	}
	e.SetContent(buf.String())

	if reportCfg.PortScanning.State {
		name := fmt.Sprintf("chart_ports_%s", reportCfg.ID)
		err = e.ReplaceImage("word/media/image1.png", fmt.Sprintf("%s.png", name))
		if err != nil {
			return fmt.Errorf("replace report image failed: %w", err)
		}

		name = fmt.Sprintf("chart_ip_ports_%s", reportCfg.ID)
		err = e.ReplaceImage("word/media/image2.png", fmt.Sprintf("%s.png", name))
		if err != nil {
			return fmt.Errorf("replace report image failed: %w", err)
		}
	}

	if reportCfg.PortScanning.State && reportCfg.Vulnerability.State {
		name := fmt.Sprintf("chart_risk_ports_%s", reportCfg.ID)
		err = e.ReplaceImage("word/media/image3.png", fmt.Sprintf("%s.png", name))
		if err != nil {
			return fmt.Errorf("replace report image failed: %w", err)
		}
	}

	if reportCfg.Vulnerability.State {
		name := fmt.Sprintf("chart_vulns_severity_%s", reportCfg.ID)
		err = e.ReplaceImage("word/media/image4.png", fmt.Sprintf("%s.png", name))
		if err != nil {
			return fmt.Errorf("replace report image failed: %w", err)
		}

		name = fmt.Sprintf("chart_vulns_detail_%s", reportCfg.ID)
		err = e.ReplaceImage("word/media/image5.png", fmt.Sprintf("%s.png", name))
		if err != nil {
			return fmt.Errorf("replace report image failed: %w", err)
		}
	}

	err = os.MkdirAll(reportCfg.Directory, 0755)
	if err != nil {
		return fmt.Errorf("create report directory failed: %w", err)
	}

	err = e.WriteToFile(fmt.Sprintf("%s/report_%s.docx", reportCfg.Directory, reportCfg.ID))
	if err != nil {
		return fmt.Errorf("write report file failed: %w", err)
	}
	return nil
}

func generateChart(reportCfg *config) ([]string, error) {
	var chartFileNames []string
	if reportCfg.PortScanning.State {
		name := fmt.Sprintf("chart_ports_%s", reportCfg.ID)
		chartFileNames = append(chartFileNames, fmt.Sprintf("%s.png", name))
		err := pieChart(&chartOptions{
			fileName: name,
			data: lo.Associate(reportCfg.PortScanning.Ports, func(p *portScanningPorts) (string, int) {
				return p.Port, p.Count
			}),
		})
		if err != nil {
			return nil, fmt.Errorf("generate report image failed: %w", err)
		}

		name = fmt.Sprintf("chart_ip_ports_%s", reportCfg.ID)
		chartFileNames = append(chartFileNames, fmt.Sprintf("%s.png", name))
		err = barChart(&chartOptions{
			fileName: name,
			data: lo.Associate(reportCfg.PortScanning.IPPorts, func(a *portScanningIPPorts) (string, int) {
				return a.IP, len(a.Ports)
			}),
			xTextRotationDegrees: 45,
		})
		if err != nil {
			return nil, fmt.Errorf("generate report image failed: %w", err)
		}
	}

	if reportCfg.PortScanning.State && reportCfg.Vulnerability.State {
		name := fmt.Sprintf("chart_risk_ports_%s", reportCfg.ID)
		chartFileNames = append(chartFileNames, fmt.Sprintf("%s.png", name))
		err := pieChart(&chartOptions{
			fileName: name,
			data: lo.Associate(reportCfg.Vulnerability.Ports, func(p *vulnerabilityPort) (string, int) {
				return p.Percent, p.Count
			}),
		})
		if err != nil {
			return nil, fmt.Errorf("generate report image failed: %w", err)
		}
	}

	if reportCfg.Vulnerability.State {
		name := fmt.Sprintf("chart_vulns_severity_%s", reportCfg.ID)
		chartFileNames = append(chartFileNames, fmt.Sprintf("%s.png", name))
		err := barChart(&chartOptions{
			fileName: name,
			data: map[string]int{
				"严重": reportCfg.Vulnerability.CriticalCount,
				"高危": reportCfg.Vulnerability.HighCount,
				"中危": reportCfg.Vulnerability.MediumCount,
				"低危": reportCfg.Vulnerability.LowCount,
			},
			width:    512,
			fontName: "hwfs.ttf",
		})
		if err != nil {
			return nil, fmt.Errorf("generate report image failed: %w", err)
		}

		name = fmt.Sprintf("chart_vulns_detail_%s", reportCfg.ID)
		chartFileNames = append(chartFileNames, fmt.Sprintf("%s.png", name))
		err = pieChart(&chartOptions{
			fileName: name,
			data: lo.Associate(reportCfg.Vulnerability.Details, func(p *vulnerabilityDetail) (string, int) {
				return p.Name, p.AssetCount
			}),
			fontName: "hwfs.ttf",
		})
		if err != nil {
			return nil, fmt.Errorf("generate report image failed: %w", err)
		}
	}
	return chartFileNames, nil
}

func pieChart(o *chartOptions) error {
	values := make([]chart.Value, 0, len(o.data))

	for k, v := range o.data {
		values = append(values, chart.Value{
			Label: k,
			Value: float64(v),
			Style: chart.Style{FontSize: 2},
		})
	}

	var font *truetype.Font
	if o.fontName != "" {
		fontBytes, err := fontFS.ReadFile(filepath.Join("font", o.fontName))
		if err != nil {
			return fmt.Errorf("read font: %w", err)
		}
		f, err := truetype.Parse(fontBytes)
		if err != nil {
			return fmt.Errorf("parse font: %w", err)
		}
		font = f
	}

	c := chart.PieChart{
		DPI:    400,
		Width:  512,
		Height: 512,
		Values: values,
		Font:   font,
	}
	f, err := os.Create(fmt.Sprintf("%s.png", o.fileName))
	if err != nil {
		return fmt.Errorf("create %s.png: %w", o.fileName, err)
	}
	defer f.Close()
	err = c.Render(chart.PNG, f)
	if err != nil {
		return fmt.Errorf("render %s.png: %w", o.fileName, err)
	}
	return nil
}

func barChart(o *chartOptions) error {
	values := make([]chart.Value, 0, len(o.data))
	for k, v := range o.data {
		values = append(values, chart.Value{
			Label: k,
			Value: float64(v),
		})
	}

	var width int
	if o.width != 0 {
		width = o.width
	}

	var font *truetype.Font
	if o.fontName != "" {
		fontBytes, err := fontFS.ReadFile(filepath.Join("font", o.fontName))
		if err != nil {
			return fmt.Errorf("read font: %w", err)
		}
		f, err := truetype.Parse(fontBytes)
		if err != nil {
			return fmt.Errorf("parse font: %w", err)
		}
		font = f
	}

	c := chart.BarChart{
		DPI:      400,
		Height:   512,
		Width:    width,
		BarWidth: 50,
		XAxis: chart.Style{
			StrokeColor:         chart.ColorBlack,
			StrokeWidth:         1,
			FontSize:            2,
			TextRotationDegrees: o.xTextRotationDegrees,
			Font:                font,
		},
		YAxis: chart.YAxis{
			Ticks: ticks(values),
			Style: chart.Style{
				StrokeColor: chart.ColorBlack,
				StrokeWidth: 1,
				FontSize:    2,
			}},
		Bars: values,
	}
	f, err := os.Create(fmt.Sprintf("%s.png", o.fileName))
	if err != nil {
		return fmt.Errorf("create %s.png: %w", o.fileName, err)
	}
	defer f.Close()
	err = c.Render(chart.PNG, f)
	if err != nil {
		return fmt.Errorf("render %s.png: %w", o.fileName, err)
	}
	return nil
}

func ticks(values []chart.Value) []chart.Tick {
	max := lo.MaxBy(values, func(a chart.Value, b chart.Value) bool {
		return a.Value > b.Value
	})
	maxValue := int(math.Ceil(max.Value))
	step := maxValue/10 + 1
	return lo.Map(lo.RangeWithSteps(0, maxValue+step, step), func(v int, _ int) chart.Tick {
		return chart.Tick{Value: float64(v), Label: cast.ToString(v)}
	})
}
