package target

import (
	"os"
	"path/filepath"
	"testing"

	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestProcessSync(t *testing.T) {
	assert := assert.New(t)

	{
		targets, err := ProcessSync([]string{
			"192.168.1.1-192.168.1.5",
			"192.168.2.0/24",
			"192.168.3.123",
			"192.168.4.234:8080",
		})
		assert.NoError(err)
		assert.Len(targets, 5+256+1+1)

		ips := util.ExpandIPRange("192.168.1.1-192.168.1.5")
		for _, ip := range ips {
			assert.Contains(targets, ip)
		}
		ips = util.ExpandCIDR("192.168.2.0/24")
		for _, ip := range ips {
			assert.Contains(targets, ip)
		}
		assert.Contains(targets, "192.168.3.123")
		assert.Contains(targets, "192.168.4.234:8080")
	}

	{
		tempDir, err := os.MkdirTemp("", "test")
		assert.NoError(err)
		defer os.RemoveAll(tempDir)

		path := filepath.Join(tempDir, "test.txt")
		err = os.MkdirAll(filepath.Dir(path), 0755)
		assert.NoError(err)
		err = os.WriteFile(path, []byte(`192.168.1.1-192.168.1.5
192.168.2.0/24
192.168.3.123
192.168.4.234:8080`), 0644)
		assert.NoError(err)

		targets, err := ProcessSync([]string{
			path,
		})
		assert.NoError(err)
		assert.Len(targets, 5+256+1+1)

		ips := util.ExpandIPRange("192.168.1.1-192.168.1.5")
		for _, ip := range ips {
			assert.Contains(targets, ip)
		}
		ips = util.ExpandCIDR("192.168.2.0/24")
		for _, ip := range ips {
			assert.Contains(targets, ip)
		}
		assert.Contains(targets, "192.168.3.123")
		assert.Contains(targets, "192.168.4.234:8080")
	}
}

func TestProcessAsync(t *testing.T) {
	assert := assert.New(t)

	{
		targets, err := ProcessAsync([]string{
			"192.168.1.1-192.168.1.5",
			"192.168.2.0/24",
			"192.168.3.123",
			"192.168.4.234:8080",
		})
		assert.NoError(err)
		assert.Len(targets, 5+256+1+1)

		ips := util.ExpandIPRange("192.168.1.1-192.168.1.5")
		for _, ip := range ips {
			assert.Contains(targets, ip)
		}
		ips = util.ExpandCIDR("192.168.2.0/24")
		for _, ip := range ips {
			assert.Contains(targets, ip)
		}
		assert.Contains(targets, "192.168.3.123")
		assert.Contains(targets, "192.168.4.234:8080")
	}

	{
		tempDir, err := os.MkdirTemp("", "test")
		assert.NoError(err)
		defer os.RemoveAll(tempDir)

		path := filepath.Join(tempDir, "test.txt")
		err = os.MkdirAll(filepath.Dir(path), 0755)
		assert.NoError(err)
		err = os.WriteFile(path, []byte(`192.168.1.1-192.168.1.5
192.168.2.0/24
192.168.3.123
192.168.4.234:8080`), 0644)
		assert.NoError(err)

		targets, err := ProcessAsync([]string{
			path,
		})
		assert.NoError(err)
		assert.Len(targets, 5+256+1+1)

		ips := util.ExpandIPRange("192.168.1.1-192.168.1.5")
		for _, ip := range ips {
			assert.Contains(targets, ip)
		}
		ips = util.ExpandCIDR("192.168.2.0/24")
		for _, ip := range ips {
			assert.Contains(targets, ip)
		}
		assert.Contains(targets, "192.168.3.123")
		assert.Contains(targets, "192.168.4.234:8080")
	}
}

func TestProcessAsyncExlude(t *testing.T) {
	assert := assert.New(t)
	targets, err := ProcessAsync([]string{
		"192.168.1.1-192.168.1.5",
		"192.168.2.0/24",
		"192.168.3.123",
		"192.168.4.234:8080",
	}, []string{
		"192.168.2.0/24",
		"192.168.3.123",
		"192.168.4.234:8080",
	}...)
	assert.NoError(err)
	assert.Len(targets, 5)
	ips := util.ExpandIPRange("192.168.1.1-192.168.1.5")
	for _, ip := range ips {
		assert.Contains(targets, ip)
	}
	ips = util.ExpandCIDR("192.168.2.0/24")
	for _, ip := range ips {
		assert.NotContains(targets, ip)
	}
	assert.NotContains(targets, "192.168.3.123")
	assert.NotContains(targets, "192.168.4.234:8080")
}

func BenchmarkProcessSync(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ProcessSync([]string{
			"192.160.0.0/16",
			"192.161.0.0/16",
			"192.162.0.0/16",
			"192.163.0.0/16",
			"192.164.0.0/16",
			"192.165.0.0/16",
			"192.166.0.0/16",
			"192.167.0.0/16",
			"192.168.0.0/16",
			"192.169.0.0/16",
		})
	}
}

func BenchmarkProcessAsync(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ProcessAsync([]string{
			"192.160.0.0/16",
			"192.161.0.0/16",
			"192.162.0.0/16",
			"192.163.0.0/16",
			"192.164.0.0/16",
			"192.165.0.0/16",
			"192.166.0.0/16",
			"192.167.0.0/16",
			"192.168.0.0/16",
			"192.169.0.0/16",
		})
	}
}

func BenchmarkProcessAsyncExclude(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ProcessAsync([]string{
			"192.160.0.0/16",
			"192.161.0.0/16",
			"192.162.0.0/16",
			"192.163.0.0/16",
			"192.164.0.0/16",
			"192.165.0.0/16",
			"192.166.0.0/16",
			"192.167.0.0/16",
			"192.168.0.0/16",
			"192.169.0.0/16",
		}, []string{
			"192.160.0.0/16",
			"192.161.0.0/16",
		}...)
	}
}

func TestSplitN(t *testing.T) {
	assert := assert.New(t)

	results, err := SplitN([]string{"192.168.1.0/24"}, 3)
	assert.NoError(err)
	assert.Len(results, 3)
	assert.ElementsMatch(results, [][]string{{"192.168.1.0-192.168.1.84"}, {"192.168.1.85-192.168.1.169"}, {"192.168.1.170-192.168.1.255"}})

	results, err = SplitN([]string{"192.168.1.0/24"}, 4)
	assert.NoError(err)
	assert.Len(results, 4)
	assert.ElementsMatch(results, [][]string{{"192.168.1.0-192.168.1.63"}, {"192.168.1.64-192.168.1.127"}, {"192.168.1.128-192.168.1.191"}, {"192.168.1.192-192.168.1.255"}})

	results, err = SplitN([]string{"192.168.1.0/24"}, 5)
	assert.NoError(err)
	assert.Len(results, 5)
	assert.ElementsMatch(results, [][]string{{"192.168.1.0-192.168.1.50"}, {"192.168.1.51-192.168.1.101"}, {"192.168.1.102-192.168.1.152"}, {"192.168.1.153-192.168.1.203"}, {"192.168.1.204-192.168.1.255"}})

	results, err = SplitN([]string{"192.168.1.0-192.168.1.255"}, 3)
	assert.NoError(err)
	assert.Len(results, 3)
	assert.ElementsMatch(results, [][]string{{"192.168.1.0-192.168.1.84"}, {"192.168.1.85-192.168.1.169"}, {"192.168.1.170-192.168.1.255"}})

	results, err = SplitN([]string{"192.168.1.0-192.168.1.255"}, 4)
	assert.NoError(err)
	assert.Len(results, 4)
	assert.ElementsMatch(results, [][]string{{"192.168.1.0-192.168.1.63"}, {"192.168.1.64-192.168.1.127"}, {"192.168.1.128-192.168.1.191"}, {"192.168.1.192-192.168.1.255"}})

	results, err = SplitN([]string{"192.168.1.0-192.168.1.255"}, 5)
	assert.NoError(err)
	assert.Len(results, 5)
	assert.ElementsMatch(results, [][]string{{"192.168.1.0-192.168.1.50"}, {"192.168.1.51-192.168.1.101"}, {"192.168.1.102-192.168.1.152"}, {"192.168.1.153-192.168.1.203"}, {"192.168.1.204-192.168.1.255"}})

	results, err = SplitN([]string{"192.168.1.0-192.168.1.255", "192.168.2.0/24"}, 3)
	assert.NoError(err)
	assert.Len(results, 3)
	assert.ElementsMatch(results, [][]string{{"192.168.1.0-192.168.1.84", "192.168.2.0-192.168.2.84"}, {"192.168.1.85-192.168.1.169", "192.168.2.85-192.168.2.169"}, {"192.168.1.170-192.168.1.255", "192.168.2.170-192.168.2.255"}})

	results, err = SplitN([]string{"192.168.1.0-192.168.1.255", "192.168.2.0/24"}, 4)
	assert.NoError(err)
	assert.Len(results, 4)
	assert.ElementsMatch(results, [][]string{{"192.168.1.0-192.168.1.63", "192.168.2.0-192.168.2.63"}, {"192.168.1.64-192.168.1.127", "192.168.2.64-192.168.2.127"}, {"192.168.1.128-192.168.1.191", "192.168.2.128-192.168.2.191"}, {"192.168.1.192-192.168.1.255", "192.168.2.192-192.168.2.255"}})

	results, err = SplitN([]string{"192.168.1.0-192.168.1.255", "192.168.2.0/24"}, 5)
	assert.NoError(err)
	assert.Len(results, 5)
	assert.ElementsMatch(results, [][]string{{"192.168.1.0-192.168.1.50", "192.168.2.0-192.168.2.50"}, {"192.168.1.51-192.168.1.101", "192.168.2.51-192.168.2.101"}, {"192.168.1.102-192.168.1.152", "192.168.2.102-192.168.2.152"}, {"192.168.1.153-192.168.1.203", "192.168.2.153-192.168.2.203"}, {"192.168.1.204-192.168.1.255", "192.168.2.204-192.168.2.255"}})

	results, err = SplitN([]string{"192.168.1.123-192.168.1.255"}, 6)
	assert.NoError(err)
	assert.Len(results, 6)
	assert.ElementsMatch(results, [][]string{{"192.168.1.123-192.168.1.144"}, {"192.168.1.145-192.168.1.166"}, {"192.168.1.167-192.168.1.188"}, {"192.168.1.189-192.168.1.210"}, {"192.168.1.211-192.168.1.232"}, {"192.168.1.233-192.168.1.255"}})

	results, err = SplitN([]string{"192.168.1.123-192.168.1.255"}, 7)
	assert.NoError(err)
	assert.Len(results, 7)
	assert.ElementsMatch(results, [][]string{{"192.168.1.123-192.168.1.141"}, {"192.168.1.142-192.168.1.160"}, {"192.168.1.161-192.168.1.179"}, {"192.168.1.180-192.168.1.198"}, {"192.168.1.199-192.168.1.217"}, {"192.168.1.218-192.168.1.236"}, {"192.168.1.237-192.168.1.255"}})

	results, err = SplitN([]string{"192.168.1.0"}, 2)
	assert.NoError(err)
	assert.Len(results, 2)
	assert.ElementsMatch(results, [][]string{{"192.168.1.0"}, {}})
}

func TestSplitBySize(t *testing.T) {
	assert := assert.New(t)

	results, err := SplitBySize([]string{
		"192.160.0.0/16",
		"192.161.0.0/16",
		"192.162.0.0/16",
	}, 65536)

	assert.NoError(err)
	assert.Equal(3, len(results))

	results, err = SplitBySize([]string{
		"192.160.0.0/16",
	}, 65536)

	assert.NoError(err)
	assert.Equal(1, len(results))
}
