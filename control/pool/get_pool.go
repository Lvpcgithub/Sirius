package pool

import (
	"github.com/panjf2000/ants/v2"
	"log"
	"sync"
)

var (
	pool *ants.PoolWithFunc
	once sync.Once
)

// 初始化协程池的函数
func InitPool(poolSize int, taskFunc func(interface{})) {
	once.Do(func() {
		var err error
		pool, err = ants.NewPoolWithFunc(poolSize, taskFunc)
		if err != nil {
			log.Fatalf("Failed to create pool: %v", err)
		}
	})
}

// 获取协程池实例
func GetPool() *ants.PoolWithFunc {
	return pool
}

// 关闭协程池
func ReleasePool() {
	if pool != nil {
		pool.Release()
	}
}
