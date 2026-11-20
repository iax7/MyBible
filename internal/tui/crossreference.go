package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/iax7/mybible/internal/module"
)

// crossReferenceChapter is the active sets' Cross-references for the Chapter
// being read, read once when the Chapter or the active sets change.
type crossReferenceChapter struct {
	book, chapter int
	active        string // the active abbreviations, sorted and joined
	byVerse       map[int][]module.CrossReference
}

// activeCrossReferenceSets is the installed sets the reader turned on, in
// discovery order.
func (m Model) activeCrossReferenceSets() []module.CrossReferenceSet {
	sets := make([]module.CrossReferenceSet, 0, len(m.crossReferenceSets))
	for _, set := range m.crossReferenceSets {
		if m.crossReferenceActive(set.Abbreviation) {
			sets = append(sets, set)
		}
	}
	return sets
}

// crossReferenceActive reports whether the reader turned one set on.
func (m Model) crossReferenceActive(abbreviation string) bool {
	for _, name := range m.activeCrossReferences {
		if name == abbreviation {
			return true
		}
	}
	return false
}

// crossReferenceKey identifies the active sets, so the cache knows when it is
// stale after a set is turned on or off.
func (m Model) crossReferenceKey() string {
	active := append([]string(nil), m.activeCrossReferences...)
	sort.Strings(active)
	return strings.Join(active, ",")
}

// loadCrossReferences reads the active sets' rows for the current Chapter and
// groups them by the Verse each link speaks to. It is the only Cross-reference
// read the reading pane does; the rows stay out of memory until a Chapter needs
// them.
func (m Model) loadCrossReferences() Model {
	m.crossReferences = crossReferenceChapter{
		book:    m.book.Number,
		chapter: m.chapter,
		active:  m.crossReferenceKey(),
	}
	sets := m.activeCrossReferenceSets()
	if len(sets) == 0 || m.text == nil || m.book.Number == 0 || m.chapter < 1 {
		return m
	}

	var rows []module.CrossReference
	for _, set := range sets {
		references, err := set.Rows(m.book.Number, m.chapter)
		if err != nil {
			m.warnings = append(m.warnings, fmt.Sprintf("%s: %v", set.Abbreviation, err))
			continue
		}
		rows = append(rows, references...)
	}
	m.crossReferences.byVerse = make(map[int][]module.CrossReference)
	for _, verse := range m.chapterVerses() {
		if references := module.CrossReferencesFor(rows, verse.Number); len(references) > 0 {
			m.crossReferences.byVerse[verse.Number] = references
		}
	}
	return m
}

// crossReferencesFresh reports whether the cache belongs to the Chapter and sets
// being read.
func (m Model) crossReferencesFresh() bool {
	return m.crossReferences.book == m.book.Number &&
		m.crossReferences.chapter == m.chapter &&
		m.crossReferences.active == m.crossReferenceKey()
}

// crossReferencesFor is the Cross-references of one Verse of the current Chapter.
func (m Model) crossReferencesFor(verse int) []module.CrossReference {
	if !m.crossReferencesFresh() {
		return nil
	}
	return m.crossReferences.byVerse[verse]
}

// crossReferencesModal is the open References popup: the Cross-references of the
// cursor's Verse, each with the text it points to.
type crossReferencesModal struct {
	open      bool
	reference module.Reference
	lines     []string
	scroll    int
}

// crossReferencesPicker is the open selector of Cross-reference sets, with the
// row the reader is on.
type crossReferencesPicker struct {
	open   bool
	cursor int
}

// openCrossReferences opens the References popup for the cursor's Verse.
func (m Model) openCrossReferences() Model {
	reference, ok := m.currentReference()
	if !ok {
		return m
	}
	if !m.crossReferencesFresh() {
		m = m.loadCrossReferences()
	}
	return m.openCrossReferencesAt(reference)
}

// openCrossReferencesAt builds the popup for a Reference. It explains the two
// empty states — no sets turned on, and no links for the Verse — because the
// reader asked for the popup on purpose.
func (m Model) openCrossReferencesAt(reference module.Reference) Model {
	width := m.commentaryWidth()
	references := m.crossReferencesFor(reference.Verse)
	var lines []string
	switch {
	case len(m.activeCrossReferenceSets()) == 0:
		lines = []string{
			m.translate("No cross-reference modules active."),
			"",
			m.translate("Press X to choose some."),
		}
	case len(references) == 0:
		lines = []string{m.translate("No references for this verse.")}
	default:
		for i, crossReference := range references {
			if i > 0 {
				lines = append(lines, "")
			}
			lines = append(lines, m.crossReferenceLines(crossReference, width)...)
		}
	}
	m.crossReferencesModal = crossReferencesModal{open: true, reference: reference, lines: lines}
	return m
}

// crossReferenceLines renders one Cross-reference: its label in the link style,
// then the text it points to, wrapped to width. A translation without the
// target Book still shows the label, as the Module's own reference text does.
func (m Model) crossReferenceLines(crossReference module.CrossReference, width int) []string {
	label := crossReference.Label(func(number int) string {
		return m.displayBook(module.Book{Number: number}).DisplayName()
	})
	lines := []string{m.styles.link.Render(label)}
	for _, line := range wrapRuns(m.crossReferencePassage(crossReference), width) {
		lines = append(lines, m.renderRuns(line))
	}
	return lines
}

// crossReferencePassage is the text a Cross-reference points to, read from the
// Active translation. Verses are joined into one paragraph, so a range reads as
// the passage rather than as a list.
func (m Model) crossReferencePassage(crossReference module.CrossReference) []module.Run {
	if m.text == nil {
		return nil
	}
	end := crossReference.TargetVerseEnd
	if end < crossReference.TargetVerse {
		end = crossReference.TargetVerse
	}
	var runs []module.Run
	for _, verse := range m.text.Chapter(crossReference.TargetBook, crossReference.TargetChapter) {
		if verse.Number < crossReference.TargetVerse || verse.Number > end {
			continue
		}
		if len(runs) > 0 {
			runs = append(runs, module.Run{Text: " "})
		}
		runs = append(runs, module.ParseMarkup(verse.Text)...)
	}
	return runs
}

// updateCrossReferences handles keystrokes while the References popup is open:
// ↑/↓ or j/k and the page keys scroll, Esc closes.
func (m Model) updateCrossReferences(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	modal := &m.crossReferencesModal
	switch msg.String() {
	case keyClose:
		*modal = crossReferencesModal{}
		return m, nil
	case keyDown, keyVerseDown:
		modal.scroll++
	case keyUp, keyVerseUp:
		modal.scroll--
	case keyPageDown:
		modal.scroll += m.crossReferencesRows()
	case keyPageUp:
		modal.scroll -= m.crossReferencesRows()
	}
	modal.scroll = clampCommentaryScroll(modal.scroll, len(modal.lines), m.crossReferencesRows())
	return m, nil
}

// crossReferencesView is the open References popup: the visible part of its
// lines and the key hint.
func (m Model) crossReferencesView() string {
	modal := m.crossReferencesModal
	rows := m.crossReferencesRows()
	scroll := clampCommentaryScroll(modal.scroll, len(modal.lines), rows)
	visible := modal.lines
	if scroll < len(visible) {
		visible = visible[scroll:]
	}
	if len(visible) > rows {
		visible = visible[:rows]
	}

	body := strings.Join(visible, "\n")
	body += "\n" + m.styles.hint.Render(m.translate("↑/↓, j/k to scroll"))
	title := m.translate(crossReferencesTitle) + " · " + m.displayReference(modal.reference)
	return m.styles.frameBox(title, keyClose, body)
}

// crossReferencesRows is how many text lines fit between the popup's edges and
// its key hint.
func (m Model) crossReferencesRows() int {
	rows := m.height - 9
	if rows < 1 {
		rows = 1
	}
	return rows
}

// openCrossReferencesPicker opens the selector of Cross-reference sets.
func (m Model) openCrossReferencesPicker() Model {
	m.crossReferencesPicker = crossReferencesPicker{open: true}
	return m
}

// updateCrossReferencesPicker handles keystrokes while the selector is open:
// ↑/↓ or j/k move, Space or Enter turns a set on or off, Esc closes.
func (m Model) updateCrossReferencesPicker(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	picker := &m.crossReferencesPicker
	switch msg.String() {
	case keyClose:
		*picker = crossReferencesPicker{}
		return m, nil
	case keyUp, keyVerseUp:
		picker.cursor--
	case keyDown, keyVerseDown:
		picker.cursor++
	case keyEnter:
		m = m.toggleCrossReference()
	default:
		if msg.Text == " " {
			m = m.toggleCrossReference()
		}
	}
	m.crossReferencesPicker.clamp(len(m.crossReferenceSets))
	return m, nil
}

// toggleCrossReference turns the highlighted set on or off and rereads the
// Chapter, so the markers follow the reader's choice at once. The empty list is
// kept non-nil: it is what tells a restart the reader wants no sets, rather than
// the first one by default.
func (m Model) toggleCrossReference() Model {
	cursor := m.crossReferencesPicker.cursor
	if cursor < 0 || cursor >= len(m.crossReferenceSets) {
		return m
	}
	abbreviation := m.crossReferenceSets[cursor].Abbreviation
	active := make([]string, 0, len(m.activeCrossReferences)+1)
	for _, name := range m.activeCrossReferences {
		if name != abbreviation {
			active = append(active, name)
		}
	}
	if !m.crossReferenceActive(abbreviation) {
		active = append(active, abbreviation)
	}
	m.activeCrossReferences = active
	m = m.persist()
	return m.loadCrossReferences()
}

// crossReferencesPickerView is the selector: one row per installed set, marked
// when it is on, with the keys that work it.
func (m Model) crossReferencesPickerView() string {
	var lines []string
	if len(m.crossReferenceSets) == 0 {
		lines = append(lines,
			m.styles.label.Render(m.translate("No cross-reference modules found.")),
			"",
			m.styles.hint.Render(m.translate("Put your *.crossreferences.SQLite3 modules there.")))
	} else {
		rows := m.crossReferencesPickerRows()
		start := windowStart(m.crossReferencesPicker.cursor, len(m.crossReferenceSets), rows)
		for i := start; i < len(m.crossReferenceSets) && i < start+rows; i++ {
			set := m.crossReferenceSets[i]
			mark := "[ ]"
			if m.crossReferenceActive(set.Abbreviation) {
				mark = "[x]"
			}
			line := truncate(mark+" "+pad(set.Abbreviation, 12)+set.Title, m.commentaryWidth())
			if i == m.crossReferencesPicker.cursor {
				line = m.styles.key.Render("> ") + line
			} else {
				line = "  " + line
			}
			lines = append(lines, line)
		}
	}
	lines = append(lines, "", m.styles.hint.Render(m.translate("space to toggle · esc to close")))
	return m.styles.frameBox(m.translate(selectCrossReferenceModules), keyClose, strings.Join(lines, "\n"))
}

// crossReferencesPickerRows is how many set rows fit above the selector's keys.
func (m Model) crossReferencesPickerRows() int {
	rows := m.height - 8
	if rows < 1 {
		rows = 1
	}
	return rows
}

// clamp keeps the selector's cursor inside the sets it lists.
func (p *crossReferencesPicker) clamp(length int) {
	if p.cursor >= length {
		p.cursor = length - 1
	}
	if p.cursor < 0 {
		p.cursor = 0
	}
}
