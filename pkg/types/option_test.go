package types

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	assert := assert.New(t)

	o := &Options{}

	err := os.WriteFile(filepath.Join(os.TempDir(), "cfg.yaml"), []byte(`targets:
  - 192.168.1.0-192.168.1.107
  - 192.168.1.109-192.168.1.255
out_log: true
monitor:
  use: true
  interval: 1s
host_discovery:
  use: true
  timeout: 1s
  concurrency: 100
  rate_limit: 1000
  format: excel
port_scanning:
  use: true
  count: 1
  concurrency: 100
  rate_limit: 1000
  format: excel
jobs:
  - name: 漏洞扫描
    rate_limit: 2000
    format: csv
    timeout: 1s
    count: 1
    template: ./templates/漏洞扫描
  - name: 资产扫描
    concurrency: 500
    timeout: 1s
    count: 1
    template: ./templates/资产识别`), 0644)
	assert.NoError(err)

	err = o.Parse(filepath.Join(os.TempDir(), "cfg.yaml"))
	assert.NoError(err)

	assert.Equal("192.168.1.0-192.168.1.107", o.Targets[0])
	assert.Equal("192.168.1.109-192.168.1.255", o.Targets[1])

	assert.Equal(1, o.HostDiscovery.Count)

	assert.Equal("1s", o.PortScanning.Timeout)
	assert.Equal("http", o.PortScanning.Ports)

	assert.Equal(150, o.Jobs[0].Concurrency)

	assert.Equal(150, o.Jobs[1].RateLimit)
	assert.Equal("csv", o.Jobs[1].Format)

	os.Remove(filepath.Join(os.TempDir(), "cfg.yaml"))
}
