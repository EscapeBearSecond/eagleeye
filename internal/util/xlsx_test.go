package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

func TestReadXlsxAll(t *testing.T) {
	defer os.Remove(os.TempDir() + "/test.xlsx")

	assert := assert.New(t)

	f := excelize.NewFile()
	f.SetSheetRow("Sheet1", "A1", &[]string{"a", "b", "c"})
	f.SetSheetRow("Sheet1", "A2", &[]string{"1", "2", "3"})

	f.SaveAs(os.TempDir() + "/test.xlsx")
	defer f.Close()

	contents, err := ReadXlsxAll(f)
	assert.Nil(err)
	assert.Equal([][]string{{"a", "b", "c"}, {"1", "2", "3"}}, contents)
}
