package apiserver

import (
	"time"

	"github.com/EscapeBearSecond/falcon/internal/util"
	"github.com/gookit/validate"
)

func init() {
	validate.AddValidator("duration", func(v string) bool {
		_, err := time.ParseDuration(v)
		return err == nil
	})
	validate.AddValidator("ports", func(v string) bool {
		r := validate.Enum(v, []string{"http", "top100", "top1000"})
		if !r {
			_, err := util.ParsePortsList(v)
			return err == nil
		}
		return r
	})
}

type Validator struct{}

func (*Validator) Validate(obj any) error {
	v := validate.Struct(obj)
	if !v.Validate() {
		return v.Errors.ErrOrNil()
	}
	return nil
}
