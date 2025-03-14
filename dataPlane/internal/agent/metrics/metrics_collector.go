package metrics

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"  // 获取CPU信息和使用率
	"github.com/shirou/gopsutil/v3/disk" // 获取磁盘信息，如分区和使用情况
	"github.com/shirou/gopsutil/v3/host" // 获取主机信息，如操作系统、平台等
	"github.com/shirou/gopsutil/v3/load" // 获取系统平均负载信息
	"github.com/shirou/gopsutil/v3/mem"  // 获取内存信息，如总量、使用量等
	"github.com/shirou/gopsutil/v3/net"  // 获取网络接口的I/O统计信息
	"io"
	"net/http" // 执行HTTP请求
	"strings"
	"time"
)

type CPUInfo struct {
	Cores     int32
	ModelName string
	Mhz       float64
	CacheSize int32
	Usage     float64
}

type MemoryInfo struct {
	Total       uint64
	Available   uint64
	Used        uint64
	UsedPercent float64
}

type DiskInfo struct {
	Device      string
	Total       uint64
	Free        uint64
	Used        uint64
	UsedPercent float64
}

type NetworkInfo struct {
	InterfaceName string
	BytesSent     uint64
	BytesRecv     uint64
	PacketsSent   uint64
	PacketsRecv   uint64
}

type HostInfo struct {
	Hostname        string
	OS              string
	Platform        string
	PlatformVersion string
	Uptime          uint64
}

type LoadInfo struct {
	Load1  float64
	Load5  float64
	Load15 float64
}

type InfoData struct {
	IP          string
	CPUInfo     CPUInfo
	MemoryInfo  MemoryInfo
	DiskInfo    DiskInfo
	NetworkInfo NetworkInfo
	HostInfo    HostInfo
	LoadInfo    LoadInfo
}

// GetIP 获取公网IP地址，使用多个备用服务提高可靠性
func GetIP() (string, error) {
	urls := []string{
		"http://icanhazip.com",
		"http://api.ipify.org",
		"http://ifconfig.me/ip",
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	var lastErr error
	for _, url := range urls {
		resp, err := client.Get(url)
		if err != nil {
			lastErr = fmt.Errorf("failed to query %s: %w", url, err)
			continue
		}

		ipBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response from %s: %w", url, err)
			continue
		}

		ip := strings.TrimSpace(string(ipBytes))
		if ip != "" {
			return ip, nil
		}
	}

	return "", fmt.Errorf("all IP services failed: %w", lastErr)
}

// GetCPUInfo 获取整体CPU信息及使用率
// 返回一个CPUInfo结构体，包含整体CPU的信息和使用率
func GetCPUInfo() (CPUInfo, error) {
	// 获取CPU的基本信息
	infos, err := cpu.Info()
	if err != nil {
		return CPUInfo{}, fmt.Errorf("Failed to get CPU Info: %v", err)
	}

	// 获取每个CPU核心的使用率
	usage, err := cpu.Percent(0, true)
	if err != nil {
		return CPUInfo{}, fmt.Errorf("Failed to get CPU usage: %v", err)
	}

	// 汇总CPU信息
	var totalCores int32
	var modelName string
	var mhz float64
	var cacheSize int32

	if len(infos) > 0 {
		totalCores = infos[0].Cores
		modelName = infos[0].ModelName
		mhz = infos[0].Mhz
		cacheSize = infos[0].CacheSize
	}

	// 计算整体使用率（取所有核心使用率的平均值）
	var totalUsage float64
	for _, u := range usage {
		totalUsage += u
	}
	averageUsage := totalUsage / float64(len(usage))

	// 创建并返回整体CPU信息
	return CPUInfo{
		Cores:     totalCores,
		ModelName: modelName,
		Mhz:       mhz,
		CacheSize: cacheSize,
		Usage:     averageUsage,
	}, nil
}

// GetMemoryInfo 获取内存信息
// 返回一个MemoryInfo结构体，包含内存的总量、可用量、使用量等
func GetMemoryInfo() (MemoryInfo, error) {
	// 获取内存的使用情况
	v, err := mem.VirtualMemory()
	if err != nil {
		return MemoryInfo{}, fmt.Errorf("Failed to get memory info: %v", err)
	}

	// 返回内存信息的结构体
	return MemoryInfo{
		Total:       v.Total,
		Available:   v.Available,
		Used:        v.Used,
		UsedPercent: v.UsedPercent,
	}, nil
}

// GetDiskInfo 获取整体磁盘信息
// 返回一个DiskInfo结构体，包含整个磁盘的使用情况
func GetDiskInfo() (DiskInfo, error) {
	// 获取根分区（"/"）的使用情况，这通常代表整个系统的磁盘使用情况
	usage, err := disk.Usage("/")
	if err != nil {
		return DiskInfo{}, fmt.Errorf("Failed to get disk usage: %v", err)
	}

	// 返回整体磁盘信息
	return DiskInfo{
		Device:      "Overall",
		Total:       usage.Total,
		Free:        usage.Free,
		Used:        usage.Used,
		UsedPercent: usage.UsedPercent,
	}, nil
}

// GetNetworkInfo 获取接口的I/O统计信息
// 返回一个NetworkInfo结构体，包含接口的流量统计信息
func GetNetworkInfo() (NetworkInfo, error) {
	interfaces, err := net.IOCounters(true)
	if err != nil {
		return NetworkInfo{}, fmt.Errorf("Failed to get network interfaces: %v", err)
	}

	// 遍历所有接口，选择第一个非环回接口
	for _, iface := range interfaces {
		if iface.Name != "lo" && iface.Name != "lo0" {
			return NetworkInfo{
				InterfaceName: iface.Name,
				BytesSent:     iface.BytesSent,
				BytesRecv:     iface.BytesRecv,
				PacketsSent:   iface.PacketsSent,
				PacketsRecv:   iface.PacketsRecv,
			}, nil
		}
	}

	return NetworkInfo{}, fmt.Errorf("No non-loopback interface found")
}

// GetHostInfo 获取主机信息
// 返回一个HostInfo结构体，包含主机的名称、操作系统、平台版本等信息
func GetHostInfo() (HostInfo, error) {
	// 获取主机的基本信息
	info, err := host.Info()
	if err != nil {
		return HostInfo{}, fmt.Errorf("Failed to get host info: %v", err)
	}

	// 返回主机信息的结构体
	return HostInfo{
		Hostname:        info.Hostname,
		OS:              info.OS,
		Platform:        info.Platform,
		PlatformVersion: info.PlatformVersion,
		Uptime:          info.Uptime,
	}, nil
}

// GetLoadInfo 获取系统平均负载信息
// 返回一个LoadInfo结构体，包含系统在过去1分钟、5分钟和15分钟内的平均负载
func GetLoadInfo() (LoadInfo, error) {
	// 获取系统平均负载信息
	avg, err := load.Avg()
	if err != nil {
		return LoadInfo{}, fmt.Errorf("Failed to get system load: %v", err)
	}

	// 返回平均负载信息的结构体
	return LoadInfo{
		Load1:  avg.Load1,
		Load5:  avg.Load5,
		Load15: avg.Load15,
	}, nil
}

// CollectSystemInfo 收集所有系统信息并返回
// 通过调用上述所有函数收集系统信息，并将它们组合成一个InfoData结构体返回
func CollectSystemInfo() (InfoData, error) {
	ip, err := GetIP()
	if err != nil {
		return InfoData{}, err
	}

	cpuInfo, err := GetCPUInfo()
	if err != nil {
		return InfoData{}, err
	}

	memoryInfo, err := GetMemoryInfo()
	if err != nil {
		return InfoData{}, err
	}

	diskInfo, err := GetDiskInfo()
	if err != nil {
		return InfoData{}, err
	}

	networkInfo, err := GetNetworkInfo()
	if err != nil {
		return InfoData{}, err
	}

	hostInfo, err := GetHostInfo()
	if err != nil {
		return InfoData{}, err
	}

	loadInfo, err := GetLoadInfo()
	if err != nil {
		return InfoData{}, err
	}

	return InfoData{
		IP:          ip,
		CPUInfo:     cpuInfo,
		MemoryInfo:  memoryInfo,
		DiskInfo:    diskInfo,
		NetworkInfo: networkInfo,
		HostInfo:    hostInfo,
		LoadInfo:    loadInfo,
	}, nil
}
