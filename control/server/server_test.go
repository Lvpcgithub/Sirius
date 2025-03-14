package server

import (
	"context"
	"control/dao"
	"control/pool"
	"fmt"
	"log"
	"testing"
	"time"
)
//测试接收节点信息
func TestServer(t *testing.T) {
	ReceiveMetrics()
}
//测试下发探测任务
func TestCreateProbeTasks(t *testing.T) {
	//调用配置文件方法
	c :=dao.UseToml()
	// 创建 Context 支持优雅退出
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// 连接数据库
	db := dao.ConnectToDB()
	conn := dao.ConnRedis()
	defer db.Close()
	defer conn.Close()
	// 初始化协程池
	poolSize := c.PoolNum // 协程池大小
	pool.InitPool(poolSize, taskHandler)
	defer pool.ReleasePool()
	// 先立即下发一次任务
	SendProbeTasksOnce(db)
	// 启动定时器，每隔 30 s下发一次任务
	interval := c.DetectCycle * time.Second
	
	go createProbeTasksWithTimer(ctx, db, conn, interval,10 *time.Second)
	// 程序运行
	log.Println("Probe task scheduler started. Press Ctrl+C to stop.")
	time.Sleep(30 * time.Minute) // 程序运行 30 分钟
	cancel()                     // 停止定时任务
	log.Println("Probe task scheduler stopped")
}
//测试接收探测信息，并存储
func TestReceiveProbe(t *testing.T) {
	fmt.Println("ReceiveProbe已启动！")
	ReceiveProbe()
}
