// Package sysinfo provides human-readable formatting of system information.
package sysinfo

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FormatJSON returns the system info as indented JSON.
func (i *Info) FormatJSON() (string, error) {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// FormatText returns the system info as a human-readable text report.
func (i *Info) FormatText() string {
	var b strings.Builder

	b.WriteString("=== System Information ===\n\n")

	// OS
	b.WriteString("-- OS --\n")
	b.WriteString(fmt.Sprintf("  Hostname: %s\n", i.OS.Hostname))
	b.WriteString(fmt.Sprintf("  OS: %s %s (%s)\n", i.OS.OS, i.OS.Platform, i.OS.Version))
	b.WriteString(fmt.Sprintf("  Kernel: %s\n", i.OS.Kernel))
	b.WriteString(fmt.Sprintf("  Arch: %s\n", i.OS.Arch))
	b.WriteString(fmt.Sprintf("  Uptime: %s\n", formatUptime(i.OS.Uptime)))
	b.WriteString("\n")

	// CPU
	b.WriteString("-- CPU --\n")
	b.WriteString(fmt.Sprintf("  Model: %s\n", i.CPU.Model))
	b.WriteString(fmt.Sprintf("  Cores: %d\n", i.CPU.Cores))
	b.WriteString(fmt.Sprintf("  Sockets: %d\n", i.CPU.Sockets))
	b.WriteString(fmt.Sprintf("  Frequency: %.0f MHz\n", i.CPU.Frequency))
	if len(i.CPU.Usage) > 0 {
		b.WriteString("  Usage per core:\n")
		for idx, pct := range i.CPU.Usage {
			b.WriteString(fmt.Sprintf("    core %d: %.1f%%\n", idx, pct))
		}
	}
	b.WriteString("\n")

	// Memory
	b.WriteString("-- Memory --\n")
	b.WriteString(fmt.Sprintf("  Total: %s\n", formatBytes(i.Memory.Total)))
	b.WriteString(fmt.Sprintf("  Used: %s (%.1f%%)\n", formatBytes(i.Memory.Used), i.Memory.Usage))
	b.WriteString(fmt.Sprintf("  Available: %s\n", formatBytes(i.Memory.Available)))
	b.WriteString("\n")

	// Disks
	b.WriteString("-- Disks --\n")
	for _, d := range i.Disks {
		b.WriteString(fmt.Sprintf("  %s on %s (%s)\n", d.Device, d.Mountpoint, d.Fstype))
		b.WriteString(fmt.Sprintf("    Total: %s\n", formatBytes(d.Total)))
		b.WriteString(fmt.Sprintf("    Used: %s (%.1f%%)\n", formatBytes(d.Used), d.UsagePercent))
		b.WriteString(fmt.Sprintf("    Free: %s\n", formatBytes(d.Free)))
	}
	b.WriteString("\n")

	// Network
	b.WriteString("-- Network --\n")
	for _, n := range i.Network {
		b.WriteString(fmt.Sprintf("  %s (MTU %d)\n", n.Name, n.MTU))
		if n.Hardware != "" {
			b.WriteString(fmt.Sprintf("    MAC: %s\n", n.Hardware))
		}
		if len(n.IPs) > 0 {
			b.WriteString(fmt.Sprintf("    IPs: %s\n", strings.Join(n.IPs, ", ")))
		}
		b.WriteString(fmt.Sprintf("    RX: %s\n", formatBytes(n.BytesRecv)))
		b.WriteString(fmt.Sprintf("    TX: %s\n", formatBytes(n.BytesSent)))
	}
	b.WriteString("\n")

	// Go runtime
	b.WriteString("-- Go Runtime --\n")
	b.WriteString(fmt.Sprintf("  Version: %s\n", i.Go.Version))
	b.WriteString(fmt.Sprintf("  OS/Arch: %s/%s\n", i.Go.OS, i.Go.Arch))

	return b.String()
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatUptime(seconds uint64) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%dm %ds", seconds/60, seconds%60)
	}
	if seconds < 86400 {
		return fmt.Sprintf("%dh %dm", seconds/3600, (seconds%3600)/60)
	}
	return fmt.Sprintf("%dd %dh", seconds/86400, (seconds%86400)/3600)
}