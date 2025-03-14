package probe

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"testing"
	"time"

	// 假设这是你的协议生成的 Go 代码包路径
	"dataPlane/internal/agent/probe/protocol"
)

// ProbeResultServiceServer 实现 ProbeResultService 服务接口
type ProbeResultServiceServer struct {
	protocol.UnimplementedProbeResultServiceServer
	t *testing.T
}

// SendProbeResults 实现 SendProbeResults 方法，用于接收探测结果
func (s *ProbeResultServiceServer) SendProbeResults(ctx context.Context, request *protocol.ProbeResultRequest) (*protocol.ProbeResultResponse, error) {
	for _, result := range request.Results {
		// 修改日志输出，将单位从 ns 改为 ms
		fmt.Printf("Received probe result: Source IP: %s, Destination IP: %s, TCP Delay: %d ms, Timestamp: %s\n",
			result.Ip1, result.Ip2, result.TcpDelay, result.Timestamp)
	}
	// 简单断言，验证接收到的结果数量是否符合预期
	if len(request.Results) != 2 {
		s.t.Errorf("Expected 2 results, got %d", len(request.Results))
	}
	response := &protocol.ProbeResultResponse{
		Status: "ok",
	}
	return response, nil
}

// 启动 gRPC 服务端，用于接收探测结果
func startServer(t *testing.T) {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	server := &ProbeResultServiceServer{t: t}
	protocol.RegisterProbeResultServiceServer(s, server)
	log.Printf("Server listening at %v", lis.Addr())
	go func() {
		if err := s.Serve(lis); err != nil {
			t.Fatalf("failed to serve: %v", err)
		}
	}()
}

// 发送探测任务
func sendProbeTasks() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := protocol.NewProbeTaskServiceClient(conn)

	// 创建探测任务
	tasks := []*protocol.ProbeTask{
		{
			Ip1: "192.168.1.100",
			Ip2: "127.0.0.1",
		},
		{
			Ip1: "192.168.1.102",
			Ip2: "127.0.0.1",
		},
	}

	request := &protocol.ProbeTaskRequest{
		Tasks: tasks,
	}

	// 发送探测任务
	response, err := c.SendProbeTasks(context.Background(), request)
	if err != nil {
		log.Fatalf("could not send probe tasks: %v", err)
	}
	fmt.Printf("Probe tasks sent. Response status: %s\n", response.Status)
}

func TestStartTcp_probe(t *testing.T) {
	// 修改 gRPC 客户端地址，让数据面将结果上报到测试服务端
	GRPCClientAddr = "localhost:50052"
	// 缩短探测间隔，加速测试
	ProbeInterval = 1 * time.Second

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 启动被测试模块
	go StartTcp_probe()

	// 启动服务端，用于接收探测结果
	go startServer(t)

	// 等待服务端启动
	time.Sleep(1 * time.Second)

	// 发送探测任务
	go sendProbeTasks()

	// 等待测试超时
	<-ctx.Done()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatal("test timed out")
	}
}
