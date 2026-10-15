// Package tui renders the reader's terminal interface.
package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Command is one entry in the top bar: the key that invokes an action and its label.
type Command struct {
	Key   string
	Label string
}

// Commands are the actions listed in the top bar, in display order.
var Commands = []Command{
	{"b", "Book"},
	{"c", "Chapter"},
	{"v", "Go To"},
	{"t", "Translation"},
	{"/", "Search"},
	{"H", "History"},
	{"?", "Help"},
	{"q", "Quit"},
}

// Keys that drive the interface but are not themselves commands.
const (
	keyQuit      = "q"
	keyForceQuit = "ctrl+c"
	keyToggleBar = "ctrl+b"
	keyHelp      = "?"
	keyClose     = "esc"
)

const emptyReference = "—"

var (
	barStyle   = lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("252"))
	keyStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	refStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))
	helpStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 3)
)

// Model is the reader's Bubble Tea model.
type Model struct {
	width, height int
	showBar       bool
	help          bool
	reference     string
}

// New returns a Model in its initial state: an empty reading pane with the top bar shown.
func New() Model {
	return Model{showBar: true}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case tea.KeyPressMsg:
		key := msg.String()
		if key == keyForceQuit {
			return m, tea.Quit
		}
		if m.help {
			if key == keyClose {
				m.help = false
			}
			return m, nil
		}
		switch key {
		case keyQuit:
			return m, tea.Quit
		case keyToggleBar:
			m.showBar = !m.showBar
		case keyHelp:
			m.help = true
		}
	}
	return m, nil
}

// View implements tea.Model.
func (m Model) View() tea.View {
	content := m.readerView()
	if m.help {
		content = m.helpView()
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// readerView is the reading pane: the top bar above an otherwise empty pane.
func (m Model) readerView() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	if !m.showBar {
		return blank(m.width, m.height)
	}
	bar := m.barView()
	return lipgloss.JoinVertical(lipgloss.Left, bar, blank(m.width, m.height-lipgloss.Height(bar)))
}

// barView lists every command with its key, and the current Reference on the right.
func (m Model) barView() string {
	parts := make([]string, 0, len(Commands))
	for _, c := range Commands {
		parts = append(parts, keyStyle.Render(c.Key)+" "+labelStyle.Render(c.Label))
	}
	left := strings.Join(parts, "  ")

	right := m.reference
	if right == "" {
		right = emptyReference
	}
	right = refStyle.Render(right)

	// shortcut: a long Reference on a narrow terminal truncates the command list,
	// rework the bar layout (drop labels or wrap) when References land in ticket 03.
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	line := left + strings.Repeat(" ", gap) + right
	return barStyle.Width(m.width).MaxWidth(m.width).Render(line)
}

// helpView is the modal listing every keybinding, centred on the screen.
func (m Model) helpView() string {
	var b strings.Builder
	b.WriteString(labelStyle.Render("Keys"))
	b.WriteString("\n\n")
	for _, c := range Commands {
		b.WriteString(binding(c.Key, c.Label))
	}
	b.WriteString(binding(keyToggleBar, "Toggle top bar"))
	b.WriteString(binding(keyClose, "Close"))

	box := helpStyle.Render(strings.TrimRight(b.String(), "\n"))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func binding(key, label string) string {
	return keyStyle.Render(pad(key, 10)) + labelStyle.Render(label) + "\n"
}

func pad(s string, width int) string {
	if n := width - lipgloss.Width(s); n > 0 {
		return s + strings.Repeat(" ", n)
	}
	return s
}

func blank(width, height int) string {
	if height <= 0 {
		return ""
	}
	return lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, "")
}
