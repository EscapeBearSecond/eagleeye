package stage

import (
	"testing"

	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestStage(t *testing.T) {
	assert := assert.New(t)

	manager := NewManager()
	manager.Put(types.StagePreExecute, 0)
	stage := manager.Get()

	assert.Equal(types.StagePreExecute, stage.Name)
	assert.Equal(float64(0), stage.Percent)
}
