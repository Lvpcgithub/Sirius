package probe

import (
	"context"
	"dataPlane/internal/agent/probe/protocol"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"sync"
)

// 全局变量用于存储接收到的探测任务，同时使用互斥锁保证并发安全
var (
	probeTasks []*protocol.ProbeTask
	taskMutex  sync.Mutex
)

// ProbeTaskServiceServer 实现 ProbeTaskService 服务接口
type ProbeTaskServiceServer struct {
	protocol.UnimplementedProbeTaskServiceServer
}

// SendProbeTasks 实现 SendProbeTasks 方法
func (s *ProbeTaskServiceServer) SendProbeTasks(ctx context.Context, request *protocol.ProbeTaskRequest) (*protocol.ProbeTaskResponse, error) {
	// 加锁，保证并发安全
	taskMutex.Lock()
	// 覆盖之前的任务
	probeTasks = make([]*protocol.ProbeTask, 0, len(request.Tasks))
	probeTasks = append(probeTasks, request.Tasks...)
	taskMutex.Unlock()

	// 打印接收到的任务信息
	for _, task := range request.Tasks {
		fmt.Printf("Received probe task: Source IP: %s, Destination IP: %s\n", task.Ip1, task.Ip2)
	}

	// 返回响应
	response := &protocol.ProbeTaskResponse{
		Status: "ok",
	}
	return response, nil
}

// GetProbeTasks 用于获取缓存的探测任务
func GetProbeTasks() []*protocol.ProbeTask {
	taskMutex.Lock()
	// 复制一份任务，避免外部修改
	tasks := make([]*protocol.ProbeTask, len(probeTasks))
	copy(tasks, probeTasks)
	taskMutex.Unlock()
	return tasks
}

// StartProbeTaskServiceServer 启动 ProbeTaskService 服务
func StartProbeTaskServiceServer() {
	// 创建 gRPC 服务器
	server := grpc.NewServer()

	// 注册 ProbeTaskService 服务
	protocol.RegisterProbeTaskServiceServer(server, &ProbeTaskServiceServer{})

	// 监听端口
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v\n", err)
	}

	// 启动服务器
	fmt.Println("ProbeTaskService server is listening on port 50051")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v\n", err)
	}
}
