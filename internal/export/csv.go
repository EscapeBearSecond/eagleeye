package export

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"sync"
	"time"
)

// csvExporter csv输出
type csvExporter struct {
	writer  *csv.Writer
	f       *os.File
	m       sync.Mutex
	scanner string

	c        context.Context
	stopLoop context.CancelFunc
	loopDone chan struct{}
}

// NewCsvExporter 实例化csv导出器
func NewCsvExporter(scanner string, headers ...any) (Exporter, error) {
	file, err := os.OpenFile(fmt.Sprintf("%s.csv", scanner), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("create [%s] output file failed: %w", scanner, err)
	}

	exporter := &csvExporter{
		writer:  csv.NewWriter(file),
		f:       file,
		scanner: scanner,
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("get [%s] output file stat failed: %w", scanner, err)
	}

	if fileInfo.Size() == 0 && len(headers) != 0 {
		result := make([]string, 0, len(headers))
		for _, vv := range headers {
			result = append(result, fmt.Sprintf("%v", vv))
		}
		err = exporter.writer.Write(result)
		if err != nil {
			return nil, fmt.Errorf("create [%s] headers failed: %w", scanner, err)
		}
		exporter.writer.Flush()
	}

	exporter.c, exporter.stopLoop = context.WithCancel(context.Background())
	exporter.loopDone = make(chan struct{}, 1)
	go exporter.loopFlush()

	return exporter, nil
}

func (exp *csvExporter) loopFlush() {
	defer close(exp.loopDone)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			exp.m.Lock()
			exp.writer.Flush()
			exp.m.Unlock()
		case <-exp.c.Done():
			return
		}
	}
}

// Export 导出
func (exp *csvExporter) Export(c context.Context, v []any) error {
	result := make([]string, 0, len(v))
	for _, vv := range v {
		result = append(result, fmt.Sprintf("%v", vv))
	}
	exp.m.Lock()
	defer exp.m.Unlock()
	err := exp.writer.Write(result)
	if err != nil {
		return err
	}

	return nil
}

// Close 关闭
func (exp *csvExporter) Close() {
	exp.stopLoop()
	<-exp.loopDone

	// csv刷新写入文件
	exp.writer.Flush()
	exp.f.Close()
}
