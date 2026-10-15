// Package tui renders the reader's terminal interface.
package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/iax7/mybible/internal/config"
	"github.com/iax7/mybible/internal/module"
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
	keyQuit        = "q"
	keyForceQuit   = "ctrl+c"
	keyToggleBar   = "ctrl+b"
	keyHelp        = "?"
	keyClose       = "esc"
	keyTranslation = "t"
	keyEnter       = "enter"
	keyUp          = "up"
	keyDown        = "down"
	keyBackspace   = "backspace"
)

// selectTranslation titles the Select Translation modal.
const selectTranslation = "Select Translation"

const emptyReference = "—"

var (
	barStyle      = lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("252"))
	keyStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	labelStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	refStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))
	warningStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	helpStyle     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 3)
	italicsStyle  = lipgloss.NewStyle().Italic(true)
	jesusStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	verseNumStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// Options configures where the reader finds its state. Empty fields fall back
// to the XDG config directory and to the Module directory named in config.
type Options struct {
	ConfigDir string
	ModuleDir string
}

// Model is the reader's Bubble Tea model.
type Model struct {
	width, height int
	showBar       bool
	help          bool
	reference     string

	configDir    string
	moduleDir    string
	translations []module.Translation
	active       string
	warnings     []string

	text    *module.Text
	book    module.Book
	chapter int

	picker choice
}

// choice is the state shared by the filterable modals that pick one item.
type choice struct {
	open   bool
	query  string
	cursor int
}

// New returns a Model after loading config and scanning the Module directory.
// Problems are reported as warnings in the view rather than as errors.
func New(opts Options) Model {
	m := Model{showBar: true, configDir: opts.ConfigDir}
	if m.configDir == "" {
		if dir, err := config.Dir(); err == nil {
			m.configDir = dir
		}
	}

	cfg, err := config.Load(m.configDir)
	if err != nil {
		m.warnings = append(m.warnings, err.Error())
	}

	m.moduleDir = opts.ModuleDir
	if m.moduleDir == "" {
		m.moduleDir = cfg.ModuleDir
	}
	if m.moduleDir != "" {
		translations, warnings := module.Discover(m.moduleDir)
		m.translations = translations
		m.warnings = append(m.warnings, warnings...)
	}

	if cfg.ActiveTranslation != "" {
		for _, t := range m.translations {
			if t.Abbreviation == cfg.ActiveTranslation {
				m.active = cfg.ActiveTranslation
				break
			}
		}
	}
	return m.loadActive()
}

// loadActive loads the Active translation's whole text into memory and points
// the reader at its first Chapter. A load failure is a warning, not a crash.
func (m Model) loadActive() Model {
	m.text, m.book, m.chapter, m.reference = nil, module.Book{}, 0, ""
	t, ok := m.activeTranslation()
	if !ok {
		return m
	}
	text, err := t.Load()
	if err != nil {
		m.warnings = append(m.warnings, fmt.Sprintf("%s: %v", t.Abbreviation, err))
		return m
	}
	m.text = text
	if books := text.Books(); len(books) > 0 {
		m.book, m.chapter = books[0], 1
		m.reference = fmt.Sprintf("%s %d", m.book.Name, m.chapter)
	}
	return m
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
		if m.picker.open {
			return m.updatePicker(msg)
		}
		switch key {
		case keyQuit:
			return m, tea.Quit
		case keyToggleBar:
			m.showBar = !m.showBar
		case keyHelp:
			m.help = true
		case keyTranslation:
			m.picker = choice{open: true}
		}
	}
	return m, nil
}

// updatePicker handles keystrokes while the Select Translation modal is open:
// typing filters the list, Enter picks the highlighted Translation, Esc closes.
func (m Model) updatePicker(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyClose:
		m.picker = choice{}
	case keyEnter:
		if t, ok := m.selectedTranslation(); ok {
			m.active = t.Abbreviation
			m.picker = choice{}
			m = m.loadActive()
			m = m.persist()
		}
	case keyBackspace:
		if runes := []rune(m.picker.query); len(runes) > 0 {
			m.picker.query = string(runes[:len(runes)-1])
		}
		m.picker.cursor = 0
	case keyUp:
		m.picker.cursor--
	case keyDown:
		m.picker.cursor++
	default:
		if msg.Text != "" {
			m.picker.query += msg.Text
			m.picker.cursor = 0
		}
	}
	m.picker.clamp(len(m.matches()))
	return m, nil
}

// persist writes the Active translation and the Module directory to config,
// reporting a write failure as a warning rather than losing the change silently.
func (m Model) persist() Model {
	if m.configDir == "" {
		m.warnings = append(m.warnings, "cannot save the Active translation: no config directory")
		return m
	}
	cfg := config.Config{ModuleDir: m.moduleDir, ActiveTranslation: m.active}
	if err := config.Save(m.configDir, cfg); err != nil {
		m.warnings = append(m.warnings, err.Error())
	}
	return m
}

// View implements tea.Model.
func (m Model) View() tea.View {
	content := m.readerView()
	if m.picker.open {
		content = m.pickerView()
	}
	if m.help {
		content = m.helpView()
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// readerView is the reading pane: the top bar above the pane content.
func (m Model) readerView() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	paneHeight := m.height
	bar := ""
	if m.showBar {
		bar = m.barView()
		paneHeight -= lipgloss.Height(bar)
	}
	pane := lipgloss.Place(m.width, paneHeight, lipgloss.Left, lipgloss.Top, m.paneContent())
	if bar == "" {
		return pane
	}
	return lipgloss.JoinVertical(lipgloss.Left, bar, pane)
}

// paneContent is the reading pane: the current Chapter when a Translation is
// loaded, otherwise an explanation of the state of the Module directory.
func (m Model) paneContent() string {
	var b strings.Builder
	switch {
	case m.moduleDir == "":
		b.WriteString("No Module directory configured.\n\n")
		b.WriteString("Set module_directory in " + config.Path(m.configDir) + "\n")
		b.WriteString("to the folder holding your MyBible *.SQLite3 modules.")
	case len(m.translations) == 0:
		b.WriteString("No Bible Modules found in " + m.moduleDir + ".\n\n")
		b.WriteString("Put your *.SQLite3 modules there, then restart the reader.")
	case m.active == "":
		b.WriteString("No Active translation.\n\n")
		b.WriteString("Press " + keyTranslation + " to choose one.")
	case m.book.Number != 0:
		b.WriteString(m.chapterView())
	default:
		t, _ := m.activeTranslation()
		b.WriteString(t.Title + "\n\n")
		b.WriteString(labelStyle.Render(t.Language))
	}
	for _, warning := range m.warnings {
		b.WriteString("\n\n" + warningStyle.Render("warning: "+warning))
	}
	return b.String()
}

// chapterView renders the current Chapter: one Verse per line, each with its
// number, keeping paragraph and line breaks.
func (m Model) chapterView() string {
	verses := m.text.Chapter(m.book.Number, m.chapter)
	var b strings.Builder
	for i, verse := range verses {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(verseNumStyle.Render(strconv.Itoa(verse.Number)))
		b.WriteString(" ")
		b.WriteString(renderMarkup(verse.Text))
	}
	return b.String()
}

// renderMarkup renders a Verse's inline markup in the supported subset.
func renderMarkup(text string) string {
	var b strings.Builder
	for _, run := range module.ParseMarkup(text) {
		b.WriteString(runStyle(run).Render(run.Text))
	}
	return b.String()
}

// runStyle is the display style for one Run of Verse text.
func runStyle(run module.Run) lipgloss.Style {
	switch {
	case run.Red && run.Italics:
		return jesusStyle.Italic(true)
	case run.Red:
		return jesusStyle
	case run.Italics:
		return italicsStyle
	}
	return lipgloss.NewStyle()
}

// pickerView is the Select Translation modal: a filter line above the matching
// Translations, centred on the screen.
func (m Model) pickerView() string {
	var b strings.Builder
	b.WriteString(labelStyle.Render(selectTranslation))
	b.WriteString("\n\n")
	b.WriteString(labelStyle.Render("Filter: ") + m.picker.query)
	b.WriteString("\n\n")

	matches := m.matches()
	if len(matches) == 0 {
		b.WriteString(labelStyle.Render("No matching Translations"))
	}
	for i, t := range matches {
		line := pad(t.Abbreviation, 8) + t.Title
		if i == m.picker.cursor {
			b.WriteString(keyStyle.Render("> "+line) + "\n")
		} else {
			b.WriteString("  " + line + "\n")
		}
	}
	b.WriteString("\n" + labelStyle.Render(keyEnter+" to select · "+keyClose+" to close"))

	box := helpStyle.Render(strings.TrimRight(b.String(), "\n"))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// matches is the Translations that satisfy the picker's filter query.
func (m Model) matches() []module.Translation {
	query := strings.ToLower(m.picker.query)
	if query == "" {
		return m.translations
	}
	var matches []module.Translation
	for _, t := range m.translations {
		if strings.Contains(strings.ToLower(t.Abbreviation), query) ||
			strings.Contains(strings.ToLower(t.Title), query) {
			matches = append(matches, t)
		}
	}
	return matches
}

// selectedTranslation is the Translation the picker's cursor is on.
func (m Model) selectedTranslation() (module.Translation, bool) {
	matches := m.matches()
	if m.picker.cursor < 0 || m.picker.cursor >= len(matches) {
		return module.Translation{}, false
	}
	return matches[m.picker.cursor], true
}

// activeTranslation is the Translation currently displayed.
func (m Model) activeTranslation() (module.Translation, bool) {
	for _, t := range m.translations {
		if t.Abbreviation == m.active {
			return t, true
		}
	}
	return module.Translation{}, false
}

// clamp keeps the picker's cursor inside the list it currently matches.
func (c *choice) clamp(length int) {
	if c.cursor >= length {
		c.cursor = length - 1
	}
	if c.cursor < 0 {
		c.cursor = 0
	}
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
