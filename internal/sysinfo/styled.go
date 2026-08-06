// Package sysinfo provides styled output of system information using lipgloss.
package sysinfo

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Lipgloss color palette.
var (
	cPrimary   = lipgloss.Color("99")
	cSecondary = lipgloss.Color("36")
	cAccent    = lipgloss.Color("213")
	cMuted     = lipgloss.Color("241")
	cValue     = lipgloss.Color("252")
	cBorder    = lipgloss.Color("63")
)

// Style definitions.
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(cPrimary).
			Padding(0, 2)

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(cPrimary).
			MarginTop(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(cSecondary).
			Bold(true)

	valueStyle = lipgloss.NewStyle().
			Foreground(cValue)

	mutedStyle = lipgloss.NewStyle().
			Foreground(cMuted)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cBorder).
			Padding(1, 2)

	barStyle = lipgloss.NewStyle().
			Foreground(cAccent)
)

// FormatStyled returns the system info as a styled terminal output using lipgloss.
func (i *Info) FormatStyled() string {
	var sections []string

	// Title
	sections = append(sections, titleStyle.Render(" System Information "))

	// OS section
	sections = append(sections, sectionStyle.Render("OS"))
	sections = append(sections, kvBox([][2]string{
		{"Hostname", i.OS.Hostname},
		{"OS", fmt.Sprintf("%s %s (%s)", i.OS.OS, i.OS.Platform, i.OS.Version)},
		{"Kernel", i.OS.Kernel},
		{"Arch", i.OS.Arch},
		{"Uptime", formatUptime(i.OS.Uptime)},
	}))

	// CPU section
	sections = append(sections, sectionStyle.Render("CPU"))
	cpuRows := [][2]string{
		{"Model", i.CPU.Model},
		{"Cores", fmt.Sprintf("%d", i.CPU.Cores)},
		{"Sockets", fmt.Sprintf("%d", i.CPU.Sockets)},
		{"Frequency", fmt.Sprintf("%.0f MHz", i.CPU.Frequency)},
	}
	if len(i.CPU.Usage) > 0 {
		var totalUsage float64
		for _, u := range i.CPU.Usage {
			totalUsage += u
		}
		avg := totalUsage / float64(len(i.CPU.Usage))
		cpuRows = append(cpuRows, [2]string{"Avg Usage", fmt.Sprintf("%.1f%%", avg)})
		cpuRows = append(cpuRows, [2]string{"Per-Core", progressBar(avg)})
	}
	sections = append(sections, kvBox(cpuRows))

	// Memory section
	sections = append(sections, sectionStyle.Render("Memory"))
	sections = append(sections, kvBox([][2]string{
		{"Total", formatBytes(i.Memory.Total)},
		{"Used", fmt.Sprintf("%s (%.1f%%)", formatBytes(i.Memory.Used), i.Memory.Usage)},
		{"Available", formatBytes(i.Memory.Available)},
		{"Usage", progressBar(i.Memory.Usage)},
	}))

	// Disks section
	sections = append(sections, sectionStyle.Render("Disks"))
	for _, d := range i.Disks {
		sections = append(sections, kvBox([][2]string{
			{"Device", d.Device},
			{"Mount", d.Mountpoint},
			{"Type", d.Fstype},
			{"Total", formatBytes(d.Total)},
			{"Used", fmt.Sprintf("%s (%.1f%%)", formatBytes(d.Used), d.UsagePercent)},
			{"Free", formatBytes(d.Free)},
		}))
	}

	// Network section
	sections = append(sections, sectionStyle.Render("Network"))
	for _, n := range i.Network {
		ips := strings.Join(n.IPs, ", ")
		if ips == "" {
			ips = "-"
		}
		mac := n.Hardware
		if mac == "" {
			mac = "-"
		}
		sections = append(sections, kvBox([][2]string{
			{"Interface", n.Name},
			{"MTU", fmt.Sprintf("%d", n.MTU)},
			{"MAC", mac},
			{"IPs", ips},
			{"RX", formatBytes(n.BytesRecv)},
			{"TX", formatBytes(n.BytesSent)},
		}))
	}

	// Go runtime section
	sections = append(sections, sectionStyle.Render("Go Runtime"))
	sections = append(sections, kvBox([][2]string{
		{"Version", i.Go.Version},
		{"OS/Arch", fmt.Sprintf("%s/%s", i.Go.OS, i.Go.Arch)},
	}))

	// Join all sections vertically with left alignment.
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// kvBox renders a list of key-value pairs inside a bordered box.
func kvBox(rows [][2]string) string {
	var lines []string
	maxLabel := 0
	for _, r := range rows {
		if len(r[0]) > maxLabel {
			maxLabel = len(r[0])
		}
	}
	for _, r := range rows {
		label := labelStyle.Render(padRight(r[0], maxLabel))
		value := valueStyle.Render(r[1])
		lines = append(lines, fmt.Sprintf("%s  %s", label, value))
	}
	body := strings.Join(lines, "\n")
	return boxStyle.Render(body)
}

// progressBar renders a text-based progress bar.
func progressBar(pct float64) string {
	const width = 30
	filled := int((pct / 100.0) * width)
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("%s %.1f%%", barStyle.Render(bar), pct)
}

func padRight(s string, width int) string {
	if ansi.StringWidth(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-ansi.StringWidth(s))
}