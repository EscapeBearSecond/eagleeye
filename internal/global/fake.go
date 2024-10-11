package global

import (
	"github.com/logrusorgru/aurora"
	"github.com/projectdiscovery/gologger/levels"
	"github.com/projectdiscovery/nuclei/v3/pkg/output"
	"github.com/projectdiscovery/nuclei/v3/pkg/progress"
)

var _ progress.Progress = (*fakeProgress)(nil)

// fakeProgress空进度
type fakeProgress struct{}

func (p *fakeProgress) Stop()                                                    {}
func (p *fakeProgress) Init(hostCount int64, rulesCount int, requestCount int64) {}
func (p *fakeProgress) AddToTotal(delta int64)                                   {}
func (p *fakeProgress) IncrementRequests()                                       {}
func (p *fakeProgress) IncrementMatched()                                        {}
func (p *fakeProgress) IncrementErrorsBy(count int64)                            {}
func (p *fakeProgress) IncrementFailedRequestsBy(count int64)                    {}
func (p *fakeProgress) SetRequests(count uint64)                                 {}

var _ output.Writer = (*fakeWriter)(nil)

// fakeWriter空输出
type fakeWriter struct{}

func (r *fakeWriter) Close()                                                              {}
func (r *fakeWriter) Colorizer() aurora.Aurora                                            { return nil }
func (r *fakeWriter) WriteFailure(event *output.InternalWrappedEvent) error               { return nil }
func (r *fakeWriter) Write(w *output.ResultEvent) error                                   { return nil }
func (r *fakeWriter) Request(templateID, url, requestType string, err error)              {}
func (r *fakeWriter) WriteStoreDebugData(host, templateID, eventType string, data string) {}

// nopWriter 空写（用于屏蔽端口扫描默认日志）
type nopWriter struct{}

func (r *nopWriter) Write(data []byte, level levels.Level) {}
