package report

import "embed"

//go:embed report_template.zip
var reportTemplateFS embed.FS

//go:embed font
var fontFS embed.FS

type config struct {
	ID            string
	Directory     string
	Meta          *meta          //元数据
	Discovery     *discovery     //探活数据
	PortScanning  *portScanning  //端口扫描数据
	Vulnerability *vulnerability //漏洞数据
}

type meta struct {
	Customer        string   //客户名
	Reporter        string   //报告人
	ReportTime      string   //报告时间
	Addresses       []string //探活地址
	IgnoreAddresses []string //忽略地址
	StartTime       string   //开始时间
	EndTime         string   //结束时间
	ElaspsedTime    string   //耗时
}

type discovery struct {
	TotalCount     int    //探活资产总数
	Count          int    //存活数
	CountPer       string //存活率
	UnusedCount    int    //未使用数
	UnusedCountPer string //未使用率
	State          bool
}

type portScanning struct {
	TotalCount int                    //开放端口资产总数
	Ports      []*portScanningPorts   //端口及资产数
	IPPorts    []*portScanningIPPorts //资产IP及端口号
	State      bool
}

type portScanningPorts struct {
	Port  string //端口
	Count int    //资产数
}

type portScanningIPPorts struct {
	IP    string   //IP
	Ports []string //端口
}

type vulnerability struct {
	TotalCount    int                    //漏洞资产数
	TypeCount     int                    //漏洞类型数
	CriticalCount int                    //严重漏洞数
	HighCount     int                    //高危漏洞数
	MediumCount   int                    //中危漏洞数
	LowCount      int                    //低危漏洞数
	Top10Count    int                    //Top10资产总漏洞数
	Ports         []*vulnerabilityPort   //端口及漏洞数
	Assets        []*vulnerabilityAsset  //资产IP及漏洞数
	Details       []*vulnerabilityDetail //漏洞详情
	State         bool
}

type vulnerabilityPort struct {
	Port    string //端口
	Count   int    //漏洞数
	Percent string //漏洞占比
}

type vulnerabilityAsset struct {
	IP            string //IP
	Count         int    //漏洞数
	CriticalCount int    //严重漏洞数
	HighCount     int    //高危漏洞数
	MediumCount   int    //中危漏洞数
	LowCount      int    //低危漏洞数
}

type vulnerabilityDetail struct {
	No          string //漏洞编号
	Name        string //漏洞名称
	AssetCount  int    //资产数
	Severity    string //漏洞等级
	Description string //漏洞描述
	Recommend   string //漏洞建议
}
