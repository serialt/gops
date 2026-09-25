package collector

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type Executor interface{ Execute(string) (string, error) }

type CPUInfo struct {
	ModelName string  `json:"model_name"`
	Cores     int     `json:"cores"`
	Threads   int     `json:"threads"`
	Usage     float64 `json:"usage"`
	VendorID  string  `json:"vendor_id"`
}
type MemoryInfo struct {
	Total, Used uint64
	Usage       float64
}
type DiskInfo struct {
	Total, Used uint64
	Usage       float64
}
type SystemInfo struct {
	OS, Kernel, Arch, Hostname, Uptime string
	CPU                                CPUInfo
	Memory                             MemoryInfo
	Disk                               DiskInfo
}

type Collector struct{ executor Executor }

func New(executor Executor) *Collector { return &Collector{executor: executor} }

func (c *Collector) CollectAll() (*SystemInfo, error) {
	info := new(SystemInfo)
	var wg sync.WaitGroup
	var cpuErr, memoryErr, diskErr error
	wg.Add(5)
	go func() { defer wg.Done(); info.CPU, cpuErr = c.collectCPU() }()
	go func() { defer wg.Done(); info.Memory, memoryErr = c.collectMemory() }()
	go func() { defer wg.Done(); info.Disk, diskErr = c.collectDisk() }()
	go func() { defer wg.Done(); info.OS, info.Kernel, info.Arch, info.Hostname = c.collectSystem() }()
	go func() { defer wg.Done(); info.Uptime = strings.TrimSpace(c.run("uptime -p 2>/dev/null || uptime")) }()
	wg.Wait()
	if cpuErr != nil && memoryErr != nil && diskErr != nil {
		return nil, fmt.Errorf("采集 CPU、内存和磁盘信息失败")
	}
	return info, nil
}

func (c *Collector) run(command string) string {
	output, _ := c.executor.Execute(command)
	return output
}
func (c *Collector) collectSystem() (string, string, string, string) {
	os := strings.TrimSpace(c.run(". /etc/os-release 2>/dev/null && echo \"$PRETTY_NAME\" || (cat /etc/redhat-release 2>/dev/null || head -1 /etc/issue 2>/dev/null)"))
	return os, strings.TrimSpace(c.run("uname -r")), strings.TrimSpace(c.run("uname -m")), strings.TrimSpace(c.run("hostname"))
}
func (c *Collector) collectCPU() (CPUInfo, error) {
	var info CPUInfo
	output, err := c.executor.Execute("lscpu 2>/dev/null || cat /proc/cpuinfo")
	if err != nil {
		return info, fmt.Errorf("获取 CPU 信息: %w", err)
	}
	for _, line := range strings.Split(output, "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		switch key {
		case "Model name", "model name":
			if info.ModelName == "" {
				info.ModelName = value
			}
		case "CPU(s)":
			if info.Threads == 0 {
				info.Threads, _ = strconv.Atoi(value)
			}
		case "Core(s) per socket":
			info.Cores, _ = strconv.Atoi(value)
		case "Vendor ID", "vendor_id":
			if info.VendorID == "" {
				info.VendorID = value
			}
		}
	}
	if info.Cores == 0 {
		info.Cores = info.Threads
	}
	usage := strings.TrimSpace(c.run("LC_ALL=C top -bn1 2>/dev/null | awk '/Cpu\\(s\\)/ {print 100-$8; exit}'"))
	info.Usage, _ = strconv.ParseFloat(usage, 64)
	return info, nil
}
func (c *Collector) collectMemory() (MemoryInfo, error) {
	var info MemoryInfo
	output, err := c.executor.Execute("free -b")
	if err != nil {
		return info, fmt.Errorf("获取内存信息: %w", err)
	}
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[0] == "Mem:" {
			info.Total, _ = strconv.ParseUint(fields[1], 10, 64)
			info.Used, _ = strconv.ParseUint(fields[2], 10, 64)
			break
		}
	}
	if info.Total > 0 {
		info.Usage = float64(info.Used) / float64(info.Total) * 100
	}
	return info, nil
}
func (c *Collector) collectDisk() (DiskInfo, error) {
	var info DiskInfo
	output, err := c.executor.Execute("df -B1 / | tail -1")
	if err != nil {
		return info, fmt.Errorf("获取磁盘信息: %w", err)
	}
	fields := strings.Fields(output)
	if len(fields) < 4 {
		return info, fmt.Errorf("无法解析磁盘信息")
	}
	info.Total, _ = strconv.ParseUint(fields[1], 10, 64)
	info.Used, _ = strconv.ParseUint(fields[2], 10, 64)
	if info.Total > 0 {
		info.Usage = float64(info.Used) / float64(info.Total) * 100
	}
	return info, nil
}
func (i CPUInfo) JSON() string { data, _ := json.Marshal(i); return string(data) }
