package flag

import (
	"bytes"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/pkg/types"
)

// JobFlag job的命令行嵌入flag
type JobFlag struct {
	jobs    *[]types.JobOptions
	flagSet *flag.FlagSet
	buf     *bytes.Buffer
}

func NewJobFlag(jobs *[]types.JobOptions) *JobFlag {
	defaultOptions := types.DefaultOptions(1)

	// job参数的内部flag，goflag形式，用于支持命令行模式
	flagSet := flag.NewFlagSet("job", flag.ContinueOnError)
	flagSet.String("m", "", "任务名称")
	flagSet.String("t", "", "任务模版（目录/文件）")
	flagSet.Bool("b", defaultOptions.Jobs[0].Headless, "开启headless模式")
	flagSet.String("a", defaultOptions.Jobs[0].Format, "任务输出格式")
	flagSet.Int("n", defaultOptions.Jobs[0].Count, "任务执行轮次")
	flagSet.String("e", defaultOptions.Jobs[0].Timeout, "任务超时时间")
	flagSet.Int("r", defaultOptions.Jobs[0].RateLimit, "任务执行频率")
	flagSet.Int("c", defaultOptions.Jobs[0].Concurrency, "任务并发数")

	// 设置flag set输出到buf中
	buf := &bytes.Buffer{}
	flagSet.SetOutput(buf)

	return &JobFlag{
		jobs:    jobs,
		flagSet: flagSet,
		buf:     buf,
	}
}

func (f *JobFlag) String() string {
	f.flagSet.Usage()
	return f.buf.String()
}

func (f *JobFlag) Set(v string) error {
	args := []string{}
	// 格式化入参
	for _, s := range strings.Split(v, " ") {
		s = strings.TrimSpace(s)
		if len(s) != 0 {
			args = append(args, strings.Split(s, " ")...)
		}
	}

	if err := f.flagSet.Parse(args); err != nil {
		return fmt.Errorf("invalid job args: %w", err)
	}

	var j types.JobOptions
	// 遍历入参并赋值
	f.flagSet.VisitAll(func(f *flag.Flag) {
		switch f.Name {
		case "m":
			j.Name = f.Value.String()
		case "t":
			j.Template = f.Value.String()
		case "a":
			j.Format = f.Value.String()
		case "e":
			j.Timeout = f.Value.String()
		case "b":
			v, err := strconv.ParseBool(f.Value.String())
			if err == nil {
				j.Headless = v
			}
		case "n":
			v, err := strconv.Atoi(f.Value.String())
			if err == nil {
				j.Count = v
			}
		case "r":
			v, err := strconv.Atoi(f.Value.String())
			if err == nil {
				j.RateLimit = v
			}
		case "c":
			v, err := strconv.Atoi(f.Value.String())
			if err == nil {
				j.Concurrency = v
			}
		}
	})
	*f.jobs = append(*f.jobs, j)
	return nil
}

func (f *JobFlag) Type() string {
	return "goflag"
}
