package quadlet

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	quadTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("70")).
			Padding(0, 2)

	quadHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("70")).
				Padding(0, 1)

	quadCellStyle = lipgloss.NewStyle().
			Padding(0, 1)
)

// FormatQuadletList renders a table of all quadlets.
func FormatQuadletList(quadlets []Quadlet) string {
	headers := []string{"NAME", "TYPE", "IMAGE", "ACTIVE", "ENABLED", "SCOPE"}

	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = lipgloss.Width(h)
	}

	rows := make([][]string, len(quadlets))
	for i, q := range quadlets {
		image := q.Image
		if image == "" {
			image = "-"
		}
		if len(image) > 40 {
			image = image[:39] + "…"
		}
		active := q.ActiveState
		if active == "" {
			active = "-"
		}
		enabled := "no"
		if q.Enabled {
			enabled = "yes"
		}
		scope := string(q.Scope)
		row := []string{
			truncateStr(q.Name, 30),
			string(q.Type),
			image,
			active,
			enabled,
			scope,
		}
		for j, cell := range row {
			if w := lipgloss.Width(cell); w > colWidths[j] {
				colWidths[j] = w
			}
		}
		rows[i] = row
	}

	var b strings.Builder
	b.WriteString(quadTitleStyle.Render(" Quadlets "))
	b.WriteString("\n\n")

	// Header
	var headerCells []string
	for i, h := range headers {
		headerCells = append(headerCells, quadHeaderStyle.Render(padRight(h, colWidths[i])))
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

	// Rows
	for _, row := range rows {
		var cells []string
		for j, cell := range row {
			style := quadCellStyle
			switch j {
			case 3: // Active column
				switch cell {
				case "active":
					cell = lipgloss.NewStyle().Foreground(lipgloss.Color("36")).Bold(true).Render(cell)
				case "failed":
					cell = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true).Render(cell)
				case "inactive", "-":
					cell = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(cell)
				}
			case 4: // Enabled column
				if cell == "yes" {
					cell = lipgloss.NewStyle().Foreground(lipgloss.Color("36")).Render(cell)
				} else {
					cell = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(cell)
				}
			case 5: // Scope column
				if cell == "user" {
					cell = lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Render(cell)
				} else {
					cell = lipgloss.NewStyle().Foreground(lipgloss.Color("213")).Render(cell)
				}
			}
			padded := padRight(cell, colWidths[j])
			cells = append(cells, style.Render(padded))
		}
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, cells...))
		b.WriteString("\n")
	}

	if len(quadlets) == 0 {
		b.WriteString(formHintStyle.Render("  (no quadlets found)"))
		b.WriteString("\n")
	}

	return b.String()
}

// FormatQuadletDetail renders a detailed view for a single quadlet.
func FormatQuadletDetail(q *Quadlet) string {
	var b strings.Builder

	b.WriteString(quadTitleStyle.Render(fmt.Sprintf(" Quadlet: %s ", q.Name)))
	b.WriteString("\n\n")

	b.WriteString(quadHeaderStyle.Render("General"))
	b.WriteString("\n")
	active := q.ActiveState
	if active == "" {
		active = "(not loaded)"
	}
	enabled := "no"
	if q.Enabled {
		enabled = "yes"
	}
	rows := [][2]string{
		{"Name", q.Name},
		{"Type", string(q.Type)},
		{"Scope", string(q.Scope)},
		{"Unit File", q.UnitFile},
		{"Image", func() string {
			if q.Image == "" {
				return "-"
			}
			return q.Image
		}()},
		{"Active State", active},
		{"Sub State", func() string {
			if q.SubState == "" {
				return "-"
			}
			return q.SubState
		}()},
		{"Enabled", enabled},
	}
	for _, r := range rows {
		label := lipgloss.NewStyle().Foreground(lipgloss.Color("70")).Bold(true).Render(padRight(r[0], 12))
		value := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(r[1])
		b.WriteString(fmt.Sprintf("  %s  %s\n", label, value))
	}

	return b.String()
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}