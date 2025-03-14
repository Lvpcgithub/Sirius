package main

import (
	"dataPlane/internal/agent/metrics" //
	"dataPlane/internal/agent/probe"
	"github.com/panjf2000/ants/v2" // 引入 ants 包
	"log"
)

func main() {
	log.Println("Starting metrics collection with ants goroutine pool...")

	// 创建一个固定大小的协程池，这里假设池大小为10
	poolSize := 10
	pool, err := ants.NewPool(poolSize)
	if err != nil {
		log.Fatalf("Failed to create ants pool: %v", err)
	}
	defer pool.Release() // 程序结束时释放协程池

	// 向协程池提交第一个任务：开始指标收集
	err = pool.Submit(func() {
		metrics.StartMetricsCollection()
	})
	if err != nil {
		log.Fatalf("Failed to submit task to ants pool: %v", err)
	}

	// 向协程池提交第二个任务：执行TCP探测
	err = pool.Submit(func() {
		probe.StartTcp_probe()
	})
	if err != nil {
		log.Fatalf("Failed to submit task to ants pool: %v", err)
	}

	// 阻塞主协程，防止程序退出
	select {}
}
