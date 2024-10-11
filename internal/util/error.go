package util

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
)

type Errors struct {
	errs    []error
	options *ErrorsOptions
}

func JoinErrors(errors []error, options ...*ErrorsOptions) *Errors {
	return &Errors{
		errs:    errors,
		options: lo.IfF(len(options) > 0 && options[0] != nil, func() *ErrorsOptions { return options[0] }).Else(nil),
	}
}

func (e *Errors) Error() string {
	useIndex := false
	separator := "; "
	if e.options != nil {
		if e.options.UseIndex {
			useIndex = true
		}
		if e.options.Separator != "" {
			separator = e.options.Separator
		}
	}

	var errorMessages []string
	for i, e := range e.errs {
		if e != nil {
			errorMessages = append(errorMessages, lo.If(useIndex, fmt.Sprintf("%d.%s", i+1, e.Error())).Else(e.Error()))
		}
	}
	return strings.Join(errorMessages, separator)
}

type ErrorsOptions struct {
	UseIndex  bool
	Separator string
}
