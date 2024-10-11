package sign

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSign(t *testing.T) {
	assert := assert.New(t)

	signature, err := Sign(Secret("test"), KeyValue("key", "value"))
	assert.NoError(err)
	assert.NotEmpty(signature)
	signature, err = Sign(KeyValue("key", "value"))
	assert.Error(err)
	assert.Empty(signature)
}
