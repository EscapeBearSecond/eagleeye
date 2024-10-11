package util

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinErrors(t *testing.T) {
	assert := assert.New(t)

	var errs []error
	errs = append(errs, errors.New("a"))
	errs = append(errs, errors.New("b"))
	errs = append(errs, errors.New("c"))
	errs = append(errs, errors.New("d"))

	err := JoinErrors(errs)
	assert.Error(err)

	errString := err.Error()
	assert.Equal(errString, "a; b; c; d")
	assert.EqualError(err, "a; b; c; d")

	err = JoinErrors(errs, &ErrorsOptions{UseIndex: true})
	assert.Error(err)

	errString = err.Error()
	assert.Equal(errString, "1.a; 2.b; 3.c; 4.d")
	assert.EqualError(err, "1.a; 2.b; 3.c; 4.d")

	err = JoinErrors(errs, &ErrorsOptions{UseIndex: true, Separator: ":: "})
	assert.Error(err)

	errString = err.Error()
	assert.Equal(errString, "1.a:: 2.b:: 3.c:: 4.d")
	assert.EqualError(err, "1.a:: 2.b:: 3.c:: 4.d")
}
