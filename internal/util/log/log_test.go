package log

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	defer os.Remove("./test.log")

	assert := assert.New(t)

	logger, err := NewLogger(WithAddSource(true), WithStdout(), WithStderr(), WithSilent(true), WithJSON(true), WithFile("./test.log"))
	assert.NoError(err)
	assert.NotNil(logger)
}

func TestMust(t *testing.T) {
	defer os.Remove("./test.log")

	assert := assert.New(t)

	logger, err := NewLogger(WithAddSource(true), WithStdout(), WithStderr(), WithSilent(true), WithJSON(true), WithFile("./test.log"))
	assert.NoError(err)
	assert.NotNil(logger)

	logger = Must(logger, err)
	assert.NotNil(logger)
}
