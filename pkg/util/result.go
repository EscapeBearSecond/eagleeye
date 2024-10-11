package util

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/EscapeBearSecond/eagleeye/internal/util"
	"github.com/EscapeBearSecond/eagleeye/pkg/types"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"github.com/xuri/excelize/v2"
)

func ReloadResult(readers ...*types.ResultReader) (*types.EntryResult, error) {
	if len(readers) == 0 {
		return nil, errors.New("no result readers")
	}

	jobSize := lo.CountBy(readers, func(reader *types.ResultReader) bool {
		return reader.Stage == types.StageJob
	})

	result := &types.EntryResult{
		JobResults: make([]*types.JobResult, 0, jobSize),
	}

	for i, reader := range readers {
		if reader == nil {
			return nil, fmt.Errorf("result reader [%d] is nil", i)
		}

		switch reader.Stage {
		case types.StageHostDiscovery:
			pr, err := reloadHostDiscovery(reader)
			if err != nil {
				return nil, fmt.Errorf("reload host discovery result failed: %w", err)
			}
			result.HostDiscoveryResult = pr
		case types.StagePortScanning:
			pr, err := reloadPortScanning(reader)
			if err != nil {
				return nil, fmt.Errorf("reload port scanning result failed: %w", err)
			}
			result.PortScanningResult = pr
		case types.StageJob:
			jr, err := reloadJob(reader)
			if err != nil {
				return nil, fmt.Errorf("reload job result failed: %w", err)
			}
			result.JobResults = append(result.JobResults, jr)
		}
	}

	return result, nil
}

func reloadHostDiscovery(reader *types.ResultReader) (*types.PingResult, error) {
	var contents [][]string
	var err error

	switch reader.Format {
	case "csv":
		csvReader := csv.NewReader(reader.Reader)
		contents, err = csvReader.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("read csv failed: %w", err)
		}
	case "excel":
		excelReader, err := excelize.OpenReader(reader.Reader)
		if err != nil {
			return nil, fmt.Errorf("open excel reader failed: %w", err)
		}
		defer excelReader.Close()

		contents, err = util.ReadXlsxAll(excelReader)
		if err != nil {
			return nil, fmt.Errorf("read excel failed: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported format: %s", reader.Format)
	}

	pr := &types.PingResult{
		Items: make([]*types.PingResultItem, 0, len(contents)-1),
	}
	for i, line := range contents {
		if i == 0 {
			continue
		}
		pr.Items = append(pr.Items, &types.PingResultItem{
			IP:     line[0],
			Active: line[1] == "是",
			OS:     lo.IfF(line[1] == "是", func() string { return line[2] }).Else(""),
			TTL:    lo.IfF(line[1] == "是", func() int { return cast.ToInt(line[3]) }).Else(0),
		})
	}

	return pr, nil
}

func reloadPortScanning(reader *types.ResultReader) (*types.PortResult, error) {

	var contents [][]string
	var err error

	switch reader.Format {
	case "csv":
		csvReader := csv.NewReader(reader.Reader)
		contents, err = csvReader.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("read csv failed: %w", err)
		}
	case "excel":
		excelReader, err := excelize.OpenReader(reader.Reader)
		if err != nil {
			return nil, fmt.Errorf("open excel reader failed: %w", err)
		}
		defer excelReader.Close()

		contents, err = util.ReadXlsxAll(excelReader)
		if err != nil {
			return nil, fmt.Errorf("read excel failed: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported format: %s", reader.Format)
	}

	pr := &types.PortResult{
		Items: make([]*types.PortResultItem, 0, len(contents)-1),
	}
	for i, line := range contents {
		if i == 0 {
			continue
		}
		pr.Items = append(pr.Items, &types.PortResultItem{
			IP:       line[0],
			Port:     cast.ToInt(line[1]),
			HostPort: net.JoinHostPort(line[0], line[1]),
		})
	}

	return pr, nil
}

func reloadJob(reader *types.ResultReader) (*types.JobResult, error) {

	var contents [][]string
	var err error

	switch reader.Format {
	case "csv":
		csvReader := csv.NewReader(reader.Reader)
		contents, err = csvReader.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("read csv failed: %w", err)
		}
	case "excel":
		excelReader, err := excelize.OpenReader(reader.Reader)
		if err != nil {
			return nil, fmt.Errorf("open excel reader failed: %w", err)
		}
		defer excelReader.Close()

		contents, err = util.ReadXlsxAll(excelReader)
		if err != nil {
			return nil, fmt.Errorf("read excel failed: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported format: %s", reader.Format)
	}

	jr := &types.JobResult{
		Items: make([]*types.JobResultItem, 0, len(contents)-1),
	}

	for i, c := range contents {
		if i == 0 {
			continue
		}

		jri := &types.JobResultItem{}
		for j := range c {
			switch j {
			case 0:
				jri.TemplateID = c[j]
			case 1:
				jri.TemplateName = c[j]
			case 2:
				jri.Type = c[j]
			case 3:
				jri.Severity = c[j]
			case 4:
				jri.Tags = c[j]
			case 5:
				jri.Host = c[j]
			case 6:
				jri.Port = c[j]
			case 7:
				jri.Scheme = c[j]
			case 8:
				jri.URL = c[j]
			case 9:
				jri.Path = c[j]
			case 10:
				jri.Matched = c[j]
			case 11:
				jri.ExtractedResults = strings.Split(c[j], "|")
			case 12:
				jri.Description = c[j]
			case 13:
				jri.Remediation = c[j]
			}
		}

		jr.Items = append(jr.Items, jri)
	}

	return jr, nil
}
