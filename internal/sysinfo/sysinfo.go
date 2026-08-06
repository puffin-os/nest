// Package sysinfo gathers system information using gopsutil.
package sysinfo

import (
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

// CPUInfo holds CPU-related information.
type CPUInfo struct {
	Model     string    `json:"model"`
	Cores     int       `json:"cores"`
	Sockets   int       `json:"sockets"`
	Frequency float64 `json:"frequency_mhz"`
	Usage     []float64 `json:"usage_percent"`
}

// MemoryInfo holds memory-related information.
type MemoryInfo struct {
	Total     uint64  `json:"total_bytes"`
	Used      uint64  `json:"used_bytes"`
	Available uint64  `json:"available_bytes"`
	Usage     float64 `json:"usage_percent"`
}

// DiskInfo holds information about a single disk partition.
type DiskInfo struct {
	Device      string  `json:"device"`
	Mountpoint  string  `json:"mountpoint"`
	Fstype      string  `json:"fstype"`
	Total       uint64  `json:"total_bytes"`
	Used        uint64  `json:"used_bytes"`
	Free        uint64  `json:"free_bytes"`
	UsagePercent float64 `json:"usage_percent"`
}

// NetworkInterface holds information about a network interface.
type NetworkInterface struct {
	Name      string `json:"name"`
	MTU       int    `json:"mtu"`
	Hardware  string `json:"hardware_addr"`
	IPs       []string `json:"ip_addresses"`
	BytesSent uint64 `json:"bytes_sent"`
	BytesRecv uint64 `json:"bytes_recv"`
}

// OSInfo holds operating system information.
type OSInfo struct {
	Hostname  string `json:"hostname"`
	OS        string `json:"os"`
	Platform  string `json:"platform"`
	Version   string `json:"version"`
	Kernel    string `json:"kernel"`
	Arch      string `json:"arch"`
	Uptime    uint64 `json:"uptime_seconds"`
	BootTime  uint64 `json:"boot_time"`
}

// Info is the aggregate system information structure.
type Info struct {
	CPU      CPUInfo             `json:"cpu"`
	Memory   MemoryInfo          `json:"memory"`
	Disks    []DiskInfo          `json:"disks"`
	Network  []NetworkInterface  `json:"network"`
	OS       OSInfo              `json:"os"`
	Go       GoInfo              `json:"go"`
}

// GoInfo holds Go runtime information.
type GoInfo struct {
	Version string `json:"version"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
}

// Gather collects all system information and returns it in a single Info struct.
func Gather() (*Info, error) {
	info := &Info{}

	if err := gatherCPU(info); err != nil {
		return nil, fmt.Errorf("cpu: %w", err)
	}
	if err := gatherMemory(info); err != nil {
		return nil, fmt.Errorf("memory: %w", err)
	}
	if err := gatherDisks(info); err != nil {
		return nil, fmt.Errorf("disk: %w", err)
	}
	if err := gatherNetwork(info); err != nil {
		return nil, fmt.Errorf("network: %w", err)
	}
	if err := gatherOS(info); err != nil {
		return nil, fmt.Errorf("os: %w", err)
	}
	gatherGo(info)

	return info, nil
}

func gatherCPU(info *Info) error {
	infos, err := cpu.Info()
	if err != nil {
		return err
	}

	if len(infos) > 0 {
		info.CPU.Model = infos[0].ModelName
		info.CPU.Frequency = infos[0].Mhz
	}

	info.CPU.Sockets = len(infos)

	cores, err := cpu.Counts(false)
	if err != nil {
		return err
	}
	info.CPU.Cores = cores

	usage, err := cpu.Percent(time.Second, true)
	if err != nil {
		return err
	}
	info.CPU.Usage = usage

	return nil
}

func gatherMemory(info *Info) error {
	v, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	info.Memory.Total = v.Total
	info.Memory.Used = v.Used
	info.Memory.Available = v.Available
	info.Memory.Usage = v.UsedPercent

	return nil
}

func gatherDisks(info *Info) error {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return err
	}

	for _, part := range partitions {
		usage, err := disk.Usage(part.Mountpoint)
		if err != nil {
			continue
		}
		info.Disks = append(info.Disks, DiskInfo{
			Device:       part.Device,
			Mountpoint:   part.Mountpoint,
			Fstype:       part.Fstype,
			Total:        usage.Total,
			Used:         usage.Used,
			Free:         usage.Free,
			UsagePercent: usage.UsedPercent,
		})
	}

	return nil
}

func gatherNetwork(info *Info) error {
	interfaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	for _, iface := range interfaces {
		ni := NetworkInterface{
			Name:     iface.Name,
			MTU:      iface.MTU,
			Hardware: iface.HardwareAddr,
		}

		for _, addr := range iface.Addrs {
			ni.IPs = append(ni.IPs, addr.Addr)
		}

		counters, err := net.IOCounters(true)
		if err == nil {
			for _, c := range counters {
				if c.Name == iface.Name {
					ni.BytesSent = c.BytesSent
					ni.BytesRecv = c.BytesRecv
					break
				}
			}
		}

		info.Network = append(info.Network, ni)
	}

	return nil
}

func gatherOS(info *Info) error {
	h, err := host.Info()
	if err != nil {
		return err
	}

	info.OS = OSInfo{
		Hostname: h.Hostname,
		OS:       h.OS,
		Platform: h.Platform,
		Version:  h.PlatformVersion,
		Kernel:   h.KernelVersion,
		Arch:     h.KernelArch,
		Uptime:   h.Uptime,
		BootTime: h.BootTime,
	}

	return nil
}

func gatherGo(info *Info) {
	info.Go = GoInfo{
		Version: runtime.Version(),
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
	}
}