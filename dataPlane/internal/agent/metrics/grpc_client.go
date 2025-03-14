package metrics

import (
	"context"
	"dataPlane/internal/agent/metrics/protocol" // 引入由protobuf生成的protocol包
	"fmt"
	"google.golang.org/grpc"
	"log"
	"time"
)

// GrpcClient 用于管理与控制面服务器的连接
type GrpcClient struct {
	client protocol.MetricsServiceClient
	conn   *grpc.ClientConn
}

// NewGrpcClient 创建并返回 gRPC 客户端实例
func NewGrpcClient(address string) (*GrpcClient, error) {
	var conn *grpc.ClientConn
	var err error
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		conn, err = grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
		if err == nil {
			break
		}
		if i < maxRetries-1 {
			time.Sleep(2 * time.Second)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to control plane after %d retries: %v", maxRetries, err)
	}
	client := protocol.NewMetricsServiceClient(conn) // 使用protobuf生成的客户端
	return &GrpcClient{client: client, conn: conn}, nil
}

// UploadMetrics 将性能数据上传到控制面服务器
func (g *GrpcClient) UploadMetrics(ctx context.Context, metrics *protocol.Metrics) error {
	resp, err := g.client.SendMetrics(ctx, metrics) // 发送Metrics数据
	if err != nil {
		log.Printf("Failed to send metrics: %v", err)
		return fmt.Errorf("failed to send metrics: %v", err)
	}
	log.Printf("Server returned status: %s", resp.Status)
	return nil
}

// Close 关闭与控制面服务器的连接
func (g *GrpcClient) Close() {
	// 关闭gRPC连接
	g.conn.Close()
}
