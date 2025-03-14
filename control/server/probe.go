package server

import (
	"context"
	"control/models"
	"control/pool"
	pb "control/proto"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"
	"github.com/gomodule/redigo/redis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// 探测任务下发函数
func sendProbeTask(client pb.ProbeTaskServiceClient, ip1, ip2 string) {
	// 填充探测任务
	req := &pb.ProbeTaskRequest{
		Tasks: []*pb.ProbeTask{
			{
				Ip1: ip1,
				Ip2: ip2,
			},
		},
	}

	// 调用 gRPC 方法
	resp, err := client.SendProbeTasks(context.Background(), req)
	if err != nil {
		log.Printf("Failed to send probe task from %s to %s: %v", ip1, ip2, err)
		return
	}
	log.Printf("Probe task from %s to %s sent successfully. Status: %s", ip1, ip2, resp.Status)
}
// 任务处理函数
func taskHandler(data interface{}) {
	// 获取任务参数
	params := data.([]interface{})
	ip1 := params[0].(string)
	ipaddrs := params[1].([]string)

	// 连接到 gRPC 服务器
	conn, err := grpc.Dial(fmt.Sprintf("%s:50051", ip1), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to connect to gRPC server at %s: %v", ip1, err)
		return
	}
	defer conn.Close()

	// 创建 gRPC 客户端
	client := pb.NewProbeTaskServiceClient(conn)

	// 将当前 IP 与其他 IP 组合，发送探测任务
	for _, ip2 := range ipaddrs {
		if ip1 != ip2 { // 避免自己探测自己
			sendProbeTask(client, ip1, ip2)
		}
	}
}
// 立即下发一次探测任务
func SendProbeTasksOnce(db *sql.DB) {
	// 查询 IP 列表
	ipaddrs, err := models.QueryIp(db)
	if err != nil {
		log.Fatalf("Failed to query IPs: %v", err)
	}

	// 使用 WaitGroup 等待所有任务完成
	var wg sync.WaitGroup

	// 遍历 IP 列表，提交任务到协程池
	for _, ip1 := range ipaddrs {
		wg.Add(1)
		go func(ip1 string) {
			defer wg.Done()
			// 提交任务到协程池
			err := pool.GetPool().Invoke([]interface{}{ip1, ipaddrs})
			if err != nil {
				log.Printf("Failed to submit task for IP %s: %v", ip1, err)
			}
		}(ip1)
	}

	// 等待当前批次任务完成
	wg.Wait()
	log.Println("Initial batch of probe tasks completed")
}
// 定时下发探测任务
func createProbeTasksWithTimer(ctx context.Context, db *sql.DB, conn redis.Conn , interval time.Duration ,computerInterval time.Duration) {
	// 创建定时器
	ticker := time.NewTicker(interval)
	tickerComputer := time.NewTicker(computerInterval)
	defer ticker.Stop()
	defer tickerComputer.Stop()

	// 使用 WaitGroup 等待所有任务完成
	var wg sync.WaitGroup

	// 定时任务循环
	for {
		select {
			// 创建通道，模拟手动停止
		case <-ctx.Done():
			log.Println("Stopping probe task scheduler...")
			return
		case <-ticker.C:
			// 查询 IP 列表
			ipaddrs, err := models.QueryIp(db)
			if err != nil {
				log.Printf("Failed to query IPs: %v", err)
				continue
			}

			// 遍历 IP 列表，提交任务到协程池
			for _, ip1 := range ipaddrs {
				wg.Add(1)
				go func(ip1 string) {
					defer wg.Done()

					// 提交任务到协程池
					err := pool.GetPool().Invoke([]interface{}{ip1, ipaddrs})
					if err != nil {
						log.Printf("Failed to submit task for IP %s: %v", ip1, err)
					}
				}(ip1)
			}

			// 等待当前批次任务完成
			wg.Wait()
			log.Println("Current batch of probe tasks completed")
		case <-tickerComputer.C:
			//定时拿到数据并计算存到mysql里面去
			ipAddresses, err := models.QueryIp(db)
			if err != nil {
				log.Printf("Failed to query IPs: %v", err)
				continue
			}
			for i := 0; i < len(ipAddresses); i++ {
				for j := 0; j < len(ipAddresses); j++ {
					if i != j {
						// 计算并存储
						models.CalculateAvgDelay(conn,db,ipAddresses[i],ipAddresses[j])
					}
				}
			}
		}
	}
}