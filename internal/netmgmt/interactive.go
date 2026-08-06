package netmgmt

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// AddType selects what kind of interface to add.
type AddType int

const (
	AddVLAN AddType = iota
	AddBond
	AddInterface
)

// AddResult holds the form output.
type AddResult struct {
	Type     AddType
	Name     string
	Parent   string
	VLANID   string
	Mode     string
	Slaves   string
	IPAddr   string
	Netmask  string
	Gateway  string
}

// addModel is the bubbletea model for the add form.
type addModel struct {
	step      int
	addType   AddType
	inputs    []textinput.Model
	done      bool
	result    *AddResult
	err       error
	width     int
	height    int
}

var (
	addTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("99")).
			Padding(0, 2)

	addLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("36")).
			Bold(true)

	addHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	addErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("203")).
			Bold(true)
)

// RunAddForm launches an interactive bubbletea form for adding network interfaces.
func RunAddForm() (*AddResult, error) {
	// First, select type
	addType, err := selectAddType()
	if err != nil {
		return nil, err
	}
	if addType < 0 {
		return nil, nil // user cancelled
	}

	// Build form based on type
	m := newAddModel(AddType(addType))
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	if m, ok := finalModel.(addModel); ok && m.result != nil {
		return m.result, nil
	}
	return nil, nil
}

// selectAddType shows a simple selection menu for the add type.
func selectAddType() (int, error) {
	items := []string{"VLAN", "Bond", "Interface (dummy)"}
	sel := newTypeSelector(items)
	p := tea.NewProgram(sel)
	final, err := p.Run()
	if err != nil {
		return -1, err
	}
	if s, ok := final.(typeSelector); ok {
		return s.selected, nil
	}
	return -1, nil
}

// --- Type Selector ---

type typeSelector struct {
	items    []string
	cursor   int
	selected int
}

func newTypeSelector(items []string) typeSelector {
	return typeSelector{
		items:    items,
		selected: -1,
	}
}

func (m typeSelector) Init() tea.Cmd { return nil }

func (m typeSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			m.selected = m.cursor
			return m, tea.Quit
		case "esc", "q":
			m.selected = -1
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m typeSelector) View() string {
	var b strings.Builder
	b.WriteString(addTitleStyle.Render(" Add Network Interface "))
	b.WriteString("\n\n")
	b.WriteString(addHintStyle.Render("Select interface type (↑/↓, Enter to confirm, Esc to cancel):"))
	b.WriteString("\n\n")
	for i, item := range m.items {
		cursor := " "
		if i == m.cursor {
			cursor = "▸"
		}
		b.WriteString(fmt.Sprintf("  %s %s\n", cursor, item))
	}
	return b.String()
}

// --- Add Form ---

func newAddModel(t AddType) addModel {
	m := addModel{
		addType: t,
		result:  &AddResult{Type: t},
	}

	var inputs []textinput.Model

	switch t {
	case AddVLAN:
		inputs = append(inputs, newInput("Interface name (e.g. eth0.100)", ""))
		inputs = append(inputs, newInput("Parent interface (e.g. eth0)", ""))
		inputs = append(inputs, newInput("VLAN ID (1-4094)", ""))
		inputs = append(inputs, newInput("IP address (optional, e.g. 192.168.1.10/24)", ""))
		inputs = append(inputs, newInput("Gateway (optional)", ""))
	case AddBond:
		inputs = append(inputs, newInput("Bond name (e.g. bond0)", ""))
		inputs = append(inputs, newInput("Bond mode (balance-rr/active-backup/802.3ad/balance-tlb/balance-alb)", "active-backup"))
		inputs = append(inputs, newInput("Slave interfaces (comma-separated, e.g. eth0,eth1)", ""))
		inputs = append(inputs, newInput("IP address (optional, e.g. 192.168.1.10/24)", ""))
		inputs = append(inputs, newInput("Gateway (optional)", ""))
	case AddInterface:
		inputs = append(inputs, newInput("Interface name (e.g. dummy0)", ""))
		inputs = append(inputs, newInput("IP address (optional, e.g. 192.168.1.10/24)", ""))
		inputs = append(inputs, newInput("Gateway (optional)", ""))
	}

	if len(inputs) > 0 {
		inputs[0].Focus()
	}

	m.inputs = inputs
	return m
}

func newInput(placeholder, defaultValue string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 200
	ti.Width = 50
	if defaultValue != "" {
		ti.SetValue(defaultValue)
	}
	return ti
}

func (m addModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m addModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		case "tab", "shift+tab", "enter":
			if msg.String() == "enter" && m.step == len(m.inputs)-1 {
				m.collectResult()
				m.done = true
				return m, tea.Quit
			}
			if msg.String() == "shift+tab" {
				m.prevInput()
			} else {
				m.nextInput()
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.inputs[m.step], cmd = m.inputs[m.step].Update(msg)
	return m, cmd
}

func (m *addModel) nextInput() {
	m.inputs[m.step].Blur()
	if m.step < len(m.inputs)-1 {
		m.step++
		m.inputs[m.step].Focus()
	}
}

func (m *addModel) prevInput() {
	m.inputs[m.step].Blur()
	if m.step > 0 {
		m.step--
		m.inputs[m.step].Focus()
	}
}

func (m *addModel) collectResult() {
	r := m.result
	switch m.addType {
	case AddVLAN:
		r.Name = m.inputs[0].Value()
		r.Parent = m.inputs[1].Value()
		r.VLANID = m.inputs[2].Value()
		r.IPAddr = m.inputs[3].Value()
		r.Gateway = m.inputs[4].Value()
	case AddBond:
		r.Name = m.inputs[0].Value()
		r.Mode = m.inputs[1].Value()
		r.Slaves = m.inputs[2].Value()
		r.IPAddr = m.inputs[3].Value()
		r.Gateway = m.inputs[4].Value()
	case AddInterface:
		r.Name = m.inputs[0].Value()
		r.IPAddr = m.inputs[1].Value()
		r.Gateway = m.inputs[2].Value()
	}
}

func (m addModel) View() string {
	if m.done {
		return ""
	}

	var b strings.Builder
	titles := map[AddType]string{
		AddVLAN:     "Add VLAN Interface",
		AddBond:     "Add Bond Interface",
		AddInterface: "Add Interface",
	}
	b.WriteString(addTitleStyle.Render(" "+titles[m.addType]+" "))
	b.WriteString("\n\n")

	labels := m.fieldLabels()
	for i, input := range m.inputs {
		label := addLabelStyle.Render(labels[i])
		b.WriteString(fmt.Sprintf("  %s\n", label))
		b.WriteString(fmt.Sprintf("  %s\n", input.View()))
		if i < len(m.inputs)-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(addHintStyle.Render("  Tab to next field · Enter to submit · Esc to cancel"))
	b.WriteString("\n")

	return b.String()
}

func (m addModel) fieldLabels() []string {
	switch m.addType {
	case AddVLAN:
		return []string{"Name", "Parent Interface", "VLAN ID", "IP Address (CIDR)", "Gateway"}
	case AddBond:
		return []string{"Name", "Bond Mode", "Slave Interfaces", "IP Address (CIDR)", "Gateway"}
	case AddInterface:
		return []string{"Name", "IP Address (CIDR)", "Gateway"}
	default:
		return nil
	}
}

// --- Remove Form ---

type removeModel struct {
	ifaces   []Interface
	cursor   int
	selected int
	done     bool
	result   string
}

// RunRemoveForm launches an interactive selection for removing an interface.
func RunRemoveForm(ifaces []Interface) (string, error) {
	m := newRemoveModel(ifaces)
	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return "", err
	}
	if rm, ok := final.(removeModel); ok && rm.done {
		return rm.result, nil
	}
	return "", nil
}

func newRemoveModel(ifaces []Interface) removeModel {
	return removeModel{
		ifaces:   ifaces,
		selected: -1,
	}
}

func (m removeModel) Init() tea.Cmd { return nil }

func (m removeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.ifaces)-1 {
				m.cursor++
			}
		case "enter":
			if m.cursor >= 0 && m.cursor < len(m.ifaces) {
				m.selected = m.cursor
				m.result = m.ifaces[m.cursor].Name
				m.done = true
				return m, tea.Quit
			}
		case "esc", "q":
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m removeModel) View() string {
	var b strings.Builder
	b.WriteString(addTitleStyle.Render(" Remove Network Interface "))
	b.WriteString("\n\n")
	b.WriteString(addHintStyle.Render("Select interface to remove (↑/↓, Enter to confirm, Esc to cancel):"))
	b.WriteString("\n\n")

	for i, iface := range m.ifaces {
		cursor := " "
		if i == m.cursor {
			cursor = "▸"
		}
		state := "DOWN"
		if iface.Up {
			state = "UP"
		}
		b.WriteString(fmt.Sprintf("  %s %-12s %-8s %-4s\n", cursor, iface.Name, iface.Type, state))
	}

	return b.String()
}