package util

import (
	"encoding/csv"
	"os"
	"testing"

	"github.com/EscapeBearSecond/eagleeye/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

func TestReloadResultFromReader(t *testing.T) {
	defer os.Remove("./在线检测.csv")
	defer os.Remove("./端口扫描.xlsx")

	assert := assert.New(t)

	csvFile, err := os.Create("./在线检测.csv")
	assert.NoError(err)
	f1 := csv.NewWriter(csvFile)
	f1.Write([]string{"主机", "存活", "系统", "TTL"})
	f1.Write([]string{"1.1.1.1", "是", "linux", "10"})
	f1.Flush()
	csvFile.Close()

	f2 := excelize.NewFile()
	f2.SetSheetRow("Sheet1", "A1", &[]string{"主机", "端口"})
	f2.SetSheetRow("Sheet1", "A2", &[]string{"1.1.1.1", "22"})
	f2.SetSheetRow("Sheet1", "A3", &[]string{"1.1.1.1", "23"})
	f2.SaveAs("./端口扫描.xlsx")
	f2.Close()

	hd, err := os.Open("./在线检测.csv")
	assert.NoError(err)
	defer hd.Close()

	ps, err := os.Open("./端口扫描.xlsx")
	assert.NoError(err)
	defer ps.Close()

	r, err := ReloadResult(&types.ResultReader{
		Format: "csv",
		Stage:  types.StageHostDiscovery,
		Reader: hd,
	}, &types.ResultReader{
		Format: "excel",
		Stage:  types.StagePortScanning,
		Reader: ps,
	})
	assert.NoError(err)

	assert.Len(r.HostDiscoveryResult.Items, 1)
	assert.Len(r.PortScanningResult.Items, 2)
}
