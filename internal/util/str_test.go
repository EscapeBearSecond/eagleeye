package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomStr(t *testing.T) {
	assert := assert.New(t)

	result := RandomStr(10)
	assert.Len(result, 10)

	result = RandomStr(5)
	assert.Len(result, 5)
}
