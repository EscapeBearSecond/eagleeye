package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPRange(t *testing.T) {
	assert := assert.New(t)

	{
		res := IsIPRange("192.168.1.255-192.168.2.2")
		assert.True(res)
	}

	{
		res := ExpandIPRange("192.168.1.255-192.168.2.2")
		assert.Equal(len(res), 4)
		assert.Contains([]string{
			"192.168.1.255",
			"192.168.2.0",
			"192.168.2.1",
			"192.168.2.2"},
			res[0])
		assert.Contains([]string{
			"192.168.1.255",
			"192.168.2.0",
			"192.168.2.1",
			"192.168.2.2"},
			res[1])
		assert.Contains([]string{
			"192.168.1.255",
			"192.168.2.0",
			"192.168.2.1",
			"192.168.2.2"},
			res[2])
		assert.Contains([]string{
			"192.168.1.255",
			"192.168.2.0",
			"192.168.2.1",
			"192.168.2.2"},
			res[3])
	}

	{
		res := IsIPRange("192.168.1.255-192.168.1.255")
		assert.False(res)
	}

	{
		res := IsIPRange("192.168.1.255")
		assert.False(res)
	}

	{
		res := IsIPRange("a.b.c.d-e.f.g.h")
		assert.False(res)
	}
}

func TestIsIP(t *testing.T) {
	assert := assert.New(t)
	res := IsIP("192.168.1.105")
	assert.True(res)

	res = IsIP("a.b.c.d")
	assert.False(res)
}

func TestIsCIDR(t *testing.T) {
	assert := assert.New(t)
	res := IsCIDR("192.168.1.0/24")
	assert.True(res)

	res = IsCIDR("192.168.1.1")
	assert.False(res)
}

func TestIsPort(t *testing.T) {
	assert := assert.New(t)
	res := IsPort("5536")
	assert.True(res)

	res = IsPort("65536")
	assert.False(res)
}

func TestIsHostPort(t *testing.T) {
	assert := assert.New(t)
	res := IsHostPort("192.168.1.1:6500")
	assert.True(res)

	res = IsHostPort("192.168.1.1")
	assert.False(res)

	res = IsHostPort("192.168.1.1:0")
	assert.False(res)

	res = IsHostPort("a.b.c.d:5600")
	assert.False(res)
}

func TestIsIPRange(t *testing.T) {
	assert := assert.New(t)
	res := IsIPRange("192.168.1.1-192.168.1.220")
	assert.True(res)

	res = IsIPRange("192.168.1.1-a.b.c.d")
	assert.False(res)

	res = IsIPRange("192.168.1.1")
	assert.False(res)

	res = IsIPRange("192.168.1.0/24")
	assert.False(res)
}

func TestToBytes(t *testing.T) {
	assert := assert.New(t)

	{
		ret := ToBytesAddr("192.168.1.1")
		assert.Equal(ret, []byte{192, 168, 1, 1})
	}
	{
		ret := ToBytesAddr("192.168.1.1:8080")
		assert.Equal(ret, []byte{192, 168, 1, 1, 31, 144})
	}
}

func TestToString(t *testing.T) {
	assert := assert.New(t)

	{
		ret := ToStringAddr([]byte{192, 168, 1, 1})
		assert.Equal(ret, "192.168.1.1")
	}
	{
		ret := ToStringAddr([]byte{192, 168, 1, 1, 31, 144})
		assert.Equal(ret, "192.168.1.1:8080")
	}
}
