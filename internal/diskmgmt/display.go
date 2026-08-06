package diskmgmt

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	diskTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("36")).
			Padding(0, 2)

	diskHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("36")).
				Padding(0, 1)

	diskCellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	diskUpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("36")).
			Bold(true)

	diskDownStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("203"))

	typeDiskColors = map[DiskType]lipgloss.Color{
		TypeDisk:       lipgloss.Color("36"),
		TypePart:       lipgloss.Color("117"),
		TypeLVM:        lipgloss.Color("213"),
		TypeCrypt:      lipgloss.Color("99"),
		TypeROM:        lipgloss.Color("241"),
		TypeLoop:       lipgloss.Color("241"),
		TypeUnknownDev: lipgloss.Color("241"),
	}
)

// FormatDiskList renders a table of all block devices.
func FormatDiskList(devices []BlockDevice) string {
	headers := []string{"PATH", "TYPE", "SIZE", "RO", "FSTYPE", "LABEL", "MOUNTPOINT"}

	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = lipgloss.Width(h)
	}

	rows := make([][]string, len(devices))
	for i, dev := range devices {
		mount := dev.Mountpoint
		if mount == "" {
			mount = "-"
		}
		fsType := dev.FSType
		if fsType == "" {
			fsType = "-"
		}
		label := dev.Label
		if label == "" {
			label = "-"
		}
		ro := "no"
		if dev.ReadOnly {
			ro = "yes"
		}
		row := []string{
			truncate(dev.Path, 20),
			string(dev.Type),
			FormatSizeBytes(dev.Size),
			ro,
			fsType,
			label,
			truncate(mount, 30),
		}
		for j, cell := range row {
			if w := lipgloss.Width(cell); w > colWidths[j] {
				colWidths[j] = w
			}
		}
		rows[i] = row
	}

	var b strings.Builder
	b.WriteString(diskTitleStyle.Render(" Block Devices "))
	b.WriteString("\n\n")

	// Header row
	var headerCells []string
	for i, h := range headers {
		headerCells = append(headerCells, diskHeaderStyle.Render(padRightStr(h, colWidths[i])))
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
			style := diskCellStyle
			switch j {
			case 1: // Type column
				color := typeDiskColors[DiskType(cell)]
				style = style.Foreground(color)
			case 3: // RO column
				if cell == "yes" {
					style = style.Foreground(lipgloss.Color("203"))
				} else {
					style = style.Foreground(lipgloss.Color("36"))
				}
			}
			padded := padRightStr(cell, colWidths[j])
			cells = append(cells, style.Render(padded))
		}
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, cells...))
		b.WriteString("\n")
	}

	return b.String()
}

// FormatDiskStatus renders a detailed status view for a single block device.
func FormatDiskStatus(dev BlockDevice) string {
	var b strings.Builder

	b.WriteString(diskTitleStyle.Render(fmt.Sprintf(" Device: %s ", dev.Path)))
	b.WriteString("\n\n")

	b.WriteString(diskHeaderStyle.Render("General"))
	b.WriteString("\n")
	ro := "no"
	if dev.ReadOnly {
		ro = "yes"
	}
	mount := dev.Mountpoint
	if mount == "" {
		mount = "(not mounted)"
	}
	statusRows := [][2]string{
		{"Path", dev.Path},
		{"Name", dev.Name},
		{"Type", string(dev.Type)},
		{"Size", FormatSizeBytes(dev.Size)},
		{"Read Only", ro},
		{"Filesystem", func() string {
			if dev.FSType == "" {
				return "-"
			}
			return dev.FSType
		}()},
		{"Label", func() string {
			if dev.Label == "" {
				return "-"
			}
			return dev.Label
		}()},
		{"UUID", func() string {
			if dev.UUID == "" {
				return "-"
			}
			return dev.UUID
		}()},
		{"Mountpoint", mount},
	}
	if dev.Model != "" {
		statusRows = append(statusRows, [2]string{"Model", dev.Model})
	}
	if dev.Vendor != "" {
		statusRows = append(statusRows, [2]string{"Vendor", dev.Vendor})
	}
	b.WriteString(renderDiskStatusRows(statusRows))

	// Children (partitions)
	if len(dev.Children) > 0 {
		b.WriteString("\n")
		b.WriteString(diskHeaderStyle.Render("Partitions"))
		b.WriteString("\n")
		for _, child := range dev.Children {
			mount := child.Mountpoint
			if mount == "" {
				mount = "-"
			}
			fs := child.FSType
			if fs == "" {
				fs = "-"
			}
			line := fmt.Sprintf("  %s  %s  %s  %s", child.Path, FormatSizeBytes(child.Size), fs, mount)
			b.WriteString(line + "\n")
		}
	}

	return b.String()
}

// FormatMountList renders a table of all mounted filesystems.
func FormatMountList(mounts []MountInfo) string {
	headers := []string{"SOURCE", "TARGET", "FSTYPE", "SIZE", "USED", "AVAIL", "USE%"}

	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = lipgloss.Width(h)
	}

	rows := make([][]string, len(mounts))
	for i, m := range mounts {
		row := []string{
			truncate(m.Source, 20),
			truncate(m.Target, 30),
			m.FSType,
			FormatSizeBytes(m.SizeBytes),
			FormatSizeBytes(m.UsedBytes),
			FormatSizeBytes(m.AvailBytes),
			fmt.Sprintf("%d%%", m.UsePercent),
		}
		for j, cell := range row {
			if w := lipgloss.Width(cell); w > colWidths[j] {
				colWidths[j] = w
			}
		}
		rows[i] = row
	}

	var b strings.Builder
	b.WriteString(diskTitleStyle.Render(" Mounted Filesystems "))
	b.WriteString("\n\n")

	var headerCells []string
	for i, h := range headers {
		headerCells = append(headerCells, diskHeaderStyle.Render(padRightStr(h, colWidths[i])))
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, headerCells...))
	b.WriteString("\n")

	var sepCells []string
	for _, w := range colWidths {
		sepCells = append(sepCells, strings.Repeat("─", w+2))
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, sepCells...))
	b.WriteString("\n")

	for _, row := range rows {
		var cells []string
		for j, cell := range row {
			style := diskCellStyle
			if j == 6 { // USE% column
				pct := row[6]
				pctNum := 0
				fmt.Sscanf(pct, "%d%%", &pctNum)
				if pctNum >= 90 {
					style = style.Foreground(lipgloss.Color("203")).Bold(true)
				} else if pctNum >= 75 {
					style = style.Foreground(lipgloss.Color("215"))
				} else {
					style = style.Foreground(lipgloss.Color("36"))
				}
			}
			padded := padRightStr(cell, colWidths[j])
			cells = append(cells, style.Render(padded))
		}
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, cells...))
		b.WriteString("\n")
	}

	return b.String()
}

func renderDiskStatusRows(rows [][2]string) string {
	var b strings.Builder
	maxLabel := 0
	for _, r := range rows {
		if len(r[0]) > maxLabel {
			maxLabel = len(r[0])
		}
	}
	for _, r := range rows {
		label := lipgloss.NewStyle().Foreground(lipgloss.Color("36")).Bold(true).Render(padRightStr(r[0], maxLabel))
		value := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(r[1])
		b.WriteString(fmt.Sprintf("  %s  %s\n", label, value))
	}
	return b.String()
}

func padRightStr(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}