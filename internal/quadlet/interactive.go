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
)

// field index constants
const (
	idxName = iota
	idxImage
	idxDesc
	idxVolumes
	idxPorts
	idxEnv
	idxRestart
	idxAutoUpdate
	numFields
)

// createModel is the bubbletea model for the create wizard.
type createModel struct {
	step   int
	inputs []textinput.Model
	done   bool
	result *QuadletSpec
	err    string
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
		result: &QuadletSpec{Restart: "always"},
	}

	m.inputs = make([]textinput.Model, numFields)

	m.inputs[idxName] = newField("myapp", "Quadlet name (used as systemd unit name)")
	m.inputs[idxImage] = newField("docker.io/library/nginx:latest", "Container image")
	m.inputs[idxDesc] = newField("My containerized app", "Description (for systemd unit)")
	m.inputs[idxVolumes] = newField("data:/data:Z", "Volumes (comma-separated, optional, e.g. data:/data, /host:/container:Z)")
	m.inputs[idxPorts] = newField("8080:80", "Ports (comma-separated, optional, e.g. 8080:80, 8443:443)")
	m.inputs[idxEnv] = newField("KEY=value", "Environment (comma-separated, optional, e.g. DEBUG=true,LOG=info)")
	m.inputs[idxRestart] = newField("always", "Restart policy (no, on-failure, always)")
	m.inputs[idxAutoUpdate] = newField("no", "Auto-update from registry? (yes/no)")

	m.inputs[0].Focus()
	return m
}

func newField(placeholder, _ string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 500
	ti.Width = 60
	return ti
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
		case "tab", "shift+tab", "enter":
			if msg.String() == "enter" && m.step == numFields-1 {
				if err := m.validate(); err != "" {
					m.err = err
					return m, nil
				}
				m.collectResult()
				m.done = true
				return m, tea.Quit
			}
			if msg.String() == "shift+tab" {
				m.prevInput()
			} else {
				m.nextInput()
			}
			m.err = ""
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.inputs[m.step], cmd = m.inputs[m.step].Update(msg)
	return m, cmd
}

func (m *createModel) nextInput() {
	m.inputs[m.step].Blur()
	if m.step < numFields-1 {
		m.step++
		m.inputs[m.step].Focus()
	}
}

func (m *createModel) prevInput() {
	m.inputs[m.step].Blur()
	if m.step > 0 {
		m.step--
		m.inputs[m.step].Focus()
	}
}

func (m createModel) validate() string {
	if m.inputs[idxName].Value() == "" {
		return "Name is required"
	}
	if m.inputs[idxImage].Value() == "" {
		return "Image is required"
	}
	return ""
}

func (m *createModel) collectResult() {
	r := m.result
	r.Name = m.inputs[idxName].Value()
	r.Image = m.inputs[idxImage].Value()
	r.Description = m.inputs[idxDesc].Value()
	if r.Description == "" {
		r.Description = r.Name
	}

	r.Volumes = parseCSV(m.inputs[idxVolumes].Value())
	r.Ports = parseCSV(m.inputs[idxPorts].Value())
	r.Environment = parseCSV(m.inputs[idxEnv].Value())

	restart := m.inputs[idxRestart].Value()
	if restart != "" {
		r.Restart = restart
	}
	if r.Restart == "" {
		r.Restart = "always"
	}

	r.AutoUpdate = strings.ToLower(m.inputs[idxAutoUpdate].Value()) == "yes"
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

	labels := []string{
		"Name",
		"Image",
		"Description",
		"Volumes (optional, comma-separated)",
		"Ports (optional, comma-separated)",
		"Environment (optional, comma-separated)",
		"Restart Policy",
		"Auto-update? (yes/no)",
	}

	var b strings.Builder
	b.WriteString(formTitleStyle.Render(" Create Quadlet "))
	b.WriteString("\n\n")

	for i, input := range m.inputs {
		label := formLabelStyle.Render(labels[i])
		if i == m.step {
			b.WriteString(fmt.Sprintf("  ▸ %s\n", label))
		} else {
			b.WriteString(fmt.Sprintf("    %s\n", label))
		}
		b.WriteString(fmt.Sprintf("    %s\n", input.View()))
		if i < numFields-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(formHintStyle.Render("  Tab to next field · Shift+Tab for previous · Enter to submit · Esc to cancel"))
	b.WriteString("\n")

	if m.err != "" {
		b.WriteString("\n")
		b.WriteString(formErrorStyle.Render(fmt.Sprintf("  Error: %s", m.err)))
		b.WriteString("\n")
	}

	return b.String()
}