package flag

import (
	"testing"

	"github.com/EscapeBearSecond/eagleeye/pkg/types"
	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	assert := assert.New(t)

	jobs := []types.JobOptions{}
	jobFlag := NewJobFlag(&jobs)
	assert.NotNil(jobFlag)

	usage := jobFlag.String()
	assert.NotEmpty(usage)

	typ := jobFlag.Type()
	assert.Equal(typ, "goflag")

	err := jobFlag.Set("-m demo -a csv -c 2000 -r 2000 -e 1s -t ./templates/资产识别")
	assert.NoError(err)

	assert.Equal((*jobFlag.jobs)[0], types.JobOptions{
		Name:        "demo",
		Template:    "./templates/资产识别",
		Format:      "csv",
		Count:       1,
		Timeout:     "1s",
		RateLimit:   2000,
		Concurrency: 2000,
	})
}
