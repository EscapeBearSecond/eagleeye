package license

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/EscapeBearSecond/falcon/internal/util"
	"github.com/EscapeBearSecond/falcon/internal/util/sign"
	"github.com/jinzhu/copier"
	"github.com/rs/xid"
	"github.com/samber/lo"
)

var (
	ErrNotExists        = fmt.Errorf("license does not exist")
	ErrInvalidSignature = fmt.Errorf("invalid signature")
	ErrExpired          = fmt.Errorf("license expired")
	ErrHardwareMismatch = fmt.Errorf("hardware mismatch")
)

var (
	// secret 用于hmac.sha256签名及aes.gcm加密
	secret = "3Y7hjsJk9wMq0Tg2LZi5N8VqW4RmX1dF"

	// license 用于存储license，由watch管理
	license *License
	rwm     sync.RWMutex
)

// Verify 验证证书
func Verify() error {
	return license.Verify()
}

// L 获取当前license（拷贝副本）
func L() *License {
	if license == nil {
		return nil
	}

	rwm.RLock()
	defer rwm.RUnlock()
	var l License
	copier.Copy(&l, license)
	return &l
}

func UseSecret(str string) {
	secret = str
}

// License 证书对象
type License struct {
	Id        string    `json:"id"`
	Subject   string    `json:"subject"`
	Issuer    string    `json:"issuer"`
	IssuedAt  string    `json:"issued_at"`
	Audience  string    `json:"audience"`
	Hardware  string    `json:"hardware"`
	ExpiresAt string    `json:"expires_at"`
	Signature string    `json:"signature"`
	ModTime   time.Time `json:"-"`
}

// Parse 解析license
func Parse(path string) (*License, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open license file failed: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("get license info failed: %w", err)
	}

	var license *License
	license, err = ParseFromReader(f)
	if err != nil {
		return nil, fmt.Errorf("parse license file failed: %w", err)
	}

	license.ModTime = info.ModTime()
	return license, nil
}

func ParseFromReader(reader io.Reader) (*License, error) {
	var license License
	err := json.NewDecoder(reader).Decode(&license)
	if err != nil {
		return nil, fmt.Errorf("invalid license file format: %w", err)
	}
	return &license, nil
}

// Sign 签名证书
func (l *License) Sign() string {
	rwm.RLock()
	defer rwm.RUnlock()
	return lo.Must(sign.Sign(
		sign.Secret(secret),
		sign.KeyValue("id", l.Id),
		sign.KeyValue("subject", l.Subject),
		sign.KeyValue("issuer", l.Issuer),
		sign.KeyValue("issued_at", l.IssuedAt),
		sign.KeyValue("audience", l.Audience),
		sign.KeyValue("hardware", l.Hardware),
		sign.KeyValue("expires_at", l.ExpiresAt),
	))
}

// Verify 验证证书
func (l *License) Verify() error {
	rwm.RLock()
	defer rwm.RUnlock()

	if l == nil {
		return ErrNotExists
	}

	if l.Signature != l.Sign() {
		return ErrInvalidSignature
	}

	expiresAt, err := time.ParseInLocation("2006-01-02", l.ExpiresAt, time.Local)
	// 签名匹配，信息未被篡改，ExpiresAt解析不应该失败，如果失败，默认成功
	if err != nil {
		return nil
	}

	now := time.Now()
	if expiresAt.Before(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)) {
		return ErrExpired
	}

	plaintext, err := util.AESDecrypt(l.Hardware, secret)
	// 签名匹配，信息未被篡改，hardware解密不应该失败，如果失败，默认成功
	if err != nil {
		return nil
	}

	// 获取设备网卡信息
	interfaces, err := net.Interfaces()
	// 如果设备网卡信息获取失败，默认成功
	if err != nil {
		return nil
	}

	hardwares := make([]string, 0)
	for _, interfaceInfo := range interfaces {
		if interfaceInfo.HardwareAddr == nil {
			continue
		}
		hardwares = append(hardwares, strings.ToLower(interfaceInfo.HardwareAddr.String()))
	}
	hardwares = lo.Uniq(hardwares)

	licenseHardwares := strings.Split(plaintext, ",")
	for _, hardware := range licenseHardwares {
		if !lo.Contains(hardwares, hardware) {
			return ErrHardwareMismatch
		}
	}
	return nil
}

// Gen 生成license
func Gen(audience string, expiresAt string, hardwares ...string) error {
	if len(hardwares) == 0 || hardwares[0] == "" {
		return fmt.Errorf("hardwares is required")
	}

	_, err := time.ParseInLocation("2006-01-02", expiresAt, time.Local)
	if err != nil {
		return fmt.Errorf("invalid expires_at format: %w", err)
	}

	now := time.Now()

	// 为了可以生成过期license，所以不做验证
	// if exp.Before(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)) {
	// 	return fmt.Errorf("expires_at must be in the future")
	// }

	licenseHardwares := make([]string, 0, len(hardwares))
	for _, hardware := range hardwares {
		hardwareAddr, err := net.ParseMAC(hardware)
		if err != nil {
			return fmt.Errorf("invalid hardware format: %w", err)
		}
		licenseHardwares = append(licenseHardwares, strings.ToLower(hardwareAddr.String()))
	}
	hardwareStr := strings.Join(licenseHardwares, ",")
	ciphertext, err := util.AESEncrypt(hardwareStr, secret)
	if err != nil {
		return fmt.Errorf("encrypt hardware failed: %w", err)
	}

	license := &License{
		Id:        xid.New().String(),
		Subject:   "cursec_license",
		Issuer:    "cursec",
		IssuedAt:  now.Format("2006-01-02"),
		Audience:  audience,
		Hardware:  ciphertext,
		ExpiresAt: expiresAt,
	}
	license.Signature = license.Sign()

	jsonBytes, err := json.MarshalIndent(license, "", "\t")
	if err != nil {
		return fmt.Errorf("marshal license failed: %w", err)
	}

	err = os.WriteFile("license.json", jsonBytes, 0644)
	if err != nil {
		return fmt.Errorf("write license failed: %w", err)
	}

	return nil
}

// =================================Watcher (用于监听license变化)========================================

// Watcher 监听license变化
type Watcher struct {
	cancel context.CancelFunc
}

// Watch 监听license并返回Watcher
func Watch(path string) (*Watcher, error) {
	l, err := Parse(path)
	if err != nil {
		return nil, fmt.Errorf("parse license failed: %w", err)
	}
	license = l

	c, cancel := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-c.Done():
				return
			case <-ticker.C:
				stat, err := os.Stat(path)
				if err != nil {
					license = nil
					continue
				}

				modTime := stat.ModTime()
				if license == nil || !modTime.Equal(license.ModTime) {
					l, err := Parse(path)
					if err != nil {
						license = nil
						continue
					}
					rwm.Lock()
					copier.Copy(license, l)
					rwm.Unlock()
				}
			}
		}
	}()
	return &Watcher{cancel: cancel}, nil
}

// Stop 停止监听
func (w *Watcher) Stop() {
	w.cancel()
}
