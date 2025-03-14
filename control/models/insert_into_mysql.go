package models

import (
	pb "control/proto"
	"database/sql"
	"fmt"
	"time"
)

// 插入节点信息
func InsertMetricsInfo(db *sql.DB, info *pb.Metrics) error {
	query := `
		INSERT INTO system_info (
			ip, 
			cpu_cores, cpu_model_name, cpu_mhz, cpu_cache_size, cpu_usage,
			memory_total, memory_available, memory_used, memory_used_percent,
			disk_device, disk_total, disk_free, disk_used, disk_used_percent,
			network_interface_name, network_bytes_sent, network_bytes_recv,
			network_packets_sent, network_packets_recv,
			hostname, os, platform, platform_version, uptime,
			load1, load5, load15, timestamp
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("Timestamp: %s\n", timestamp)
	_, err := db.Exec(query,
		info.Ip,
		info.CpuInfo.Cores, info.CpuInfo.ModelName, info.CpuInfo.Mhz, info.CpuInfo.CacheSize, info.CpuInfo.Usage, //5
		info.MemoryInfo.Total, info.MemoryInfo.Available, info.MemoryInfo.Used, info.MemoryInfo.UsedPercent, //4
		info.DiskInfo.Device, info.DiskInfo.Total, info.DiskInfo.Free, info.DiskInfo.Used, info.DiskInfo.UsedPercent, //5
		info.NetworkInfo.InterfaceName, info.NetworkInfo.BytesSent, info.NetworkInfo.BytesRecv, //3
		info.NetworkInfo.PacketsSent, info.NetworkInfo.PacketsRecv, //2
		info.HostInfo.Hostname, info.HostInfo.Os, info.HostInfo.Platform, info.HostInfo.PlatformVersion, info.HostInfo.Uptime, //5
		info.LoadInfo.Load1, info.LoadInfo.Load5, info.LoadInfo.Load15, timestamp, //4
	)
	return err
}

// 插入链路信息
func InsertLinkInfo(db *sql.DB, sourceIP string, destinationIP string, delay float64, timestamp string) error {
	query := `
		INSERT INTO link_info (SourceIP, DestinationIP, Delay, Timestamp)
		VALUES (?, ?, ?, ?)
	`
	_, err := db.Exec(query, sourceIP, destinationIP, delay, timestamp)
	return err
}
