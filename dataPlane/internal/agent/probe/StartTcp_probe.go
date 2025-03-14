package probe

func StartTcp_probe() {
	// 启动 ProbeTaskServiceServer，处理探测任务的接收
	go StartProbeTaskServiceServer()

	// 启动定时探测循环，定时执行 TCP 探测并上报
	go StartProbeLoop()

	// 保持程序运行
	select {}
}
