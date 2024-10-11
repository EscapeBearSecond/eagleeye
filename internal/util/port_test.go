package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePortsList(t *testing.T) {
	assert := assert.New(t)

	ports, err := ParsePortsList("1,2,3")
	assert.NoError(err)
	assert.Equal(3, len(ports))
	assert.Equal(1, ports[0])
	assert.Equal(2, ports[1])
	assert.Equal(3, ports[2])

	ports, err = ParsePortsList("1-3")
	assert.NoError(err)
	assert.Equal(3, len(ports))
	assert.Equal(1, ports[0])
	assert.Equal(2, ports[1])
	assert.Equal(3, ports[2])

	ports, err = ParsePortsList("1,2,3-5")
	assert.NoError(err)
	assert.Equal(5, len(ports))
	assert.Equal(1, ports[0])
	assert.Equal(2, ports[1])
	assert.Equal(3, ports[2])
	assert.Equal(4, ports[3])
	assert.Equal(5, ports[4])
}
