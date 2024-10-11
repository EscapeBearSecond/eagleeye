package mmh3

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/spaolacci/murmur3"
)

func processFavicon(str string) ([]byte, error) {
	if u, err := url.Parse(str); err != nil || u.Scheme == "" || u.Host == "" {
		faviconBytes, err := os.ReadFile(str)
		if err != nil {
			return nil, fmt.Errorf("failed to read favicon: %w", err)
		}
		return faviconBytes, nil
	}

	//创建不验证证书的客户端
	client := http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS10,
				Renegotiation:      tls.RenegotiateOnceAsClient,
				CipherSuites: []uint16{
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256, tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384, tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,

					tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA, tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA, tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,

					tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_RSA_WITH_AES_256_GCM_SHA384,

					tls.TLS_RSA_WITH_AES_128_CBC_SHA,
					tls.TLS_RSA_WITH_AES_256_CBC_SHA,

					tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
					tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,

					tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256, tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
					tls.TLS_RSA_WITH_AES_128_CBC_SHA256,

					tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA, tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
					tls.TLS_RSA_WITH_RC4_128_SHA,
				},
			},
		},
	}

	// 下载favicon
	resp, err := client.Get(str)
	if err != nil {
		return nil, fmt.Errorf("failed to download favicon: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download favicon: status code %d", resp.StatusCode)
	}

	// 读取favicon内容
	faviconBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read favicon: %w", err)
	}

	return faviconBytes, nil
}

func insertInto(s string, interval int, sep rune) string {
	var buffer bytes.Buffer
	before := interval - 1
	last := len(s) - 1
	for i, char := range s {
		buffer.WriteRune(char)
		if i%interval == before && i != last {
			buffer.WriteRune(sep)
		}
	}
	buffer.WriteRune(sep)
	return buffer.String()
}

func calculateHash(faviconBase64 string) int32 {
	hasher := murmur3.New32WithSeed(0)
	hasher.Write([]byte(faviconBase64))

	return int32(hasher.Sum32())
}

type Base64Func func(data []byte) string

func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func Base64PyEncode(data []byte) string {
	stdBase64 := base64.StdEncoding.EncodeToString(data)
	return insertInto(stdBase64, 76, '\n')
}

func Hash(str string, base64Func Base64Func) (int32, error) {
	iconBytes, err := processFavicon(str)
	if err != nil {
		return 0, err
	}
	return calculateHash(base64Func(iconBytes)), nil
}
