package job

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/EscapeBearSecond/eagleeye/internal/export"
	"github.com/EscapeBearSecond/eagleeye/pkg/types"
	"github.com/samber/lo"
)

var fileHeaders = []any{
	headerTemplateID,
	headerTemplateName,
	headerTemplateType,
	headerTemplateSeverity,
	headerTemplateTags,
	headerHost,
	headerPort,
	headerScheme,
	headerURL,
	headerPath,
	headerHitCredential,
	headerMatched,
	headerExtractedResults,
	headerDescription,
	headerRemediation,
}

var _ exporter = (*fileExporter)(nil)

// fileExporter 支持csv,excel输出
//
// 实际调用了export的Exporter
type fileExporter struct {
	resultPool

	ctx    context.Context
	cancel context.CancelFunc

	workerC  chan *types.JobResultItem
	exporter export.Exporter
	workers  sync.WaitGroup
}

// 实例化file输出，支持csv和excel格式
func newFileExporter(format exportFormat, name string) (*fileExporter, error) {
	var (
		exporter export.Exporter
		err      error
	)

	switch format {
	case csvMode:
		exporter, err = export.NewCsvExporter(name, fileHeaders...)
		if err != nil {
			return nil, fmt.Errorf("create csv exporter failed: %w", err)
		}
	case excelMode:
		exporter, err = export.NewExcelExporter(name, fileHeaders...)
		if err != nil {
			return nil, fmt.Errorf("create excel exporter failed: %w", err)
		}
	}

	c, cancel := context.WithCancel(context.Background())
	exp := &fileExporter{
		ctx:      c,
		cancel:   cancel,
		workerC:  make(chan *types.JobResultItem, 500),
		workers:  sync.WaitGroup{},
		exporter: exporter,
	}

	go exp.runWorkers(2)
	return exp, nil
}

func (exp *fileExporter) Export(c context.Context, result *types.JobResultItem) error {
	select {
	case <-c.Done():
		exp.cancel()
	case exp.workerC <- result:
	}
	return nil
}

func (exp *fileExporter) runWorkers(size int) {
	exp.workers.Add(size)

	for range size {
		go exp.work()
	}
}

func (exp *fileExporter) work() {
	defer exp.workers.Done()

	for {
		select {
		case <-exp.ctx.Done():
			return
		case result := <-exp.workerC:
			exp.exporter.Export(exp.ctx, []any{
				result.TemplateID,
				result.TemplateName,
				result.Type,
				result.Severity,
				result.Tags,
				result.Host,
				result.Port,
				result.Scheme,
				result.URL,
				result.Path,
				strings.Join(lo.MapToSlice(result.HitCredential, func(k string, v any) string { return fmt.Sprintf("%s:%v", k, v) }), "|"),
				result.Matched,
				strings.Join(result.ExtractedResults, "|"),
				result.Description,
				result.Remediation,
			})

			exp.PutResult(result)
		}
	}
}

func (exp *fileExporter) Stop() error {
	exp.cancel()
	exp.workers.Wait()

	exp.exporter.Close()
	close(exp.workerC)

	return nil
}
