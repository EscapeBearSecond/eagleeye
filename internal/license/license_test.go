package license

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/EscapeBearSecond/eagleeye/internal/util"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestGen(t *testing.T) {
	defer os.Remove("./license.json")

	assert := assert.New(t)

	err := Gen("test", "2030-01-01", "3c:22:fb:80:3b:45")
	assert.Nil(err)

	l, err := Parse("./license.json")
	assert.Nil(err)

	assert.Equal("test", l.Audience)
	assert.Equal("2030-01-01", l.ExpiresAt)

	plaintext, err := util.AESDecrypt(l.Hardware, secret)
	assert.Nil(err)

	assert.Equal("3c:22:fb:80:3b:45", plaintext)
}

func TestSign(t *testing.T) {
	assert := assert.New(t)
	l := &License{
		Id:        xid.New().String(),
		Subject:   "cursec_license",
		Issuer:    "cursec",
		IssuedAt:  time.Now().Format("2006-01-02"),
		Audience:  "test",
		Hardware:  "3c:22:fb:80:3b:45",
		ExpiresAt: "2030-01-01",
	}
	assert.NotEmpty(l.Sign())
}

func TestWatchUpdateField(t *testing.T) {
	defer os.Remove("./license.json")

	assert := assert.New(t)

	Gen("test", "2030-01-01", "3c:22:fb:80:3b:45")

	watcher, err := Watch("./license.json")
	assert.Nil(err)
	defer watcher.Stop()

	assert.Nil(Verify())

	l, err := Parse("./license.json")
	assert.Nil(err)
	l.Hardware = "123"

	f, err := os.OpenFile("./license.json", os.O_WRONLY, 0644)
	assert.Nil(err)

	encoder := json.NewEncoder(f)
	err = encoder.Encode(l)
	assert.Nil(err)
	f.Close()

	time.Sleep(6 * time.Second)

	assert.NotNil(Verify())
}

func TestWatchRemoveFile(t *testing.T) {
	defer os.Remove("./license.json")

	assert := assert.New(t)

	Gen("test", "2030-01-01", "3c:22:fb:80:3b:45")

	watcher, err := Watch("./license.json")
	assert.Nil(err)
	defer watcher.Stop()

	assert.Nil(Verify())

	err = os.Remove("./license.json")
	assert.Nil(err)

	time.Sleep(6 * time.Second)

	assert.NotNil(Verify())
}

func TestWatchChangeFormat(t *testing.T) {
	defer os.Remove("./license.json")

	assert := assert.New(t)

	Gen("test", "2030-01-01", "3c:22:fb:80:3b:45")

	watcher, err := Watch("./license.json")
	assert.Nil(err)
	defer watcher.Stop()

	assert.Nil(Verify())

	l, err := Parse("./license.json")
	assert.Nil(err)

	f, err := os.OpenFile("./license.json", os.O_WRONLY, 0644)
	assert.Nil(err)

	encoder := yaml.NewEncoder(f)
	err = encoder.Encode(l)
	assert.Nil(err)
	f.Close()

	time.Sleep(6 * time.Second)

	assert.NotNil(Verify())
}

func TestL(t *testing.T) {
	assert := assert.New(t)
	assert.Nil(L())
}
