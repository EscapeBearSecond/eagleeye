package types

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/projectdiscovery/nuclei/v3/pkg/output"
	"github.com/samber/lo"
)

type JobResult struct {
	EntryID string           `json:"-"`
	Name    string           `json:"name"`
	Kind    string           `json:"kind"`
	Items   []*JobResultItem `json:"items"`
}

type JobResultItem struct {
	EntryID          string         `json:"-"`
	TemplateID       string         `json:"template_id"`
	TemplateName     string         `json:"template_name"`
	Type             string         `json:"type"`
	Severity         string         `json:"severity"`
	Host             string         `json:"host"`
	Port             string         `json:"port"`
	Scheme           string         `json:"scheme"`
	URL              string         `json:"url"`
	Path             string         `json:"path"`
	HitCredential    map[string]any `json:"hit_credential"`
	Matched          string         `json:"matched"`
	ExtractedResults []string       `json:"extracted_results"`
	Description      string         `json:"description"`
	Remediation      string         `json:"remediation"`
	Tags             string         `json:"tags"`
}

func NewJobResultItem() *JobResultItem {
	return &JobResultItem{}
}

func (jr *JobResultItem) WithEntryID(entryID string) *JobResultItem {
	jr.EntryID = entryID
	return jr
}

func (jr *JobResultItem) Fill(event *output.ResultEvent) *JobResultItem {
	jr.TemplateID = event.TemplateID
	jr.TemplateName = event.Info.Name
	jr.Type = event.Type
	jr.Severity = event.Info.SeverityHolder.Severity.String()
	jr.Host = lo.If(event.IP != "", event.IP).Else(strings.Split(event.Host, ":")[0])
	jr.Port = event.Port
	jr.Scheme = event.Scheme
	jr.URL = event.URL
	jr.Path = event.Path
	jr.HitCredential = event.Metadata
	jr.Matched = event.Matched
	jr.ExtractedResults = event.ExtractedResults
	jr.Description = event.Info.Description
	jr.Remediation = event.Info.Remediation
	jr.Tags = event.Info.Tags.String()

	return jr
}

type PortResult struct {
	EntryID string            `json:"-"`
	Items   []*PortResultItem `json:"items"`
}

type PortResultItem struct {
	EntryID  string `json:"-"`
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	HostPort string `json:"host_port"`
}

type PingResult struct {
	EntryID string            `json:"-"`
	Items   []*PingResultItem `json:"items"`
}

type PingResultItem struct {
	EntryID string `json:"-"`
	IP      string `json:"ip"`
	OS      string `json:"os"`
	TTL     int    `json:"ttl"`
	Active  bool   `json:"active"`
}

type EntryResult struct {
	EntryID             string       `json:"-"`
	HostDiscoveryResult *PingResult  `json:"host_discovery_result"`
	PortScanningResult  *PortResult  `json:"plan_scanning_result"`
	JobResults          []*JobResult `json:"job_results"`
	Targets             []string     `json:"-"`
	ExcludeTargets      []string     `json:"-"`
	StartTime           time.Time    `json:"-"`
	EndTime             time.Time    `json:"-"`
}

// PingResultCallback ping结果回调
type PingResultCallback func(context.Context, *PingResult) error

// PortResultCallback port结果回调
type PortResultCallback func(context.Context, *PortResult) error

// JobResultCallback job结果回调
type JobResultCallback func(context.Context, *JobResult) error

// RawTemplate 原始模板
type RawTemplate struct {
	ID       string // 模板ID
	Original string // 模板内容
}

// GetTemplates 获取模板
type GetTemplates func() []*RawTemplate

type Stage struct {
	Name    StageName
	Percent float64
	Entries map[StageEntryName]any
}

type StageName string
type StageEntryName string

const (
	StagePreExecute    StageName = "PreExecute"
	StageHostDiscovery StageName = "HostDiscovery"
	StagePortScanning  StageName = "PortScanning"
	StageJob           StageName = "Job"
	StagePostExecute   StageName = "PostExecute"
)

const (
	StageEntryJobKind  StageEntryName = "Kind"
	StageEntryJobIndex StageEntryName = "Index"
)

type ResultReader struct {
	Format string
	Stage  StageName
	Reader io.Reader
}
