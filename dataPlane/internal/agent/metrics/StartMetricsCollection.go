package metrics

import (
	"context"
	"dataPlane/internal/agent/metrics/protocol"
	"fmt"
	"log"
	"math"
	"time"
)

// 包级全局变量，可在外部修改
var (
	ServerAddr     = "localhost:50051" // 默认服务端地址
	ReportInterval = 30 * time.Second  // 默认上报间隔
)

// convertToProtoMetrics 辅助函数，用于将 InfoData 转换为 protocol.Metrics
func convertToProtoMetrics(info InfoData) *protocol.Metrics {
	// 将字节单位转换为兆（MB）
	toMB := func(bytes uint64) uint64 {
		return (uint64(bytes) / (1024 * 1024) * 100) / 100 // 转换为MB
	}

	return &protocol.Metrics{
		Ip: info.IP,
		CpuInfo: &protocol.CPUInfo{
			Cores:     info.CPUInfo.Cores,
			ModelName: info.CPUInfo.ModelName,
			Mhz:       info.CPUInfo.Mhz,
			CacheSize: info.CPUInfo.CacheSize,
			// 保留两位小数
			Usage: math.Round(info.CPUInfo.Usage*100) / 100,
		},
		MemoryInfo: &protocol.MemoryInfo{
			// 将内存信息从字节转换为兆（MB）
			Total:     toMB(info.MemoryInfo.Total),
			Available: toMB(info.MemoryInfo.Available),
			Used:      toMB(info.MemoryInfo.Used),
			// 保留两位小数
			UsedPercent: math.Round(info.MemoryInfo.UsedPercent*100) / 100,
		},
		DiskInfo: &protocol.DiskInfo{
			Device: info.DiskInfo.Device,
			// 将磁盘信息从字节转换为兆（MB）
			Total: toMB(info.DiskInfo.Total),
			Free:  toMB(info.DiskInfo.Free),
			Used:  toMB(info.DiskInfo.Used),
			// 保留两位小数
			UsedPercent: math.Round(info.DiskInfo.UsedPercent*100) / 100,
		},
		NetworkInfo: &protocol.NetworkInfo{
			InterfaceName: info.NetworkInfo.InterfaceName,
			BytesSent:     info.NetworkInfo.BytesSent,
			BytesRecv:     info.NetworkInfo.BytesRecv,
			PacketsSent:   info.NetworkInfo.PacketsSent,
			PacketsRecv:   info.NetworkInfo.PacketsRecv,
		},
		HostInfo: &protocol.HostInfo{
			Hostname:        info.HostInfo.Hostname,
			Os:              info.HostInfo.OS,
			Platform:        info.HostInfo.Platform,
			PlatformVersion: info.HostInfo.PlatformVersion,
			Uptime:          info.HostInfo.Uptime,
		},
		LoadInfo: &protocol.LoadInfo{
			// 保留两位小数
			Load1:  math.Round(info.LoadInfo.Load1*100) / 100,
			Load5:  math.Round(info.LoadInfo.Load5*100) / 100,
			Load15: math.Round(info.LoadInfo.Load15*100) / 100,
		},
	}
}

func StartMetricsCollection() {
	// 创建gRPC客户端，连接到控制面服务器
	grpcClient, err := NewGrpcClient(ServerAddr)
	if err != nil {
		log.Fatalf("Error creating gRPC client: %v", err)
	}
	defer grpcClient.Close()

	// 设置定时器
	ticker := time.NewTicker(ReportInterval)
	defer ticker.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		select {
		case <-ticker.C:
			// 创建带超时的上下文
			uploadCtx, uploadCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer uploadCancel()

			// 收集系统信息
			info, err := CollectSystemInfo()
			if err != nil {
				log.Printf("Error collecting system info: %v", err)
				continue
			}

			// 创建Metrics数据结构
			metricsData := convertToProtoMetrics(info)

			// 上传数据到控制面
			err = grpcClient.UploadMetrics(uploadCtx, metricsData)
			if err != nil {
				log.Printf("Error sending metrics: %v", err)
			} else {
				fmt.Println("Metrics successfully sent to control plane")
			}
		case <-ctx.Done():
			return
		}
	}
}
