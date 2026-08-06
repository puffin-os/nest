package quadlet

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

var (
	formTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("70")).
			Padding(0, 2)

	formLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("36")).
			Bold(true)

	formHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	formErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("203")).
			Bold(true)

	formSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("36")).
				Bold(true)

	formOptionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

// fieldKind distinguishes text input fields from selection fields.
type fieldKind int

const (
	kindText fieldKind = iota
	kindSelect
)

// fieldDef describes a single field in the create wizard.
type fieldDef struct {
	label    string
	kind     fieldKind
	placeholder string
	options  []string // for kindSelect
	defaultIdx int    // for kindSelect
}

// fieldIndex constants
const (
	idxName = iota
	idxImage
	idxDesc
	idxVolumes
	idxPorts
	idxEnv
	idxNetwork
	idxRestart
	idxAutoUpdate
	numFields
)

var fieldDefs = []fieldDef{
	{label: "Name", kind: kindText, placeholder: "myapp"},
	{label: "Image", kind: kindText, placeholder: "docker.io/library/nginx:latest"},
	{label: "Description", kind: kindText, placeholder: "My containerized app"},
	{label: "Volumes (optional, comma-separated)", kind: kindText, placeholder: "data:/data:Z, /host:/container:Z"},
	{label: "Ports (optional, comma-separated)", kind: kindText, placeholder: "8080:80, 8443:443"},
	{label: "Environment (optional, comma-separated)", kind: kindText, placeholder: "DEBUG=true, LOG=info"},
	{label: "Network", kind: kindSelect, options: []string{"host", "bridge", "none", "slirp4netns"}, defaultIdx: 0},
	{label: "Restart Policy", kind: kindSelect, options: []string{"always", "on-failure", "no", "on-abnormal", "on-watchdog"}, defaultIdx: 0},
	{label: "Auto-update from registry?", kind: kindSelect, options: []string{"no", "yes"}, defaultIdx: 0},
}

// createModel is the bubbletea model for the create wizard.
type createModel struct {
	step        int
	textInputs  []textinput.Model
	selectIdxs  []int
	done        bool
	result      *QuadletSpec
	err         string
}

// RunCreateForm launches an interactive bubbletea form for creating a quadlet.
func RunCreateForm() (*QuadletSpec, error) {
	m := newCreateModel()
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}
	if m, ok := finalModel.(createModel); ok && m.result != nil {
		return m.result, nil
	}
	return nil, nil
}

func newCreateModel() createModel {
	m := createModel{
		result:     &QuadletSpec{Restart: "always"},
		textInputs: make([]textinput.Model, numFields),
		selectIdxs: make([]int, numFields),
	}

	for i, fd := range fieldDefs {
		if fd.kind == kindText {
			ti := textinput.New()
			ti.Placeholder = fd.placeholder
			ti.CharLimit = 500
			ti.Width = 60
			m.textInputs[i] = ti
		} else {
			m.selectIdxs[i] = fd.defaultIdx
		}
	}

	// Focus first text field
	for i, fd := range fieldDefs {
		if fd.kind == kindText {
			m.textInputs[i].Focus()
			break
		}
	}

	return m
}

func (m createModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m createModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		case "tab", "shift+tab":
			if msg.String() == "shift+tab" {
				m.prevField()
			} else {
				m.nextField()
			}
			m.err = ""
			return m, nil
		case "enter":
			if m.step == numFields-1 {
				if err := m.validate(); err != "" {
					m.err = err
					return m, nil
				}
				m.collectResult()
				m.done = true
				return m, tea.Quit
			}
			m.nextField()
			m.err = ""
			return m, nil
		case "up", "k":
			if fieldDefs[m.step].kind == kindSelect {
				if m.selectIdxs[m.step] > 0 {
					m.selectIdxs[m.step]--
				}
				return m, nil
			}
		case "down", "j":
			if fieldDefs[m.step].kind == kindSelect {
				if m.selectIdxs[m.step] < len(fieldDefs[m.step].options)-1 {
					m.selectIdxs[m.step]++
				}
				return m, nil
			}
		}
	}

	// Only forward keystrokes to text inputs
	if fieldDefs[m.step].kind == kindText {
		var cmd tea.Cmd
		m.textInputs[m.step], cmd = m.textInputs[m.step].Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *createModel) nextField() {
	if fieldDefs[m.step].kind == kindText {
		m.textInputs[m.step].Blur()
	}
	if m.step < numFields-1 {
		m.step++
		if fieldDefs[m.step].kind == kindText {
			m.textInputs[m.step].Focus()
		}
	}
}

func (m *createModel) prevField() {
	if fieldDefs[m.step].kind == kindText {
		m.textInputs[m.step].Blur()
	}
	if m.step > 0 {
		m.step--
		if fieldDefs[m.step].kind == kindText {
			m.textInputs[m.step].Focus()
		}
	}
}

func (m createModel) validate() string {
	if m.textInputs[idxName].Value() == "" {
		return "Name is required"
	}
	if m.textInputs[idxImage].Value() == "" {
		return "Image is required"
	}
	return ""
}

func (m *createModel) collectResult() {
	r := m.result
	r.Name = m.textInputs[idxName].Value()
	r.Image = m.textInputs[idxImage].Value()
	r.Description = m.textInputs[idxDesc].Value()
	if r.Description == "" {
		r.Description = r.Name
	}

	r.Volumes = parseCSV(m.textInputs[idxVolumes].Value())
	r.Ports = parseCSV(m.textInputs[idxPorts].Value())
	r.Environment = parseCSV(m.textInputs[idxEnv].Value())

	r.Network = fieldDefs[idxNetwork].options[m.selectIdxs[idxNetwork]]
	r.Restart = fieldDefs[idxRestart].options[m.selectIdxs[idxRestart]]
	r.AutoUpdate = fieldDefs[idxAutoUpdate].options[m.selectIdxs[idxAutoUpdate]] == "yes"
}

// parseCSV splits a comma-separated string, trimming whitespace.
func parseCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var result []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func (m createModel) View() string {
	if m.done {
		return ""
	}

	var b strings.Builder
	b.WriteString(formTitleStyle.Render(" Create Quadlet "))
	b.WriteString("\n\n")

	for i, fd := range fieldDefs {
		active := i == m.step
		cursor := "  "
		if active {
			cursor = "▸ "
		}
		label := formLabelStyle.Render(fd.label)
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, label))

		if fd.kind == kindText {
			b.WriteString(fmt.Sprintf("    %s\n", m.textInputs[i].View()))
		} else {
			// Render selection options
			selIdx := m.selectIdxs[i]
			for j, opt := range fd.options {
				marker := "○"
				style := formOptionStyle
				if j == selIdx {
					marker = "●"
					style = formSelectedStyle
				}
				b.WriteString(fmt.Sprintf("    %s %s\n", marker, style.Render(opt)))
			}
		}

		if i < numFields-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	if fieldDefs[m.step].kind == kindSelect {
		b.WriteString(formHintStyle.Render("  Tab to next field · ↑/↓ to select · Enter to submit · Esc to cancel"))
	} else {
		b.WriteString(formHintStyle.Render("  Tab to next field · Shift+Tab for previous · Enter to submit · Esc to cancel"))
	}
	b.WriteString("\n")

	if m.err != "" {
		b.WriteString("\n")
		b.WriteString(formErrorStyle.Render(fmt.Sprintf("  Error: %s", m.err)))
		b.WriteString("\n")
	}

	return b.String()
}