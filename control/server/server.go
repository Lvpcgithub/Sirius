package server

import (
	"context"
	"control/config"
	"control/dao"
	"control/models"
	pb "control/proto"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/gomodule/redigo/redis"
	"google.golang.org/grpc"
)

// 节点信息接收结构体重写
type Server struct {
	pb.UnimplementedMetricsServiceServer
}
// 探测结构体重写
type Probe struct {
	pb.UnimplementedProbeResultServiceServer
}

// 节点信息上传方法实现
func (s *Server) SendMetrics(ctx context.Context, req *pb.Metrics) (*pb.Response, error) {
	//连接数据库
	db := dao.ConnectToDB()
	if db == nil {
		return &pb.Response{Status: "error"}, fmt.Errorf("unable to connect to the database")
	}
	defer db.Close()
        fmt.Println(req)
        if req==nil {
		return &pb.Response{Status: "error"}, fmt.Errorf("invalid request")
	}
	// 将数据插入数据库，调用sql语句
	err := models.InsertMetricsInfo(db, req)
	if err != nil {
		return nil, err
	}
	return &pb.Response{Status: "ok"}, nil
}
// 开启8080端口，接收节点信息上报
func ReceiveMetrics() {
	c := dao.UseToml()
	// 开启端口
	listen, err := net.Listen("tcp", "0.0.0.0:"+c.ReceivePort)
	if err != nil {
		panic(err)
	}
	//创建grpc服务
	grpcServer := grpc.NewServer()
	//注册服务
	pb.RegisterMetricsServiceServer(grpcServer, &Server{})
	//启动服务
	err = grpcServer.Serve(listen)
	if err != nil {
		panic(err)
	}
}

// SendProbeResults 接收探测结果并处理
func (p *Probe) SendProbeResults(ctx context.Context, req *pb.ProbeResultRequest) (*pb.ProbeResultResponse, error) {
	// 获取 Redis 连接
	conn := dao.ConnRedis()
	c := dao.UseToml()
	// 设置列表的过期时间（单位：hour）
	expireDuration := c.ExpireDuration * time.Hour// 一天
	// 遍历探测结果，处理探测结果，存入 Redis 里
	for _, result := range req.Results {
		log.Printf("Received probe result: IP1=%s, IP2=%s, TCP Delay=%d ms, Timestamp=%s",
			result.Ip1, result.Ip2, result.TcpDelay, result.Timestamp)
		// 组合 Redis 键，key 为 ip1:ip2
		key := result.Ip1 + ":" + result.Ip2
		value, err := json.Marshal(config.ProbeResult{
			SourceIP:      result.Ip1,
			DestinationIP: result.Ip2,
			Delay:         result.TcpDelay,
			Timestamp:     result.Timestamp,
		})
		if err != nil {
			log.Printf("Error marshalling result to JSON: %v", err)
			return nil, err
		}
		// 检查 key 是否存在
		exists, err := redis.Int(conn.Do("EXISTS", key)) // 使用 ConnRedis() 封装的连接
		if err != nil {
			log.Printf("Error checking if key exists: %v", err)
			return nil, err
		}
		// 如果 key 不存在，说明已经过期，重新插入并设置过期时间
		if exists == 0 {
			// 使用 Redis 的 LPUSH 命令将数据插入列表
			_, err := conn.Do("LPUSH", key, value)
			if err != nil {
				log.Printf("Error storing result in redis: %v", err)
				return nil, err
			}
			// 设置过期时间（仅在首次插入时）
			_, err = conn.Do("EXPIRE", key, expireDuration)
			if err != nil {
				log.Printf("Error setting expiration for key: %v", err)
				return nil, err
			}
		} 
		//key 存在，直接插入数据，
		_, lpushErr := conn.Do("LPUSH", key, value)
		if lpushErr != nil {
			log.Printf("Error storing result in redis: %v", lpushErr)
			return nil, lpushErr
		}
	}
	// 返回成功响应
	return &pb.ProbeResultResponse{Status: "ok"}, nil
}

// 开启8081端口，接收探测信息
func ReceiveProbe() {
	c := dao.UseToml()
	// 创建 gRPC 服务器
	server := grpc.NewServer()

	// 注册 ProbeResultService
	pb.RegisterProbeResultServiceServer(server, &Probe{})

	// 监听端口
	lis, err := net.Listen("tcp", "0.0.0.0:" + c.DetectPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	// 启动服务器
	log.Println("ProbeResultService server is running on port 8081...")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}