package job

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/EscapeBearSecond/falcon/pkg/types"
	"github.com/pterm/pterm"
	"github.com/samber/lo"
)

var consoleHeaders = []string{
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
}

var _ exporter = (*consoleExporter)(nil)

// consoleExporter 控制台输出
type consoleExporter struct {
	resultPool
	table [][]string
	m     sync.Mutex
}

func newConsoleExporter() *consoleExporter {
	exporter := &consoleExporter{}
	exporter.table = [][]string{
		consoleHeaders,
	}
	return exporter
}

func (e *consoleExporter) Export(c context.Context, result *types.JobResultItem) error {
	select {
	case <-c.Done():
	default:
		e.m.Lock()
		e.table = append(e.table, []string{
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
		})
		e.m.Unlock()

		e.PutResult(result)
	}

	return nil
}

func (e *consoleExporter) Stop() error {
	if len(e.table) != 1 {
		pterm.DefaultTable.WithHasHeader().
			WithRowSeparator("-").
			WithHeaderRowSeparator("-").
			WithData(e.table).
			Render()
	}
	return nil
}
