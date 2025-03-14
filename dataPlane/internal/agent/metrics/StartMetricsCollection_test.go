package metrics

import (
	"context"
	"dataPlane/internal/agent/metrics/protocol"
	"log"
	"net"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
)

// mockServer 是一个模拟的 gRPC 服务端，用于测试
type mockServer struct {
	protocol.UnimplementedMetricsServiceServer
	mu           sync.Mutex
	receivedData []*protocol.Metrics
}

// SendMetrics 实现了 MetricsService 服务的 SendMetrics 方法
func (s *mockServer) SendMetrics(ctx context.Context, req *protocol.Metrics) (*protocol.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.receivedData = append(s.receivedData, req)

	// 打印完整的 Metrics 对象信息（包括所有字段）
	log.Printf("Received full metrics data: %+v", req)

	return &protocol.Response{Status: "OK"}, nil
}

// TestStartMetricsCollection 测试 StartMetricsCollection 函数能否正确发送数据并被服务端接收
func TestStartMetricsCollection(t *testing.T) {
	// 创建一个监听随机端口的 TCP 监听器
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer lis.Close()

	// 创建模拟的 gRPC 服务端
	mockSrv := &mockServer{}
	grpcServer := grpc.NewServer()
	protocol.RegisterMetricsServiceServer(grpcServer, mockSrv)

	// 启动模拟的 gRPC 服务端
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Fatalf("Server exited with error: %v", err)
		}
	}()
	defer grpcServer.Stop()

	// 修改 ServerAddr 为模拟服务端的地址
	originalAddr := ServerAddr
	ServerAddr = lis.Addr().String()
	defer func() { ServerAddr = originalAddr }()

	// 修改 ReportInterval 为更短的时间，以便更快地进行测试
	originalInterval := ReportInterval
	ReportInterval = 1 * time.Second
	defer func() { ReportInterval = originalInterval }()

	// 创建带超时的上下文
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 定义超时限制，确保测试不无限期运行
	done := make(chan struct{})
	go func() {
		StartMetricsCollection()
		close(done)
	}()

	// 设置一个测试超时，避免测试挂起
	select {
	case <-done:
		// 测试正常完成
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out")
	}

	// 检查服务端是否收到了数据
	mockSrv.mu.Lock()
	defer mockSrv.mu.Unlock()
	if len(mockSrv.receivedData) == 0 {
		t.Error("No metrics data received by the server")
	} else {
		for _, metrics := range mockSrv.receivedData {
			if metrics.Ip == "" {
				t.Error("Received metrics data with empty IP")
			}
			if metrics.CpuInfo == nil {
				t.Error("Received metrics data without CPUInfo")
			}
			if metrics.MemoryInfo == nil {
				t.Error("Received metrics data without MemoryInfo")
			}
			if metrics.DiskInfo == nil {
				t.Error("Received metrics data without DiskInfo")
			}
			if metrics.NetworkInfo == nil {
				t.Error("Received metrics data without NetworkInfo")
			}
			if metrics.HostInfo == nil {
				t.Error("Received metrics data without HostInfo")
			}
			if metrics.LoadInfo == nil {
				t.Error("Received metrics data without LoadInfo")
			}
		}
	}
}
