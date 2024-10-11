package job

import (
	"sync"

	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/internal/global"
	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/pkg/types"
)

// resultPool 结果对象池
type resultPool struct {
	pool *sync.Pool
	once sync.Once
}

// GetResult 结果实例化
func (g *resultPool) GetResult() *types.JobResultItem {
	if global.UseSyncPool() {
		g.once.Do(func() {
			if g.pool == nil {
				g.pool = &sync.Pool{
					New: func() any {
						return types.NewJobResultItem()
					},
				}
			}
		})

		return g.pool.Get().(*types.JobResultItem)
	} else {
		return types.NewJobResultItem()
	}
}

func (g *resultPool) PutResult(result *types.JobResultItem) {
	if global.UseSyncPool() {
		g.pool.Put(result)
	}
}
