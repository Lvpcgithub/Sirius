package config

import "time"

//配置文件结构体
type ConfigInfo struct {
	PoolNum        int           //协程池数量
	ReceivePort    string        //接收节点信息端口号
	DetectPort     string		 //接收探测信息端口号
	DetectCycle    time.Duration //下发一次探测任务时长 单位ns *time.Second 变成秒
	ExpireDuration time.Duration //redis列表过期
	CalculateCycle time.Duration // redis计算周期
	K              int           //路径数量
	Theta          float64       //惩罚系数
	Skip           int           //跳数限制
}

//探测结构体
type ProbeResult struct {
	SourceIP      string  `json:"ip1"`
	DestinationIP string  `json:"ip2"`
	Delay         int64 `json:"tcp_delay"`
	Timestamp     string  `json:"timestamp"`
}

