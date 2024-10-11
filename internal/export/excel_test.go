package export

import (
	"context"
	"os"
	"testing"

	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

func TestExcelExporter(t *testing.T) {
	assert := assert.New(t)

	{
		exporter, err := NewExcelExporter("testScanner")
		assert.NoError(err)
		assert.NotNil(exporter)
		assert.IsType(&excelExporter{}, exporter)
		exporter.Close()
	}

	{
		exporter, err := NewExcelExporter("testScanner", []any{"a", "b", "c", "d"}...)
		assert.NoError(err)
		assert.NotNil(exporter)

		assert.IsType(&excelExporter{}, exporter)
		excelExporter := exporter.(*excelExporter)

		err = excelExporter.Export(context.Background(), []any{"是", "content1", "否", "content2"})
		assert.NoError(err)

		assert.Equal(uint32(2), excelExporter.row.Load())
		exporter.Close()

		f, err := excelize.OpenFile("./testScanner.xlsx")
		assert.NoError(err)
		defer f.Close()

		a1, err := f.GetCellValue("Sheet1", "A1")
		assert.NoError(err)
		assert.Equal("a", a1)
		c1, err := f.GetCellValue("Sheet1", "C1")
		assert.NoError(err)
		assert.Equal("c", c1)
		a2, err := f.GetCellValue("Sheet1", "A2")
		assert.NoError(err)
		assert.Equal("是", a2)
		c2, err := f.GetCellValue("Sheet1", "C2")
		assert.NoError(err)
		assert.Equal("否", c2)
		s1, err := f.GetCellStyle("Sheet1", "A1")
		assert.NoError(err)
		s2, err := f.GetCellStyle("Sheet1", "C1")
		assert.NoError(err)
		assert.Equal(excelExporter.styles[Header], s1)
		assert.Equal(excelExporter.styles[Header], s2)
		s3, err := f.GetCellStyle("Sheet1", "A2")
		assert.NoError(err)
		s4, err := f.GetCellStyle("Sheet1", "C2")
		assert.NoError(err)
		assert.Equal(excelExporter.styles[true], s3)
		assert.Equal(excelExporter.styles[false], s4)

		os.Remove("./testScanner.xlsx")
	}
}

func TestGtSheetMaxRow(t *testing.T) {
	defer os.Remove("./testScanner.xlsx")

	assert := assert.New(t)

	exporter, err := NewExcelExporter("testScanner", []any{"TITLE"}...)
	assert.NoError(err)

	for i := range 1_500_000 {
		exporter.Export(context.Background(), []any{i})
	}
	exporter.Close()

	f, err := os.Open("./testScanner.xlsx")
	assert.NoError(err)
	defer f.Close()

	reader, err := excelize.OpenReader(f)
	assert.NoError(err)
	defer reader.Close()

	contents, err := util.ReadXlsxAll(reader)
	assert.NoError(err)

	assert.Equal(1_500_001, len(contents))
}
