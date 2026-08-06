package netmgmt

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	listTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("99")).
			Padding(0, 2)

	listHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			Padding(0, 1)

	listCellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	listUpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("36")).
			Bold(true)

	listDownStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("203"))

	typeColors = map[InterfaceType]lipgloss.Color{
		TypePhysical: lipgloss.Color("36"),
		TypeVLAN:     lipgloss.Color("213"),
		TypeBond:     lipgloss.Color("99"),
		TypeBridge:   lipgloss.Color("215"),
		TypeLoopback: lipgloss.Color("241"),
		TypeVirtual:  lipgloss.Color("117"),
		TypeUnknown:  lipgloss.Color("241"),
	}
)

// FormatList renders a table of all network interfaces.
func FormatList(ifaces []Interface) string {
	headers := []string{"NAME", "TYPE", "STATE", "MTU", "MAC", "IP ADDRESSES", "RX", "TX"}

	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = lipgloss.Width(h)
	}

	rows := make([][]string, len(ifaces))
	for i, iface := range ifaces {
		state := "DOWN"
		if iface.Up {
			state = "UP"
		}
		ips := strings.Join(iface.IPAddresses, ", ")
		if ips == "" {
			ips = "-"
		}
		if len(ips) > 50 {
			ips = ips[:49] + "…"
		}
		mac := iface.MAC
		if mac == "" {
			mac = "-"
		}
		name := iface.Name
		if len(name) > 14 {
			name = name[:13] + "…"
		}
		row := []string{
			name,
			string(iface.Type),
			state,
			fmt.Sprintf("%d", iface.MTU),
			mac,
			ips,
			formatBytes(iface.RxBytes),
			formatBytes(iface.TxBytes),
		}
		for j, cell := range row {
			if w := lipgloss.Width(cell); w > colWidths[j] {
				colWidths[j] = w
			}
		}
		rows[i] = row
	}

	var b strings.Builder
	b.WriteString(listTitleStyle.Render(" Network Interfaces "))
	b.WriteString("\n\n")

	// Header row
	var headerCells []string
	for i, h := range headers {
		headerCells = append(headerCells, listHeaderStyle.Render(padRightStr(h, colWidths[i])))
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
			style := listCellStyle
			// Special styling for certain columns
			switch j {
			case 1: // Type column
				color := typeColors[InterfaceType(cell)]
				style = style.Foreground(color)
			case 2: // State column
				if cell == "UP" {
					cell = listUpStyle.Render(cell)
				} else {
					cell = listDownStyle.Render(cell)
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

// FormatStatus renders a detailed status view for a single interface.
func FormatStatus(iface Interface, bond *BondInfo, vlan *VLANInfo) string {
	var b strings.Builder

	b.WriteString(listTitleStyle.Render(fmt.Sprintf(" Interface: %s ", iface.Name)))
	b.WriteString("\n\n")

	// General section
	b.WriteString(listHeaderStyle.Render("General"))
	b.WriteString("\n")
	statusRows := [][2]string{
		{"Name", iface.Name},
		{"Type", string(iface.Type)},
		{"State", func() string {
			if iface.Up {
				return "UP"
			}
			return "DOWN"
		}()},
		{"MTU", fmt.Sprintf("%d", iface.MTU)},
		{"MAC", func() string {
			if iface.MAC == "" {
				return "-"
			}
			return iface.MAC
		}()},
		{"Speed", func() string {
			if iface.Speed == "" {
				return "-"
			}
			return iface.Speed
		}()},
		{"Duplex", func() string {
			if iface.Duplex == "" {
				return "-"
			}
			return iface.Duplex
		}()},
	}
	b.WriteString(renderStatusRows(statusRows))

	// Addresses section
	b.WriteString("\n")
	b.WriteString(listHeaderStyle.Render("Addresses"))
	b.WriteString("\n")
	if len(iface.IPAddresses) == 0 {
		b.WriteString("  (none)\n")
	} else {
		for _, ip := range iface.IPAddresses {
			b.WriteString(fmt.Sprintf("  %s\n", ip))
		}
	}

	// Statistics section
	b.WriteString("\n")
	b.WriteString(listHeaderStyle.Render("Statistics"))
	b.WriteString("\n")
	statRows := [][2]string{
		{"RX Bytes", formatBytes(iface.RxBytes)},
		{"TX Bytes", formatBytes(iface.TxBytes)},
		{"RX Packets", fmt.Sprintf("%d", iface.RxPackets)},
		{"TX Packets", fmt.Sprintf("%d", iface.TxPackets)},
		{"RX Errors", fmt.Sprintf("%d", iface.RxErrors)},
		{"TX Errors", fmt.Sprintf("%d", iface.TxErrors)},
	}
	b.WriteString(renderStatusRows(statRows))

	// Bond details if applicable
	if bond != nil {
		b.WriteString("\n")
		b.WriteString(listHeaderStyle.Render("Bond"))
		b.WriteString("\n")
		bondRows := [][2]string{
			{"Mode", bond.Mode},
			{"Active Slave", bond.Active},
			{"Primary", bond.Primary},
			{"Slaves", strings.Join(bond.Slaves, ", ")},
		}
		b.WriteString(renderStatusRows(bondRows))
	}

	// VLAN details if applicable
	if vlan != nil {
		b.WriteString("\n")
		b.WriteString(listHeaderStyle.Render("VLAN"))
		b.WriteString("\n")
		vlanRows := [][2]string{
			{"Parent", vlan.ParentIF},
			{"VLAN ID", fmt.Sprintf("%d", vlan.VLANID)},
		}
		b.WriteString(renderStatusRows(vlanRows))
	}

	return b.String()
}

func renderStatusRows(rows [][2]string) string {
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