// Package tui renders the reader's terminal interface.
package tui

import (
	"fmt"
	"strings"
	"time"

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

// Commands are every action the reader has, in display order. The Help modal
// lists them all; the top bar announces only barCommands.
var Commands = []Command{
	{"b", "Book"},
	{"c", "Chapter"},
	{"v", "Go To"},
	{"t", "Translation"},
	{"m", "Commentary"},
	{keyEnter, "Open commentary"},
	{"/", "Search"},
	{"H", "History"},
	{keyLog, "Log"},
	{keyToggleBar, "Toggle top bar"},
	{keyBack, "Back"},
	{keyForward, "Forward"},
	{"?", "Help"},
	{"q", "Quit"},
}

// barCommands are the Commands the top bar announces, in display order. The
// Active translation and the Reference being read lead the bar, each with the
// key that changes it.
var barCommands = []Command{
	{keyHistory, "History"},
	{keySearch, "Search"},
	{keyHelp, "Help"},
}

// MovementCommands are the keys that move the Verse cursor or the Chapter view,
// listed in the Help modal apart from the commands.
var MovementCommands = []Command{
	{keyVerseDown, "Verse down"},
	{keyVerseUp, "Verse up"},
	{keyPageDown, "Page down"},
	{keyPageUp, "Page up"},
	{keyPrevChapter, "Previous Chapter"},
	{keyNextChapter, "Next Chapter"},
	{keyTop, "Top of Chapter"},
	{keyBottom, "Bottom of Chapter"},
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
	keyCommentary  = "m"
	keySearch      = "/"
	keyHistory     = "H"
	keyLog         = "L"
	keyEnter       = "enter"
	keyTab         = "tab"
	keyPrevTab     = "shift+tab"
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
	selectCommentary  = "Select Commentary"
	selectBook        = "Select Book"
	selectChapter     = "Select Chapter"
	selectReference   = "Go to Reference"
	searchTitle       = "Search"
	historyTitle      = "History"
	logTitle          = "Log"
	commentaryTitle   = "Commentary"
)

const emptyReference = "—"

// Options configures where the reader finds its state. Empty fields fall back
// to the XDG config directory and to the Module directory named in config, and
// then to the reader's default data directory.
type Options struct {
	ConfigDir string
	ModuleDir string
}

// Reader is the module layer the TUI reads through: the Active translation's
// Books and Chapters as values, with lookup by Book number and by search term.
// It is defined here, at its point of use; the module layer implements it.
type Reader interface {
	Books() []module.Book
	Book(number int) (module.Book, bool)
	BookIndex(number int) int
	Chapter(number, chapter int) []module.Verse
	Subheadings(number, chapter int) []module.Subheading
	Search(query string) []module.Result
}

var _ Reader = (*module.Text)(nil) // the module layer's implementation

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

	commentaries     []module.Commentary
	activeCommentary string
	commentary       commentaryChapter // the Active commentary's rows for the current Chapter
	commentaryModal  commentaryModal

	cfg    config.Config // the loaded config, kept so saving state does not drop keys
	lang   lang
	styles styles

	text    Reader
	book    module.Book
	chapter int
	cursor  int // index of the current Verse within the Chapter
	top     int // first visible rendered line of the Chapter

	picker  choice
	results []module.Result // the Search modal's current hits

	trail    []module.Reference // the session's Visits in visit order
	position int                // index into trail of the current Reference

	history       []visit // Visits kept across restarts, in visit order
	historyFailed bool    // a History read or write failed: warn once, not on every jump
	now           func() time.Time
}

// New returns a Model after loading config and scanning the Module directory.
// Problems are reported as warnings in the view rather than as errors.
func New(opts Options) Model {
	m := Model{showBar: true, configDir: opts.ConfigDir, now: time.Now}
	if m.configDir == "" {
		if dir, err := config.Dir(); err == nil {
			m.configDir = dir
		}
	}

	cfg, err := config.Load(m.configDir)
	if err != nil {
		m.warnings = append(m.warnings, err.Error())
	}
	m.cfg = cfg
	m.lang = resolveLanguage(cfg.Language)
	m.styles = newStyles(paletteFor(cfg.Theme, cfg.Colors), bookColorsFor(cfg.Theme))
	m = m.loadHistory()

	m.moduleDir = opts.ModuleDir
	if m.moduleDir == "" {
		m.moduleDir = cfg.ModuleDir
	}
	if m.moduleDir == "" {
		// Without a folder of its own the reader falls back to the default, so
		// Modules work before anything is configured. The pane explains the
		// state when even that cannot be resolved.
		if dir, err := config.DefaultModuleDir(); err == nil {
			m.moduleDir = dir
		}
	}
	if m.moduleDir != "" {
		translations, warnings := module.Discover(m.moduleDir)
		m.translations = translations
		m.warnings = append(m.warnings, warnings...)

		commentaries, warnings := module.DiscoverCommentaries(m.moduleDir)
		m.commentaries = commentaries
		m.warnings = append(m.warnings, warnings...)
	}

	m.activeCommentary = cfg.ActiveCommentary
	if _, ok := m.activeCommentaryModule(); !ok {
		m.activeCommentary = ""
		if len(m.commentaries) > 0 {
			m.activeCommentary = m.commentaries[0].Abbreviation
		}
	}

	m.active = cfg.ActiveTranslation
	if _, ok := m.activeTranslation(); !ok {
		m.active = ""
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
	m.commentary = commentaryChapter{}
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
	return m.loadCommentary()
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.commentaryModal.open {
			m = m.openCommentaryAt(m.commentaryModal.reference)
		}

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
		if m.commentaryModal.open {
			return m.updateCommentary(msg)
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
		case keyCommentary:
			m = m.openPicker(pickerCommentary)
		case keyEnter:
			m = m.openCommentary()
		case keySearch:
			m = m.openPicker(pickerSearch)
		case keyHistory:
			m = m.openPicker(pickerHistory)
		case keyLog:
			m = m.openPicker(pickerLog)
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
// last-read Reference to config, keeping the theme and language it was loaded
// with, and reporting a write failure as a warning rather than losing the
// change silently.
func (m Model) persist() Model {
	if m.configDir == "" {
		m.warnings = append(m.warnings, "cannot save the reader's state: no config directory")
		return m
	}
	cfg := m.cfg
	cfg.ModuleDir = m.moduleDir
	cfg.ActiveTranslation = m.active
	cfg.ActiveCommentary = m.activeCommentary
	cfg.LastReference = ""
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
	if box := m.modalView(); box != "" {
		x, y := centred(box, m.width, m.height)
		content = overlay(content, box, m.width, m.height, x, y)
		if popup := m.pickerPopupView(); popup != "" {
			content = overlay(content, popup, m.width, m.height, x+pickerPopupInset, y+lipgloss.Height(box))
		}
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// modalView is the open modal's box, or an empty string when nothing is open.
func (m Model) modalView() string {
	switch {
	case m.help:
		return m.helpView()
	case m.commentaryModal.open:
		return m.commentaryView()
	case m.picker.isOpen():
		return m.pickerView()
	}
	return ""
}

// centred is where a box sits centred on a screen of the given size.
func centred(box string, width, height int) (int, int) {
	x := (width - lipgloss.Width(box)) / 2
	y := (height - lipgloss.Height(box)) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	return x, y
}

// overlay draws box at (x, y) over base without erasing what base shows around
// it: the box covers the cells it needs and the reading pane stays visible
// behind. The position is kept inside the screen so a box never pushes the
// screen taller or wider than it is.
func overlay(base, box string, width, height, x, y int) string {
	if x > width-lipgloss.Width(box) {
		x = width - lipgloss.Width(box)
	}
	if y > height-lipgloss.Height(box) {
		y = height - lipgloss.Height(box)
	}
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	composite := lipgloss.NewCompositor(
		lipgloss.NewLayer(base),
		lipgloss.NewLayer(box).X(x).Y(y),
	).Render()
	return padLines(composite, width)
}

// padLines pads every line of text to width cells, so a screen always keeps its
// shape whatever the overlays left behind.
func padLines(text string, width int) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if w := lipgloss.Width(line); w < width {
			lines[i] = line + strings.Repeat(" ", width-w)
		}
	}
	return strings.Join(lines, "\n")
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
// loaded, otherwise an explanation of the state of the Module directory. Warnings
// are not shown here; they live in the Log modal.
func (m Model) paneContent(height int) string {
	var b strings.Builder
	switch {
	case m.moduleDir == "":
		b.WriteString(m.translate("No Module directory configured.") + "\n\n")
		b.WriteString(fmt.Sprintf(m.translate("Set module_directory in %s"), config.Path(m.configDir)) + "\n")
		b.WriteString(m.translate("to the folder holding your MyBible *.SQLite3 modules."))
	case len(m.translations) == 0:
		b.WriteString(fmt.Sprintf(m.translate("No Bible Modules found in %s."), m.moduleDir) + "\n\n")
		b.WriteString(m.translate("Put your *.SQLite3 modules there, then restart the reader."))
	case m.active == "":
		b.WriteString(m.translate("No Active translation.") + "\n\n")
		fmt.Fprintf(&b, m.translate("Press %s to choose one."), keyTranslation)
	case m.book.Number != 0:
		b.WriteString(m.chapterView(height))
	default:
		t, _ := m.activeTranslation()
		b.WriteString(t.Title + "\n\n")
		b.WriteString(m.styles.label.Render(t.Language))
	}
	return b.String()
}

// renderRuns renders a line of Verse Runs with each Run's style.
func (m Model) renderRuns(runs []module.Run) string {
	var b strings.Builder
	for _, run := range runs {
		b.WriteString(m.runStyle(run).Render(run.Text))
	}
	return b.String()
}

// runStyle is the display style for one Run of Verse text.
func (m Model) runStyle(run module.Run) lipgloss.Style {
	switch {
	case run.Red && run.Italics:
		return m.styles.jesus.Italic(true)
	case run.Red:
		return m.styles.jesus
	case run.Italics:
		return m.styles.italics
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

// barView lists the Active translation and the Reference being read with the
// commands that act on them. Entries wrap onto extra lines whole, so a label is
// never split across lines.
func (m Model) barView() string {
	parts := make([]string, 0, len(barCommands)+3)
	parts = append(parts, m.barEntry(keyTranslation, m.activeLabel()))
	// The Reference fills its Book's color like a selected tab, so it skips the
	// label style the other entries carry.
	parts = append(parts, m.styles.key.Render(keyReference)+" "+m.referenceLabel())
	for _, c := range barCommands {
		parts = append(parts, m.barEntry(c.Key, m.translate(c.Label)))
	}
	if len(m.warnings) > 0 {
		parts = append(parts, m.logEntry())
	}
	lines := wrapParts(parts, m.width)

	for i, line := range lines {
		lines[i] = m.styles.bar.Width(m.width).MaxWidth(m.width).Render(line)
	}
	return strings.Join(lines, "\n")
}

// logEntry is the top bar's Log entry: the key that opens the Log and a warning
// sign, in the theme's warning color.
func (m Model) logEntry() string {
	return m.styles.key.Render(keyLog) + " " + m.styles.label.Render(m.translate(logTitle)) + " " + m.styles.warning.Render("⚠")
}

// barEntry is one top bar entry: the key that changes the state it shows.
func (m Model) barEntry(key, label string) string {
	return m.styles.key.Render(key) + " " + m.styles.label.Render(label)
}

// activeLabel is the abbreviation of the Active translation, or a placeholder
// when the reader holds none.
func (m Model) activeLabel() string {
	if m.active == "" {
		return emptyReference
	}
	return m.active
}

// referenceLabel is the Book and Chapter being read, filled with the Book's
// classification color like a selected tab, or a placeholder before the reader
// opens a Book.
func (m Model) referenceLabel() string {
	if m.book.Number == 0 {
		return m.styles.reference.Render(emptyReference)
	}
	return m.styles.bookTab(m.book.Number).Render(fmt.Sprintf(" %s %d ", m.book.Name, m.chapter))
}

// wrapParts lays out command entries on as few lines as fit width, never
// splitting one entry across lines.
func wrapParts(parts []string, width int) []string {
	var lines []string
	line := ""
	for _, part := range parts {
		switch {
		case line == "":
			line = part
		case lipgloss.Width(line)+2+lipgloss.Width(part) <= width:
			line += "  " + part
		default:
			lines = append(lines, line)
			line = part
		}
	}
	return append(lines, line)
}

// helpView is the modal listing every keybinding, centred on the screen.
func (m Model) helpView() string {
	var b strings.Builder
	b.WriteString(m.styles.label.Render(m.translate("Movement")) + "\n")
	for _, c := range MovementCommands {
		b.WriteString(binding(m.styles, c.Key, m.translate(c.Label)))
	}
	b.WriteString("\n" + m.styles.label.Render(m.translate("Commands")) + "\n")
	for _, c := range Commands {
		b.WriteString(binding(m.styles, c.Key, m.translate(c.Label)))
	}
	b.WriteString(binding(m.styles, keyClose, m.translate("Close")))

	return m.styles.frameBox(m.translate("Keys"), keyClose, strings.TrimRight(b.String(), "\n"))
}

func binding(s styles, key, label string) string {
	return s.key.Render(pad(key, 10)) + s.label.Render(label) + "\n"
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
