package job

import (
	"sync"

	"github.com/EscapeBearSecond/eagleeye/internal/global"
	"github.com/EscapeBearSecond/eagleeye/pkg/types"
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
