package report

import (
	"github.com/EscapeBearSecond/eagleeye/pkg/types"
)

type options struct {
	result     *types.EntryResult
	jobIndexes []int
	reporter   string
	customer   string
	directory  string
}

type Option func(*options)

func WithEntryResult(result *types.EntryResult) Option {
	return func(o *options) {
		o.result = result
	}
}

func WithJobIndexes(idx ...int) Option {
	return func(o *options) {
		o.jobIndexes = append(o.jobIndexes, idx...)
	}
}

func WithReporter(name string) Option {
	return func(o *options) {
		o.reporter = name
	}
}

func WithCustomer(customer string) Option {
	return func(o *options) {
		o.customer = customer
	}
}

func WithDirectory(directory string) Option {
	return func(o *options) {
		o.directory = directory
	}
}

type chartOptions struct {
	fileName             string
	data                 map[string]int
	fontName             string
	width                int
	xTextRotationDegrees float64
}
