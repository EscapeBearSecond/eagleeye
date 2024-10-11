package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"
)

func AESEncrypt(plaintext string, key string) (ciphertext string, err error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(gcm.Seal(nonce, nonce, []byte(plaintext), nil)), nil
}

func AESDecrypt(ciphertext string, key string) (plaintext string, err error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	cipherbytes, err := hex.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	plainBytes, err := gcm.Open(nil,
		cipherbytes[:gcm.NonceSize()],
		cipherbytes[gcm.NonceSize():],
		nil,
	)
	if err != nil {
		return "", err
	}
	return string(plainBytes), nil
}
