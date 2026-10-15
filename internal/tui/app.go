// Package tui renders the reader's terminal interface.
package tui

import (
	"fmt"
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

// Keys that open a modal or a command.
const (
	keyQuit        = "q"
	keyForceQuit   = "ctrl+c"
	keyToggleBar   = "ctrl+b"
	keyHelp        = "?"
	keyBook        = "b"
	keyChapter     = "c"
	keyReference   = "v"
	keyTranslation = "t"
	keySearch      = "/"
	keyHistory     = "H"
	keyEnter       = "enter"
	keyClose       = "esc"
	keyUp          = "up"
	keyDown        = "down"
	keyBackspace   = "backspace"
)

// Keys that move the Verse cursor and the Chapter view.
const (
	keyVerseDown   = "j"
	keyVerseUp     = "k"
	keyPageDown    = "ctrl+d"
	keyPageUp      = "ctrl+u"
	keyPrevChapter = "["
	keyNextChapter = "]"
	keyTop         = "g"
	keyBottom      = "G"
)

// Keys that walk the History. A jump records an entry; these retrace them.
const (
	keyBack    = "alt+left"
	keyForward = "alt+right"
)

// Modal titles.
const (
	selectTranslation = "Select Translation"
	selectBook        = "Select Book"
	selectChapter     = "Select Chapter"
	selectReference   = "Go to Reference"
	searchTitle       = "Search"
	historyTitle      = "History"
)

const emptyReference = "—"

var (
	barStyle       = lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("252"))
	keyStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	labelStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	refStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))
	warningStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	helpStyle      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 3)
	italicsStyle   = lipgloss.NewStyle().Italic(true)
	jesusStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	verseNumStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorNumStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
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

	configDir    string
	moduleDir    string
	translations []module.Translation
	active       string
	warnings     []string

	text    *module.Text
	book    module.Book
	chapter int
	cursor  int // index of the current Verse within the Chapter
	top     int // first visible rendered line of the Chapter

	picker  choice
	results []module.Result // the Search modal's current hits

	history  []module.Reference // visited References in visit order
	position int                // index into history of the current Reference
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
	m = m.loadActive()

	if cfg.LastReference != "" {
		if reference, ok := module.ParseReference(cfg.LastReference); ok {
			m, _ = m.jumpTo(reference)
		}
	}
	return m
}

// loadActive loads the Active translation's whole text into memory and points
// the reader at its first Chapter. A load failure is a warning, not a crash.
func (m Model) loadActive() Model {
	m.text, m.book, m.chapter, m.cursor, m.top = nil, module.Book{}, 0, 0, 0
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
			return m.persist(), tea.Quit
		}
		if m.help {
			if key == keyClose {
				m.help = false
			}
			return m, nil
		}
		if m.picker.isOpen() {
			return m.updatePicker(msg)
		}
		switch key {
		case keyQuit:
			return m.persist(), tea.Quit
		case keyToggleBar:
			m.showBar = !m.showBar
		case keyHelp:
			m.help = true
		case keyBook:
			m = m.openPicker(pickerBook)
		case keyChapter:
			m = m.openPicker(pickerChapter)
		case keyReference:
			m = m.openPicker(pickerReference)
		case keyTranslation:
			m = m.openPicker(pickerTranslation)
		case keySearch:
			m = m.openPicker(pickerSearch)
		case keyHistory:
			m = m.openPicker(pickerHistory)
		case keyBack:
			m = m.back()
		case keyForward:
			m = m.forward()
		case keyVerseDown:
			m = m.moveCursor(1)
		case keyVerseUp:
			m = m.moveCursor(-1)
		case keyPageDown:
			m = m.moveCursor(m.pageSize())
		case keyPageUp:
			m = m.moveCursor(-m.pageSize())
		case keyPrevChapter:
			m = m.moveChapter(-1)
		case keyNextChapter:
			m = m.moveChapter(1)
		case keyTop:
			m = m.setCursor(0)
		case keyBottom:
			m = m.setCursor(len(m.chapterVerses()) - 1)
		}
		m = m.ensureVisible()
	}
	return m, nil
}

// persist writes the Active translation, the Module directory, and the
// last-read Reference to config, reporting a write failure as a warning rather
// than losing the change silently.
func (m Model) persist() Model {
	if m.configDir == "" {
		m.warnings = append(m.warnings, "cannot save the reader's state: no config directory")
		return m
	}
	cfg := config.Config{ModuleDir: m.moduleDir, ActiveTranslation: m.active}
	if reference, ok := m.currentReference(); ok {
		cfg.LastReference = reference.String()
	}
	if err := config.Save(m.configDir, cfg); err != nil {
		m.warnings = append(m.warnings, err.Error())
	}
	return m
}

// currentReference is the Reference of the cursor's Verse, when a Chapter is
// loaded.
func (m Model) currentReference() (module.Reference, bool) {
	verses := m.chapterVerses()
	if m.book.Number == 0 || m.cursor < 0 || m.cursor >= len(verses) {
		return module.Reference{}, false
	}
	return module.Reference{Book: m.book, Chapter: m.chapter, Verse: verses[m.cursor].Number}, true
}

// View implements tea.Model.
func (m Model) View() tea.View {
	content := m.readerView()
	if m.picker.isOpen() {
		content = m.pickerView()
	}
	if m.help {
		content = m.helpView()
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// paneHeight is the height left for the reading pane below the top bar.
func (m Model) paneHeight() int {
	if m.showBar {
		if height := m.height - lipgloss.Height(m.barView()); height > 0 {
			return height
		}
		return 0
	}
	return m.height
}

// readerView is the reading pane: the top bar above the pane content.
func (m Model) readerView() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	bar := ""
	if m.showBar {
		bar = m.barView()
	}
	paneHeight := m.height - lipgloss.Height(bar)
	pane := lipgloss.Place(m.width, paneHeight, lipgloss.Left, lipgloss.Top, m.paneContent(paneHeight))
	if bar == "" {
		return pane
	}
	return lipgloss.JoinVertical(lipgloss.Left, bar, pane)
}

// paneContent is the reading pane: the current Chapter when a Translation is
// loaded, otherwise an explanation of the state of the Module directory.
func (m Model) paneContent(height int) string {
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
		b.WriteString(m.chapterView(height))
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

// activeTranslation is the Translation currently displayed.
func (m Model) activeTranslation() (module.Translation, bool) {
	for _, t := range m.translations {
		if t.Abbreviation == m.active {
			return t, true
		}
	}
	return module.Translation{}, false
}

// barView lists every command with its key, and the current Reference on the right.
func (m Model) barView() string {
	parts := make([]string, 0, len(Commands))
	for _, c := range Commands {
		parts = append(parts, keyStyle.Render(c.Key)+" "+labelStyle.Render(c.Label))
	}
	left := strings.Join(parts, "  ")

	right := emptyReference
	if reference, ok := m.currentReference(); ok {
		right = reference.String()
	}
	right = refStyle.Render(right)

	// When the commands and the Reference do not fit side by side, the Reference
	// moves to its own line so neither is silently cut off.
	if lipgloss.Width(left)+lipgloss.Width(right)+1 > m.width {
		gap := m.width - lipgloss.Width(right)
		if gap < 0 {
			gap = 0
		}
		return barStyle.Width(m.width).Render(left) + "\n" +
			barStyle.Width(m.width).Render(strings.Repeat(" ", gap)+right)
	}

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
	b.WriteString(binding(keyBack, "Back"))
	b.WriteString(binding(keyForward, "Forward"))
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

// truncate shortens s to width cells, marking a cut with an ellipsis.
func truncate(s string, width int) string {
	if width < 1 || lipgloss.Width(s) <= width {
		return s
	}
	var b strings.Builder
	used := 0
	for _, r := range s {
		w := lipgloss.Width(string(r))
		if used+w > width-1 {
			break
		}
		b.WriteRune(r)
		used += w
	}
	return b.String() + "…"
}
