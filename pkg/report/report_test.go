package report

import (
	"os"
	"testing"
	"time"

	"github.com/EscapeBearSecond/falcon/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/wcharczuk/go-chart/v2"
)

func TestCatalog(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(catalogLtHunderd(1), "一")
	assert.Equal(catalogLtHunderd(2), "二")
	assert.Equal(catalogLtHunderd(3), "三")
	assert.Equal(catalogLtHunderd(4), "四")
	assert.Equal(catalogLtHunderd(5), "五")
	assert.Equal(catalogLtHunderd(6), "六")
	assert.Equal(catalogLtHunderd(7), "七")
	assert.Equal(catalogLtHunderd(8), "八")
	assert.Equal(catalogLtHunderd(9), "九")
	assert.Equal(catalogLtHunderd(10), "十")

	assert.Equal(catalogLtHunderd(11), "十一")
	assert.Equal(catalogLtHunderd(12), "十二")
	assert.Equal(catalogLtHunderd(13), "十三")
	assert.Equal(catalogLtHunderd(14), "十四")
	assert.Equal(catalogLtHunderd(15), "十五")
	assert.Equal(catalogLtHunderd(16), "十六")
	assert.Equal(catalogLtHunderd(17), "十七")
	assert.Equal(catalogLtHunderd(18), "十八")
	assert.Equal(catalogLtHunderd(19), "十九")
	assert.Equal(catalogLtHunderd(20), "二十")

	assert.Equal(catalogLtHunderd(21), "二十一")
	assert.Equal(catalogLtHunderd(22), "二十二")
	assert.Equal(catalogLtHunderd(23), "二十三")
	assert.Equal(catalogLtHunderd(24), "二十四")
	assert.Equal(catalogLtHunderd(25), "二十五")
	assert.Equal(catalogLtHunderd(26), "二十六")
	assert.Equal(catalogLtHunderd(27), "二十七")
	assert.Equal(catalogLtHunderd(28), "二十八")
	assert.Equal(catalogLtHunderd(29), "二十九")
	assert.Equal(catalogLtHunderd(30), "三十")
}

func TestTicks(t *testing.T) {
	assert := assert.New(t)

	assert.ElementsMatch(ticks([]chart.Value{
		{Value: 0},
		{Value: 5},
	}), []chart.Tick{
		{Value: 0, Label: "0"},
		{Value: 1, Label: "1"},
		{Value: 2, Label: "2"},
		{Value: 3, Label: "3"},
		{Value: 4, Label: "4"},
		{Value: 5, Label: "5"},
	})

	assert.ElementsMatch(ticks([]chart.Value{
		{Value: 0},
		{Value: 11},
	}), []chart.Tick{
		{Value: 0, Label: "0"},
		{Value: 2, Label: "2"},
		{Value: 4, Label: "4"},
		{Value: 6, Label: "6"},
		{Value: 8, Label: "8"},
		{Value: 10, Label: "10"},
		{Value: 12, Label: "12"},
	})

	assert.ElementsMatch(ticks([]chart.Value{
		{Value: 0},
		{Value: 23},
	}), []chart.Tick{
		{Value: 0, Label: "0"},
		{Value: 3, Label: "3"},
		{Value: 6, Label: "6"},
		{Value: 9, Label: "9"},
		{Value: 12, Label: "12"},
		{Value: 15, Label: "15"},
		{Value: 18, Label: "18"},
		{Value: 21, Label: "21"},
		{Value: 24, Label: "24"},
	})

	assert.ElementsMatch(ticks([]chart.Value{
		{Value: 0},
		{Value: 36},
	}), []chart.Tick{
		{Value: 0, Label: "0"},
		{Value: 4, Label: "4"},
		{Value: 8, Label: "8"},
		{Value: 12, Label: "12"},
		{Value: 16, Label: "16"},
		{Value: 20, Label: "20"},
		{Value: 24, Label: "24"},
		{Value: 28, Label: "28"},
		{Value: 32, Label: "32"},
		{Value: 36, Label: "36"},
	})
}

func TestGenerate(t *testing.T) {
	assert := assert.New(t)

	defer os.Remove("./report_123456.docx")

	err := Generate(WithCustomer("测试公司"),
		WithDirectory("."),
		WithJobIndexes(0),
		WithReporter("张三"),
		WithEntryResult(&types.EntryResult{
			EntryID: "123456",
			HostDiscoveryResult: &types.PingResult{
				EntryID: "123456",
				Items: []*types.PingResultItem{
					{IP: "192.168.1.2", OS: "Linux", TTL: 64, Active: true},
					{IP: "192.168.1.3", OS: "Linux", TTL: 64, Active: false},
					{IP: "192.168.1.4", OS: "Linux", TTL: 64, Active: true},
					{IP: "192.168.1.5", OS: "Linux", TTL: 64, Active: false},
					{IP: "192.168.1.6", OS: "Linux", TTL: 64, Active: true},
					{IP: "192.168.1.7", OS: "Linux", TTL: 64, Active: false},
					{IP: "192.168.1.8", OS: "Linux", TTL: 64, Active: true},
					{IP: "192.168.1.9", OS: "Linux", TTL: 64, Active: false},
					{IP: "192.168.1.10", OS: "Linux", TTL: 64, Active: true},
					{IP: "192.168.1.11", OS: "Linux", TTL: 64, Active: false},
				},
			},
			PortScanningResult: &types.PortResult{
				EntryID: "123456",
				Items: []*types.PortResultItem{
					{IP: "192.168.1.2", Port: 80},
					{IP: "192.168.1.2", Port: 5432},
					{IP: "192.168.1.2", Port: 22},
					{IP: "192.168.1.2", Port: 3389},
					{IP: "192.168.1.2", Port: 443},
					{IP: "192.168.1.2", Port: 3306},
					{IP: "192.168.1.2", Port: 1433},
					{IP: "192.168.1.2", Port: 8080},
					{IP: "192.168.1.3", Port: 80},
					{IP: "192.168.1.3", Port: 5432},
					{IP: "192.168.1.4", Port: 80},
					{IP: "192.168.1.4", Port: 5432},
					{IP: "192.168.1.5", Port: 80},
					{IP: "192.168.1.5", Port: 3389},
					{IP: "192.168.1.6", Port: 80},
					{IP: "192.168.1.7", Port: 80},
					{IP: "192.168.1.8", Port: 80},
					{IP: "192.168.1.9", Port: 80},
					{IP: "192.168.1.10", Port: 80},
					{IP: "192.168.1.10", Port: 3389},
					{IP: "192.168.1.11", Port: 80},
				},
			},
			JobResults: []*types.JobResult{
				{
					EntryID: "123456",
					Name:    "漏洞扫描",
					Kind:    "0",
					Items: []*types.JobResultItem{
						{
							EntryID:          "123456",
							TemplateID:       "123",
							TemplateName:     "sample template1",
							Severity:         "high",
							Host:             "192.168.1.2",
							Port:             "1433",
							ExtractedResults: []string{},
							Description:      "sample description1",
							Remediation:      "sample remediation1",
							Tags:             "",
						},
						{
							EntryID:          "123456",
							TemplateID:       "456",
							TemplateName:     "sample template2",
							Severity:         "high",
							Host:             "192.168.1.2",
							Port:             "3389",
							ExtractedResults: []string{},
							Description:      "sample description2",
							Remediation:      "sample remediation2",
							Tags:             "",
						},
						{
							EntryID:          "123456",
							TemplateID:       "456",
							TemplateName:     "sample template2",
							Severity:         "high",
							Host:             "192.168.1.5",
							Port:             "3389",
							ExtractedResults: []string{},
							Description:      "sample description2",
							Remediation:      "sample remediation2",
							Tags:             "",
						},
						{
							EntryID:          "123456",
							TemplateID:       "456",
							TemplateName:     "sample template2",
							Severity:         "high",
							Host:             "192.168.1.10",
							Port:             "3389",
							ExtractedResults: []string{},
							Description:      "sample description2",
							Remediation:      "sample remediation2",
							Tags:             "",
						},
					},
				},
			},
			Targets: []string{
				"192.168.1.0/24",
			},
			ExcludeTargets: []string{},
			StartTime:      time.Now().Add(-2 * time.Minute),
			EndTime:        time.Now().Add(-2 * time.Minute),
		}))

	assert.Nil(err)
}

func TestGenerateWithSomeEmpty(t *testing.T) {
	assert := assert.New(t)

	defer os.Remove("./report_123456.docx")

	err := Generate(
		WithDirectory("."),
		WithJobIndexes(0),
		WithEntryResult(&types.EntryResult{
			EntryID: "123456",
			HostDiscoveryResult: &types.PingResult{
				EntryID: "123456",
				Items: []*types.PingResultItem{
					{IP: "192.168.1.2", OS: "Linux", TTL: 64, Active: true},
					{IP: "192.168.1.3", OS: "Linux", TTL: 64, Active: false},
					{IP: "192.168.1.4", OS: "Linux", TTL: 64, Active: true},
					{IP: "192.168.1.5", OS: "Linux", TTL: 64, Active: false},
					{IP: "192.168.1.6", OS: "Linux", TTL: 64, Active: true},
					{IP: "192.168.1.7", OS: "Linux", TTL: 64, Active: false},
					{IP: "192.168.1.8", OS: "Linux", TTL: 64, Active: true},
					{IP: "192.168.1.9", OS: "Linux", TTL: 64, Active: false},
					{IP: "192.168.1.10", OS: "Linux", TTL: 64, Active: true},
					{IP: "192.168.1.11", OS: "Linux", TTL: 64, Active: false},
				},
			},
			PortScanningResult: &types.PortResult{
				EntryID: "123456",
				Items: []*types.PortResultItem{
					{IP: "192.168.1.2", Port: 80},
					{IP: "192.168.1.2", Port: 5432},
					{IP: "192.168.1.2", Port: 22},
					{IP: "192.168.1.2", Port: 3389},
					{IP: "192.168.1.2", Port: 443},
					{IP: "192.168.1.2", Port: 3306},
					{IP: "192.168.1.2", Port: 1433},
					{IP: "192.168.1.2", Port: 8080},
					{IP: "192.168.1.3", Port: 80},
					{IP: "192.168.1.3", Port: 5432},
					{IP: "192.168.1.4", Port: 80},
					{IP: "192.168.1.4", Port: 5432},
					{IP: "192.168.1.5", Port: 80},
					{IP: "192.168.1.5", Port: 3389},
					{IP: "192.168.1.6", Port: 80},
					{IP: "192.168.1.7", Port: 80},
					{IP: "192.168.1.8", Port: 80},
					{IP: "192.168.1.9", Port: 80},
					{IP: "192.168.1.10", Port: 80},
					{IP: "192.168.1.10", Port: 3389},
					{IP: "192.168.1.11", Port: 80},
				},
			},
			JobResults: []*types.JobResult{
				{
					EntryID: "123456",
					Name:    "漏洞扫描",
					Kind:    "0",
					Items: []*types.JobResultItem{
						{
							EntryID:          "123456",
							TemplateID:       "123",
							TemplateName:     "sample template1",
							Severity:         "high",
							Host:             "192.168.1.2",
							Port:             "1433",
							ExtractedResults: []string{},
							Description:      "sample description1",
							Remediation:      "sample remediation1",
							Tags:             "",
						},
						{
							EntryID:          "123456",
							TemplateID:       "456",
							TemplateName:     "sample template2",
							Severity:         "high",
							Host:             "192.168.1.2",
							Port:             "3389",
							ExtractedResults: []string{},
							Description:      "sample description2",
							Remediation:      "sample remediation2",
							Tags:             "",
						},
						{
							EntryID:          "123456",
							TemplateID:       "456",
							TemplateName:     "sample template2",
							Severity:         "high",
							Host:             "192.168.1.5",
							Port:             "3389",
							ExtractedResults: []string{},
							Description:      "sample description2",
							Remediation:      "sample remediation2",
							Tags:             "",
						},
						{
							EntryID:          "123456",
							TemplateID:       "456",
							TemplateName:     "sample template2",
							Severity:         "high",
							Host:             "192.168.1.10",
							Port:             "3389",
							ExtractedResults: []string{},
							Description:      "sample description2",
							Remediation:      "sample remediation2",
							Tags:             "",
						},
					},
				},
			},
		}))

	assert.Nil(err)
}
