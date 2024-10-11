package eagleeye

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"testing"
	"time"

	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/internal/util"
	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

func TestNewEngine(t *testing.T) {
	defer os.RemoveAll("./results")

	assert := assert.New(t)

	engine, err := NewEngine(WithDirectory("./results"))
	assert.NoError(err)
	defer engine.Close()

	assert.Equal(engine.dir, "./results")
	assert.Len(engine.entries, 0)
}

func TestNewEntry(t *testing.T) {
	defer os.RemoveAll("./results")

	assert := assert.New(t)

	engine, err := NewEngine(WithDirectory("./results"))
	assert.NoError(err)
	defer engine.Close()

	options := &types.Options{
		Targets: []string{
			"192.168.1.0-192.168.1.255",
		},
		ExcludeTargets: []string{
			"192.168.1.108",
		},
		PortScanning: types.PortScanningOptions{
			Use:         true,
			Timeout:     "2s",
			Count:       1,
			Format:      "csv",
			Ports:       "http",
			RateLimit:   1000,
			Concurrency: 1000,
		},
		HostDiscovery: types.HostDiscoveryOptions{
			Use:         true,
			Timeout:     "2s",
			Count:       1,
			Format:      "csv",
			RateLimit:   1000,
			Concurrency: 1000,
		},
		Jobs: []types.JobOptions{
			{
				Name: "资产扫描",
				Kind: "asset-scan",
				GetTemplates: func() []*types.RawTemplate {
					var templates []*types.RawTemplate
					templates = append(templates, &types.RawTemplate{
						ID: "pgsql-detect",
						Original: `id: pgsql-detect

info:
  name: PostgreSQL Authentication - Detect
  author: nybble04,geeknik
  severity: info
  description: |
    PostgreSQL authentication error messages which could reveal information useful in formulating further attacks were detected.
  reference:
    - https://www.postgresql.org/docs/current/errcodes-appendix.html
    - https://www.postgresql.org/docs/current/client-authentication-problems.html
  classification:
    cvss-metrics: CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:N
    cwe-id: CWE-200
  metadata:
    max-request: 1
    shodan-query: port:5432 product:"PostgreSQL"
    verified: true
  tags: network,postgresql,db,detect

tcp:
  - inputs:
      - data: "000000500003000075736572006e75636c6569006461746162617365006e75636c6569006170706c69636174696f6e5f6e616d65007073716c00636c69656e745f656e636f64696e6700555446380000"
        type: hex
      - data: "7000000036534352414d2d5348412d32353600000000206e2c2c6e3d2c723d000000000000000000000000000000000000000000000000"
        type: hex

    host:
      - "{{Hostname}}"
    port: 5432
    read-size: 2048

    matchers-condition: and
    matchers:
      - type: word
        part: body
        words:
          - "C0A000"                  # Error code for unsupported frontend protocol
          - "C08P01"                  # Error code for invalide startup packet layout
          - "28000"                   # Error code for invalid_authorization_specification
          - "28P01"                   # Error code for invalid_password
          - "SCRAM-SHA-256"           # Authentication prompt
          - "pg_hba.conf"             # Client authentication config file
          - "user \"nuclei\""         # The user nuclei (sent in request) doesn't exist
          - "database \"nuclei\""     # The db nuclei (sent in request) doesn't exist"
        condition: or

      - type: word
        words:
          - "HTTP/1.1"
        negative: true
# digest: 4a0a004730450220190550562f0223183090e8ca4117ace44d725bdece7b84c58edaed8d93935aa7022100872d6d635b69589e7e99749cae0639a48551bbdee2d3d7038aa4699257a00383:922c64590222798bb761d5b6d8e72950`,
					})
					return templates
				},
				Format:      "csv",
				Count:       1,
				Timeout:     "2s",
				RateLimit:   2000,
				Concurrency: 2000,
			},
		},
	}

	entry, err := engine.NewEntry(options)
	assert.NoError(err)
	defer entry.cancel()
}

func TestEntryRun(t *testing.T) {
	defer os.RemoveAll("./results")

	assert := assert.New(t)

	engine, err := NewEngine(WithDirectory("./results"))
	assert.NoError(err)
	defer engine.Close()

	options := &types.Options{
		Targets: []string{
			"192.168.1.0-192.168.1.255",
		},
		ExcludeTargets: []string{
			"192.168.1.108",
		},
		PortScanning: types.PortScanningOptions{
			Use:         true,
			Timeout:     "2s",
			Count:       1,
			Format:      "csv",
			Ports:       "http",
			RateLimit:   1000,
			Concurrency: 1000,
		},
		HostDiscovery: types.HostDiscoveryOptions{
			Use:         true,
			Timeout:     "2s",
			Count:       1,
			Format:      "csv",
			RateLimit:   1000,
			Concurrency: 1000,
		},
		Jobs: []types.JobOptions{
			{
				Name: "资产扫描",
				Kind: "asset-scan",
				GetTemplates: func() []*types.RawTemplate {
					var templates []*types.RawTemplate
					templates = append(templates, &types.RawTemplate{
						ID: "pgsql-detect",
						Original: `id: pgsql-detect

info:
  name: PostgreSQL Authentication - Detect
  author: nybble04,geeknik
  severity: info
  description: |
    PostgreSQL authentication error messages which could reveal information useful in formulating further attacks were detected.
  reference:
    - https://www.postgresql.org/docs/current/errcodes-appendix.html
    - https://www.postgresql.org/docs/current/client-authentication-problems.html
  classification:
    cvss-metrics: CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:N
    cwe-id: CWE-200
  metadata:
    max-request: 1
    shodan-query: port:5432 product:"PostgreSQL"
    verified: true
  tags: network,postgresql,db,detect

tcp:
  - inputs:
      - data: "000000500003000075736572006e75636c6569006461746162617365006e75636c6569006170706c69636174696f6e5f6e616d65007073716c00636c69656e745f656e636f64696e6700555446380000"
        type: hex
      - data: "7000000036534352414d2d5348412d32353600000000206e2c2c6e3d2c723d000000000000000000000000000000000000000000000000"
        type: hex

    host:
      - "{{Hostname}}"
    port: 5432
    read-size: 2048

    matchers-condition: and
    matchers:
      - type: word
        part: body
        words:
          - "C0A000"                  # Error code for unsupported frontend protocol
          - "C08P01"                  # Error code for invalide startup packet layout
          - "28000"                   # Error code for invalid_authorization_specification
          - "28P01"                   # Error code for invalid_password
          - "SCRAM-SHA-256"           # Authentication prompt
          - "pg_hba.conf"             # Client authentication config file
          - "user \"nuclei\""         # The user nuclei (sent in request) doesn't exist
          - "database \"nuclei\""     # The db nuclei (sent in request) doesn't exist"
        condition: or

      - type: word
        words:
          - "HTTP/1.1"
        negative: true
# digest: 4a0a004730450220190550562f0223183090e8ca4117ace44d725bdece7b84c58edaed8d93935aa7022100872d6d635b69589e7e99749cae0639a48551bbdee2d3d7038aa4699257a00383:922c64590222798bb761d5b6d8e72950`,
					})
					return templates
				},
				Format:      "csv",
				Count:       1,
				Timeout:     "2s",
				RateLimit:   2000,
				Concurrency: 2000,
			},
		},
	}

	entry, err := engine.NewEntry(options)
	assert.NoError(err)

	stopPrintStage := make(chan struct{}, 1)
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				t.Logf("stage: %#v\n", entry.Stage())
			case <-stopPrintStage:
				return
			}
		}
	}()
	defer close(stopPrintStage)

	err = entry.Run(context.Background())
	assert.NoError(err)

	ret := entry.Result()
	t.Logf("result: %#v\n", ret)
}

func TestEntryStop(t *testing.T) {
	defer os.RemoveAll("./results")

	assert := assert.New(t)

	engine, err := NewEngine(WithDirectory("./results"))
	assert.NoError(err)
	defer engine.Close()

	options := &types.Options{
		Targets: []string{
			"192.168.1.0-192.168.1.255",
		},
		ExcludeTargets: []string{
			"192.168.1.108",
		},
		PortScanning: types.PortScanningOptions{
			Use:         true,
			Timeout:     "2s",
			Count:       1,
			Format:      "csv",
			Ports:       "http",
			RateLimit:   1000,
			Concurrency: 1000,
		},
		HostDiscovery: types.HostDiscoveryOptions{
			Use:         true,
			Timeout:     "2s",
			Count:       1,
			Format:      "csv",
			RateLimit:   1000,
			Concurrency: 1000,
		},
		Jobs: []types.JobOptions{
			{
				Name: "资产扫描",
				Kind: "asset-scan",
				GetTemplates: func() []*types.RawTemplate {
					var templates []*types.RawTemplate
					templates = append(templates, &types.RawTemplate{
						ID: "pgsql-detect",
						Original: `id: pgsql-detect

info:
  name: PostgreSQL Authentication - Detect
  author: nybble04,geeknik
  severity: info
  description: |
    PostgreSQL authentication error messages which could reveal information useful in formulating further attacks were detected.
  reference:
    - https://www.postgresql.org/docs/current/errcodes-appendix.html
    - https://www.postgresql.org/docs/current/client-authentication-problems.html
  classification:
    cvss-metrics: CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:N
    cwe-id: CWE-200
  metadata:
    max-request: 1
    shodan-query: port:5432 product:"PostgreSQL"
    verified: true
  tags: network,postgresql,db,detect

tcp:
  - inputs:
      - data: "000000500003000075736572006e75636c6569006461746162617365006e75636c6569006170706c69636174696f6e5f6e616d65007073716c00636c69656e745f656e636f64696e6700555446380000"
        type: hex
      - data: "7000000036534352414d2d5348412d32353600000000206e2c2c6e3d2c723d000000000000000000000000000000000000000000000000"
        type: hex

    host:
      - "{{Hostname}}"
    port: 5432
    read-size: 2048

    matchers-condition: and
    matchers:
      - type: word
        part: body
        words:
          - "C0A000"                  # Error code for unsupported frontend protocol
          - "C08P01"                  # Error code for invalide startup packet layout
          - "28000"                   # Error code for invalid_authorization_specification
          - "28P01"                   # Error code for invalid_password
          - "SCRAM-SHA-256"           # Authentication prompt
          - "pg_hba.conf"             # Client authentication config file
          - "user \"nuclei\""         # The user nuclei (sent in request) doesn't exist
          - "database \"nuclei\""     # The db nuclei (sent in request) doesn't exist"
        condition: or

      - type: word
        words:
          - "HTTP/1.1"
        negative: true
# digest: 4a0a004730450220190550562f0223183090e8ca4117ace44d725bdece7b84c58edaed8d93935aa7022100872d6d635b69589e7e99749cae0639a48551bbdee2d3d7038aa4699257a00383:922c64590222798bb761d5b6d8e72950`,
					})
					return templates
				},
				Format:      "csv",
				Count:       1,
				Timeout:     "2s",
				RateLimit:   2000,
				Concurrency: 2000,
			},
		},
	}

	entry, err := engine.NewEntry(options)
	assert.NoError(err)

	go func() {
		time.Sleep(5 * time.Second)
		entry.Stop()
	}()

	err = entry.Run(context.Background())
	assert.ErrorIs(err, types.ErrHasBeenStopped)

	ret := entry.Result()
	assert.Nil(ret)
}

func TestCsvReuseEntryID(t *testing.T) {
	assert := assert.New(t)

	engine, err := NewEngine()
	assert.NoError(err)
	defer engine.Close()

	options := &types.Options{
		Targets: []string{
			"192.168.1.100-192.168.1.200",
		},
		ExcludeTargets: []string{
			"192.168.1.108",
		},
		PortScanning: types.PortScanningOptions{
			Use:         true,
			Timeout:     "2s",
			Count:       1,
			Format:      "csv",
			Ports:       "http",
			RateLimit:   1000,
			Concurrency: 1000,
		},
		HostDiscovery: types.HostDiscoveryOptions{
			Use:         true,
			Timeout:     "2s",
			Count:       1,
			Format:      "csv",
			RateLimit:   1000,
			Concurrency: 1000,
		},
		Jobs: []types.JobOptions{
			{
				Name: "资产扫描",
				Kind: "asset-scan",
				GetTemplates: func() []*types.RawTemplate {
					var templates []*types.RawTemplate
					templates = append(templates, &types.RawTemplate{
						ID: "pgsql-detect",
						Original: `id: pgsql-detect

info:
  name: PostgreSQL Authentication - Detect
  author: nybble04,geeknik
  severity: info
  description: |
    PostgreSQL authentication error messages which could reveal information useful in formulating further attacks were detected.
  reference:
    - https://www.postgresql.org/docs/current/errcodes-appendix.html
    - https://www.postgresql.org/docs/current/client-authentication-problems.html
  classification:
    cvss-metrics: CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:N
    cwe-id: CWE-200
  metadata:
    max-request: 1
    shodan-query: port:5432 product:"PostgreSQL"
    verified: true
  tags: network,postgresql,db,detect

tcp:
  - inputs:
      - data: "000000500003000075736572006e75636c6569006461746162617365006e75636c6569006170706c69636174696f6e5f6e616d65007073716c00636c69656e745f656e636f64696e6700555446380000"
        type: hex
      - data: "7000000036534352414d2d5348412d32353600000000206e2c2c6e3d2c723d000000000000000000000000000000000000000000000000"
        type: hex

    host:
      - "{{Hostname}}"
    port: 5432
    read-size: 2048

    matchers-condition: and
    matchers:
      - type: word
        part: body
        words:
          - "C0A000"                  # Error code for unsupported frontend protocol
          - "C08P01"                  # Error code for invalide startup packet layout
          - "28000"                   # Error code for invalid_authorization_specification
          - "28P01"                   # Error code for invalid_password
          - "SCRAM-SHA-256"           # Authentication prompt
          - "pg_hba.conf"             # Client authentication config file
          - "user \"nuclei\""         # The user nuclei (sent in request) doesn't exist
          - "database \"nuclei\""     # The db nuclei (sent in request) doesn't exist"
        condition: or

      - type: word
        words:
          - "HTTP/1.1"
        negative: true
# digest: 4a0a004730450220190550562f0223183090e8ca4117ace44d725bdece7b84c58edaed8d93935aa7022100872d6d635b69589e7e99749cae0639a48551bbdee2d3d7038aa4699257a00383:922c64590222798bb761d5b6d8e72950`,
					})
					return templates
				},
				Format:      "csv",
				Count:       1,
				Timeout:     "2s",
				RateLimit:   2000,
				Concurrency: 2000,
			},
		},
	}

	id := NewID()
	defer os.RemoveAll(fmt.Sprintf("./%s", id))

	entry, err := engine.NewEntry(options, WithID(id))
	assert.NoError(err)

	err = entry.Run(context.Background())
	assert.NoError(err)

	pingSize := len(entry.Result().HostDiscoveryResult.Items)
	portSize := len(entry.Result().PortScanningResult.Items)
	jobSize := len(entry.Result().JobResults[0].Items)

	entry2, err := engine.NewEntry(options, WithID(id))
	assert.NoError(err)

	err = entry2.Run(context.Background())
	assert.NoError(err)

	pingSize += len(entry2.Result().HostDiscoveryResult.Items)
	portSize += len(entry2.Result().PortScanningResult.Items)
	jobSize += len(entry2.Result().JobResults[0].Items)

	f1, err := os.Open(fmt.Sprintf("./%s/在线检测.csv", id))
	assert.NoError(err)
	content1, err := csv.NewReader(f1).ReadAll()
	assert.NoError(err)
	size1 := len(content1)
	f1.Close()

	f2, err := os.Open(fmt.Sprintf("./%s/端口扫描.csv", id))
	assert.NoError(err)
	content2, err := csv.NewReader(f2).ReadAll()
	assert.NoError(err)
	size2 := len(content2)
	f2.Close()

	f3, err := os.Open(fmt.Sprintf("./%s/资产扫描.csv", id))
	assert.NoError(err)
	content3, err := csv.NewReader(f3).ReadAll()
	assert.NoError(err)
	size3 := len(content3)
	f3.Close()

	assert.Equal(pingSize, size1-1)
	assert.Equal(portSize, size2-1)
	assert.Equal(jobSize, size3-1)
}

func TestExcelReuseEntryID(t *testing.T) {
	assert := assert.New(t)

	engine, err := NewEngine()
	assert.NoError(err)
	defer engine.Close()

	options := &types.Options{
		Targets: []string{
			"192.168.1.100-192.168.1.200",
		},
		ExcludeTargets: []string{
			"192.168.1.108",
		},
		PortScanning: types.PortScanningOptions{
			Use:         true,
			Timeout:     "2s",
			Count:       1,
			Format:      "excel",
			Ports:       "http",
			RateLimit:   1000,
			Concurrency: 1000,
		},
		HostDiscovery: types.HostDiscoveryOptions{
			Use:         true,
			Timeout:     "2s",
			Count:       1,
			Format:      "excel",
			RateLimit:   1000,
			Concurrency: 1000,
		},
		Jobs: []types.JobOptions{
			{
				Name: "资产扫描",
				Kind: "asset-scan",
				GetTemplates: func() []*types.RawTemplate {
					var templates []*types.RawTemplate
					templates = append(templates, &types.RawTemplate{
						ID: "pgsql-detect",
						Original: `id: pgsql-detect

info:
  name: PostgreSQL Authentication - Detect
  author: nybble04,geeknik
  severity: info
  description: |
    PostgreSQL authentication error messages which could reveal information useful in formulating further attacks were detected.
  reference:
    - https://www.postgresql.org/docs/current/errcodes-appendix.html
    - https://www.postgresql.org/docs/current/client-authentication-problems.html
  classification:
    cvss-metrics: CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:N
    cwe-id: CWE-200
  metadata:
    max-request: 1
    shodan-query: port:5432 product:"PostgreSQL"
    verified: true
  tags: network,postgresql,db,detect

tcp:
  - inputs:
      - data: "000000500003000075736572006e75636c6569006461746162617365006e75636c6569006170706c69636174696f6e5f6e616d65007073716c00636c69656e745f656e636f64696e6700555446380000"
        type: hex
      - data: "7000000036534352414d2d5348412d32353600000000206e2c2c6e3d2c723d000000000000000000000000000000000000000000000000"
        type: hex

    host:
      - "{{Hostname}}"
    port: 5432
    read-size: 2048

    matchers-condition: and
    matchers:
      - type: word
        part: body
        words:
          - "C0A000"                  # Error code for unsupported frontend protocol
          - "C08P01"                  # Error code for invalide startup packet layout
          - "28000"                   # Error code for invalid_authorization_specification
          - "28P01"                   # Error code for invalid_password
          - "SCRAM-SHA-256"           # Authentication prompt
          - "pg_hba.conf"             # Client authentication config file
          - "user \"nuclei\""         # The user nuclei (sent in request) doesn't exist
          - "database \"nuclei\""     # The db nuclei (sent in request) doesn't exist"
        condition: or

      - type: word
        words:
          - "HTTP/1.1"
        negative: true
# digest: 4a0a004730450220190550562f0223183090e8ca4117ace44d725bdece7b84c58edaed8d93935aa7022100872d6d635b69589e7e99749cae0639a48551bbdee2d3d7038aa4699257a00383:922c64590222798bb761d5b6d8e72950`,
					})
					return templates
				},
				Format:      "excel",
				Count:       1,
				Timeout:     "2s",
				RateLimit:   2000,
				Concurrency: 2000,
			},
		},
	}

	id := NewID()
	defer os.RemoveAll(fmt.Sprintf("./%s", id))

	entry, err := engine.NewEntry(options, WithID(id))
	assert.NoError(err)

	err = entry.Run(context.Background())
	assert.NoError(err)

	pingSize := len(entry.Result().HostDiscoveryResult.Items)
	portSize := len(entry.Result().PortScanningResult.Items)
	jobSize := len(entry.Result().JobResults[0].Items)

	entry2, err := engine.NewEntry(options, WithID(id))
	assert.NoError(err)

	err = entry2.Run(context.Background())
	assert.NoError(err)

	pingSize += len(entry2.Result().HostDiscoveryResult.Items)
	portSize += len(entry2.Result().PortScanningResult.Items)
	jobSize += len(entry2.Result().JobResults[0].Items)

	f1, err := excelize.OpenFile(fmt.Sprintf("./%s/在线检测.xlsx", id))
	assert.NoError(err)
	content1, err := util.ReadXlsxAll(f1)
	assert.NoError(err)
	size1 := len(content1)
	f1.Close()

	f2, err := excelize.OpenFile(fmt.Sprintf("./%s/端口扫描.xlsx", id))
	assert.NoError(err)
	content2, err := util.ReadXlsxAll(f2)
	assert.NoError(err)
	size2 := len(content2)
	f2.Close()

	f3, err := excelize.OpenFile(fmt.Sprintf("./%s/资产扫描.xlsx", id))
	assert.NoError(err)
	content3, err := util.ReadXlsxAll(f3)
	assert.NoError(err)
	size3 := len(content3)
	f3.Close()

	assert.Equal(pingSize, size1-1)
	assert.Equal(portSize, size2-1)
	assert.Equal(jobSize, size3-1)
}
