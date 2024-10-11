package job

import (
	"context"
	"errors"

	"github.com/EscapeBearSecond/eagleeye/pkg/types"
)

var (
	ErrNotSupportExporter = errors.New("unsupport exporter mode")
	ErrEmptyOutputName    = errors.New("empty output name")
)

const (
	headerTemplateID       string = "编号"
	headerTemplateName     string = "名称"
	headerTemplateType     string = "检测类型"
	headerTemplateSeverity string = "等级"
	headerTemplateTags     string = "标签"
	headerHost             string = "主机"
	headerPort             string = "端口"
	headerScheme           string = "协议"
	headerURL              string = "URL"
	headerPath             string = "路径"
	headerMatched          string = "匹配值"
	headerExtractedResults string = "提取结果"
	headerDescription      string = "描述"
	headerRemediation      string = "修复建议"
	headerHitCredential    string = "命中口令"
)

type exportFormat string

const (
	csvMode     exportFormat = "csv"
	consoleMode exportFormat = "console"
	excelMode   exportFormat = "excel"
)

// exporter 格式化输出接口
type exporter interface {
	Export(context.Context, *types.JobResultItem) error
	Stop() error
	GetResult() *types.JobResultItem
}

func newExporter(format exportFormat, args ...string) (exporter, error) {
	switch format {
	case csvMode, excelMode:
		if len(args) == 0 || args[0] == "" {
			return nil, ErrEmptyOutputName
		}
		exporter, err := newFileExporter(format, args[0])
		if err != nil {
			return nil, err
		}
		return exporter, nil
	case consoleMode:
		exporter := newConsoleExporter()
		return exporter, nil
	default:
		return nil, ErrNotSupportExporter
	}
}
