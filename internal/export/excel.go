package export

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/EscapeBearSecond/falcon/internal/util"
	"github.com/xuri/excelize/v2"
)

// excelExporter excel输出
type excelExporter struct {
	f       *excelize.File
	row     *atomic.Uint32
	styles  styles
	scanner string
	m       sync.Mutex

	c        context.Context
	stopLoop context.CancelFunc
	loopDone chan struct{}
}

// NewCsvExporter 实例化excel导出器
func NewExcelExporter(scanner string, headers ...any) (Exporter, error) {
	exporter := &excelExporter{
		scanner: scanner,
		row:     &atomic.Uint32{},
	}

	filename := fmt.Sprintf("%s.xlsx", scanner)

	var f *excelize.File
	_, err := os.Stat(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("get [%s] output file stat failed: %w", scanner, err)
		}

		f = excelize.NewFile()
	} else {
		f, err = excelize.OpenFile(filename)
		if err != nil {
			return nil, fmt.Errorf("open [%s] output file failed: %w", scanner, err)
		}
		size, err := util.XlsxRowsSize(f)
		if err != nil {
			return nil, fmt.Errorf("get [%s] rows size failed: %w", scanner, err)
		}
		exporter.row.Store(size)
	}

	exporter.f = f

	// 表头格式
	headerStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"A6DCED"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create [%s] style failed: %w", scanner, err)
	}

	// 否定状态的格式
	negStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"FF0000"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create [%s] style failed: %w", scanner, err)
	}

	// 肯定状态的格式
	posStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"00FF00"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create [%s] style failed: %w", scanner, err)
	}

	// 配置格式
	exporter.styles = newStyles(
		newStyleItem(Positive, posStyle),
		newStyleItem(Negative, negStyle),
		newStyleItem(Header, headerStyle),
	)

	if exporter.row.Load() == 0 && len(headers) > 0 {
		exporter.row.Add(1)

		err = f.SetCellStyle("Sheet1", "A1", fmt.Sprintf("%c1", byte('A')+uint8(len(headers))-1), exporter.styles[Header])
		if err != nil {
			return nil, fmt.Errorf("set [%s] headers style failed: %w", scanner, err)
		}
		err = f.SetSheetRow("Sheet1", "A1", &headers)
		if err != nil {
			return nil, fmt.Errorf("create [%s] headers failed: %w", scanner, err)
		}
	}

	// 文件不存在，创建文件
	if f.Path == "" {
		err = f.SaveAs(fmt.Sprintf("%s.xlsx", scanner))
		if err != nil {
			return nil, fmt.Errorf("create [%s] output file failed: %w", scanner, err)
		}
	} else {
		err = f.Save()
		if err != nil {
			return nil, fmt.Errorf("save [%s] output file failed: %w", scanner, err)
		}
	}

	exporter.c, exporter.stopLoop = context.WithCancel(context.Background())
	exporter.loopDone = make(chan struct{}, 1)
	go exporter.loopSave()

	return exporter, nil
}

func (exp *excelExporter) loopSave() {
	defer close(exp.loopDone)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			exp.m.Lock()
			exp.f.Save()
			exp.m.Unlock()
		case <-exp.c.Done():
			return
		}
	}
}

// Export 导出
func (exp *excelExporter) Export(c context.Context, v []any) error {
	row := exp.row.Add(1)

	exp.m.Lock()
	defer exp.m.Unlock()
	err := exp.exportRow(c, row, v)
	if err != nil {
		return err
	}

	return nil
}

func (exp *excelExporter) exportRow(_ context.Context, rowNum uint32, v []any) error {
	pageNum := (rowNum-1)/1_048_576 + 1
	rowNum = (rowNum-1)%1_048_576 + 1

	sheetName := fmt.Sprintf("Sheet%d", pageNum)
	if pageNum != uint32(exp.f.SheetCount) {
		exp.f.NewSheet(sheetName)
	}

	// 配置单元格格式
	for i, vv := range v {
		col := byte('A') + byte(i)
		if st := exp.styles.style(vv); st != 0 {
			exp.f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col, rowNum), fmt.Sprintf("%c%d", col, rowNum), st)
		}
	}

	exp.f.SetSheetRow(sheetName, fmt.Sprintf("A%d", rowNum), &v)

	return nil
}

// Close 关闭
func (exp *excelExporter) Close() {
	exp.stopLoop()
	<-exp.loopDone

	exp.f.Save()
	exp.f.Close()
}
