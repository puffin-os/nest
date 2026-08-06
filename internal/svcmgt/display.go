package svcmgt

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	svcTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("70")).
			Padding(0, 2)

	svcHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("70")).
			Padding(0, 1)

	svcCellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	stateColors = map[ServiceState]lipgloss.Color{
		StateActive:     lipgloss.Color("36"),
		StateInactive:   lipgloss.Color("241"),
		StateFailed:     lipgloss.Color("203"),
		StateActivating: lipgloss.Color("215"),
		StateUnknown:    lipgloss.Color("241"),
	}
)

// FormatServiceList renders a table of all systemd services.
func FormatServiceList(services []Service) string {
	headers := []string{"UNIT", "LOAD", "ACTIVE", "SUB", "ENABLED", "DESCRIPTION"}

	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = lipgloss.Width(h)
	}

	rows := make([][]string, len(services))
	for i, svc := range services {
		enabled := "no"
		if svc.Enabled {
			enabled = "yes"
		}
		desc := svc.Description
		if len(desc) > 50 {
			desc = desc[:49] + "…"
		}
		row := []string{
			truncateSvc(svc.Unit, 40),
			svc.LoadState,
			string(svc.ActiveState),
			svc.SubState,
			enabled,
			desc,
		}
		for j, cell := range row {
			if w := lipgloss.Width(cell); w > colWidths[j] {
				colWidths[j] = w
			}
		}
		rows[i] = row
	}

	var b strings.Builder
	b.WriteString(svcTitleStyle.Render(" Systemd Services "))
	b.WriteString("\n\n")

	// Header row
	var headerCells []string
	for i, h := range headers {
		headerCells = append(headerCells, svcHeaderStyle.Render(padRightSvc(h, colWidths[i])))
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, headerCells...))
	b.WriteString("\n")

	// Separator
	var sepCells []string
	for _, w := range colWidths {
		sepCells = append(sepCells, strings.Repeat("─", w+2))
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, sepCells...))
	b.WriteString("\n")

	// Data rows
	for _, row := range rows {
		var cells []string
		for j, cell := range row {
			style := svcCellStyle
			switch j {
			case 2: // Active column
				color := stateColors[ServiceState(cell)]
				style = style.Foreground(color).Bold(true)
			case 4: // Enabled column
				if cell == "yes" {
					style = style.Foreground(lipgloss.Color("36"))
				} else {
					style = style.Foreground(lipgloss.Color("241"))
				}
			}
			padded := padRightSvc(cell, colWidths[j])
			cells = append(cells, style.Render(padded))
		}
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, cells...))
		b.WriteString("\n")
	}

	return b.String()
}

// FormatServiceDetail renders a detailed view for a single service.
func FormatServiceDetail(detail *ServiceDetail) string {
	var b strings.Builder

	b.WriteString(svcTitleStyle.Render(fmt.Sprintf(" Service: %s ", detail.Unit)))
	b.WriteString("\n\n")

	// General section
	b.WriteString(svcHeaderStyle.Render("General"))
	b.WriteString("\n")
	enabled := "no"
	if detail.Enabled {
		enabled = "yes"
	}
	statusRows := [][2]string{
		{"Unit", detail.Unit},
		{"Description", detail.Description},
		{"Load State", detail.LoadState},
		{"Active State", string(detail.ActiveState)},
		{"Sub State", detail.SubState},
		{"Enabled", enabled},
		{"Type", detail.Type},
		{"PID", fmt.Sprintf("%d", detail.PID)},
	}
	b.WriteString(renderSvcStatusRows(statusRows))

	// Execution section
	b.WriteString("\n")
	b.WriteString(svcHeaderStyle.Render("Execution"))
	b.WriteString("\n")
	execRows := [][2]string{
		{"Exec Start", detail.ExecStart},
		{"Fragment Path", detail.FragmentPath},
		{"Main PID", fmt.Sprintf("%d", detail.ExecMainPID)},
		{"Started", func() string {
			if detail.Timestamp == "" {
				return "-"
			}
			return detail.Timestamp
		}()},
	}
	b.WriteString(renderSvcStatusRows(execRows))

	// Resources section
	b.WriteString("\n")
	b.WriteString(svcHeaderStyle.Render("Resources"))
	b.WriteString("\n")
	resRows := [][2]string{
		{"Memory Current", formatSvcBytes(detail.MemoryCurrent)},
		{"Memory Peak", formatSvcBytes(detail.MemoryPeak)},
		{"CPU Usage", fmt.Sprintf("%.2f s", float64(detail.CPUUsageNS)/1e9)},
	}
	b.WriteString(renderSvcStatusRows(resRows))

	// Environment section (if any)
	if len(detail.Environment) > 0 {
		b.WriteString("\n")
		b.WriteString(svcHeaderStyle.Render("Environment"))
		b.WriteString("\n")
		for k, v := range detail.Environment {
			b.WriteString(fmt.Sprintf("  %s=%s\n", k, v))
		}
	}

	return b.String()
}

func renderSvcStatusRows(rows [][2]string) string {
	var b strings.Builder
	maxLabel := 0
	for _, r := range rows {
		if len(r[0]) > maxLabel {
			maxLabel = len(r[0])
		}
	}
	for _, r := range rows {
		label := lipgloss.NewStyle().Foreground(lipgloss.Color("70")).Bold(true).Render(padRightSvc(r[0], maxLabel))
		value := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(r[1])
		b.WriteString(fmt.Sprintf("  %s  %s\n", label, value))
	}
	return b.String()
}

func padRightSvc(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func truncateSvc(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func formatSvcBytes(bytes uint64) string {
	if bytes == 0 {
		return "-"
	}
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