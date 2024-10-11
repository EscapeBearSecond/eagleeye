package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptDecrypt(t *testing.T) {
	assert := assert.New(t)

	key := "12345612345612345612345612345612"

	plaintext := "hello world"
	ciphertext, err := AESEncrypt(plaintext, key)
	assert.Nil(err)

	plaintext2, err := AESDecrypt(ciphertext, key)
	assert.Nil(err)
	assert.Equal(plaintext, plaintext2)
}
