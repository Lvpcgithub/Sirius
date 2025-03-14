package probe

import (
	"context"
	"dataPlane/internal/agent/probe/protocol"
	"fmt"
	"google.golang.org/grpc"
	"net"
	"time"
)

// 定义全局变量来配置 gRPC 客户端地址和探测循环时间间隔
var (
	GRPCClientAddr = "124.70.34.63:8081"
	ProbeInterval  = 10 * time.Second
)

// ProbeTask 结构体定义
type ProbeTask struct {
	IP1 string
	IP2 string
}

// ProbeResult 结构体定义
type ProbeResult struct {
	IP1       string
	IP2       string
	TCPDelay  int64 // 直接使用 int64 存储毫秒数
	Timestamp time.Time
}

// performTCPProbe 执行 TCP 探测并返回探测结果
func performTCPProbe(ip1 string, ip2 string) (*ProbeResult, error) {
	// 记录开始时间
	startTime := time.Now()

	// 建立 TCP 连接
	conn, err := net.DialTimeout("tcp", ip2+":50051", 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("error connecting to %s: %v", ip2, err)
	}
	defer conn.Close()

	// 计算 TCP 延迟（转换为毫秒）
	tcpDelay := time.Since(startTime).Milliseconds()

	// 返回探测结果
	result := &ProbeResult{
		IP1:       ip1,
		IP2:       ip2,
		TCPDelay:  tcpDelay,
		Timestamp: time.Now(),
	}

	return result, nil
}

// SendProbeResults 发送探测结果
func SendProbeResults(results []*ProbeResult) {
	// 连接到 gRPC 服务器，使用全局变量 GRPCClientAddr
	conn, err := grpc.Dial(GRPCClientAddr, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}
	defer conn.Close()

	// 创建 ProbeResultService 客户端
	client := protocol.NewProbeResultServiceClient(conn)

	// 创建 ProbeResultRequest 消息
	var protoResults []*protocol.ProbeResult
	for _, result := range results {
		protoResults = append(protoResults, &protocol.ProbeResult{
			Ip1:       result.IP1,
			Ip2:       result.IP2,
			TcpDelay:  result.TCPDelay, // 直接使用毫秒值
			Timestamp: result.Timestamp.Format(time.RFC3339),
		})
	}
	request := &protocol.ProbeResultRequest{
		Results: protoResults,
	}

	// 调用 SendProbeResults 方法
	response, err := client.SendProbeResults(context.Background(), request)
	if err != nil {
		fmt.Printf("Failed to send probe results: %v\n", err)
		return
	}

	// 打印响应消息
	fmt.Printf("Response: %s\n", response.Status)
}

// StartProbeLoop 启动定时探测循环
func StartProbeLoop() {
	// 使用全局变量 ProbeInterval 创建定时器
	ticker := time.NewTicker(ProbeInterval)
	defer ticker.Stop()

	for range ticker.C {
		// 调用 GetProbeTasks 函数获取探测任务
		tasks := GetProbeTasks()
		var results []*ProbeResult
		for _, task := range tasks {
			result, err := performTCPProbe(task.Ip1, task.Ip2)
			if err != nil {
				fmt.Printf("Error performing probe for %s -> %s: %v\n", task.Ip1, task.Ip2, err)
				continue
			}
			results = append(results, result)
		}
		if len(results) > 0 {
			SendProbeResults(results)
		}
	}
}
