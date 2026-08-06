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
			Padding(0, 1).
			Background(lipgloss.Color("236"))

	labelStyle = lipgloss.NewStyle().
			Foreground(cSecondary).
			Bold(true)

	valueStyle = lipgloss.NewStyle().
			Foreground(cValue)

	barStyle = lipgloss.NewStyle().
			Foreground(cAccent)

	outerBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cBorder).
			Padding(1, 2)
)

// section represents a titled block of key-value rows.
type section struct {
	title string
	rows  [][2]string
}

// FormatStyled returns the system info as a styled terminal output using lipgloss.
// The layout uses a 2-column grid inside a single bordered frame.
func (i *Info) FormatStyled() string {
	left, right := i.buildColumns()

	// Compute column width, capped to keep the frame reasonable.
	const maxColWidth = 60
	colWidth := 0
	for _, s := range append(left, right...) {
		w := ansi.StringWidth(renderSection(s, 0))
		if w > colWidth {
			colWidth = w
		}
	}
	if colWidth > maxColWidth {
		colWidth = maxColWidth
	}

	// Render each column's sections at a fixed width so they align.
	maxSections := len(left)
	if len(right) > maxSections {
		maxSections = len(right)
	}

	var rowBlocks []string
	for idx := 0; idx < maxSections; idx++ {
		var l, r string
		if idx < len(left) {
			l = renderSection(left[idx], colWidth)
		}
		if idx < len(right) {
			r = renderSection(right[idx], colWidth)
		}
		if l == "" {
			l = strings.Repeat(" ", colWidth)
		}
		if r == "" {
			r = strings.Repeat(" ", colWidth)
		}
		rowBlocks = append(rowBlocks, lipgloss.JoinHorizontal(lipgloss.Top, l, "  ", r))
	}

	body := strings.Join(rowBlocks, "\n")
	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(" System Information "),
		"",
		body,
	)

	return outerBoxStyle.Render(content)
}

// buildColumns distributes sections into two columns.
func (i *Info) buildColumns() (left, right []section) {
	// OS section
	left = append(left, section{title: "OS", rows: [][2]string{
		{"Hostname", i.OS.Hostname},
		{"OS", fmt.Sprintf("%s %s (%s)", i.OS.OS, i.OS.Platform, i.OS.Version)},
		{"Kernel", i.OS.Kernel},
		{"Arch", i.OS.Arch},
		{"Uptime", formatUptime(i.OS.Uptime)},
	}})

	// CPU section
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
		cpuRows = append(cpuRows, [2]string{"Per-Core", progressBar(avg, 20)})
	}
	right = append(right, section{title: "CPU", rows: cpuRows})

	// Memory section
	left = append(left, section{title: "Memory", rows: [][2]string{
		{"Total", formatBytes(i.Memory.Total)},
		{"Used", fmt.Sprintf("%s (%.1f%%)", formatBytes(i.Memory.Used), i.Memory.Usage)},
		{"Available", formatBytes(i.Memory.Available)},
		{"Usage", progressBar(i.Memory.Usage, 20)},
	}})

	// Go runtime section
	right = append(right, section{title: "Go Runtime", rows: [][2]string{
		{"Version", i.Go.Version},
		{"OS/Arch", fmt.Sprintf("%s/%s", i.Go.OS, i.Go.Arch)},
	}})

	// Disks: alternate between columns
	for idx, d := range i.Disks {
		s := section{title: "Disk: " + d.Mountpoint, rows: [][2]string{
			{"Device", d.Device},
			{"Type", d.Fstype},
			{"Total", formatBytes(d.Total)},
			{"Used", fmt.Sprintf("%s (%.1f%%)", formatBytes(d.Used), d.UsagePercent)},
			{"Free", formatBytes(d.Free)},
		}}
		if idx%2 == 0 {
			left = append(left, s)
		} else {
			right = append(right, s)
		}
	}

	// Network: alternate between columns
	for idx, n := range i.Network {
		ips := strings.Join(n.IPs, ", ")
		if ips == "" {
			ips = "-"
		}
		mac := n.Hardware
		if mac == "" {
			mac = "-"
		}
		s := section{title: "Net: " + n.Name, rows: [][2]string{
			{"MTU", fmt.Sprintf("%d", n.MTU)},
			{"MAC", mac},
			{"IPs", ips},
			{"RX", formatBytes(n.BytesRecv)},
			{"TX", formatBytes(n.BytesSent)},
		}}
		if idx%2 == 0 {
			left = append(left, s)
		} else {
			right = append(right, s)
		}
	}

	return left, right
}

// renderSection renders a section title and its key-value rows at a target width.
func renderSection(s section, width int) string {
	var lines []string

	// Section header bar
	lines = append(lines, sectionStyle.Render(s.title))

	// Key-value rows
	maxLabel := 0
	for _, r := range s.rows {
		if len(r[0]) > maxLabel {
			maxLabel = len(r[0])
		}
	}
	// Available width for the value column.
	valueWidth := 0
	if width > 0 {
		valueWidth = width - maxLabel - 4 // 2 indent + 2 separator
		if valueWidth < 10 {
			valueWidth = 10
		}
	}
	for _, r := range s.rows {
		label := labelStyle.Render(padRight(r[0], maxLabel))
		val := r[1]
		if valueWidth > 0 && ansi.StringWidth(val) > valueWidth {
			val = ansi.Truncate(val, valueWidth-1, "…")
		}
		value := valueStyle.Render(val)
		lines = append(lines, fmt.Sprintf("  %s  %s", label, value))
	}

	body := strings.Join(lines, "\n")

	if width > 0 {
		return lipgloss.NewStyle().Width(width).Render(body)
	}
	return body
}

// progressBar renders a text-based progress bar at the given width.
func progressBar(pct float64, width int) string {
	if width <= 0 {
		width = 30
	}
	filled := int((pct / 100.0) * float64(width))
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