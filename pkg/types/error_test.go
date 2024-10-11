package types

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorsIs(t *testing.T) {
	assert := assert.New(t)

	errAbc := fmt.Errorf("abc %w", ErrInvalidTemplates)
	errDef := fmt.Errorf("def %w", errAbc)

	assert.True(errors.Is(errDef, ErrInvalidTemplates))
}
