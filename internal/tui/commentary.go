package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/iax7/mybible/internal/module"
)

// commentaryChapter is the Active commentary's rows for the Chapter being read,
// read once when the Chapter or the Active commentary changes.
type commentaryChapter struct {
	book, chapter int
	active        string
	rows          []module.CommentaryRow
}

// commentaryModal is the open Commentary modal: one tab per Commentary module
// that speaks to the cursor's Verse, and the line the reader has scrolled to.
type commentaryModal struct {
	open      bool
	reference module.Reference
	tabs      []commentaryTab
	cursor    int
	scroll    int
}

// commentaryTab is one module's Commentary on the Verse, already wrapped to the
// modal's width.
type commentaryTab struct {
	abbreviation string
	lines        []string
}

// activeCommentaryModule is the Commentary module the reader chose, when it is
// still installed.
func (m Model) activeCommentaryModule() (module.Commentary, bool) {
	for _, commentary := range m.commentaries {
		if commentary.Abbreviation == m.activeCommentary {
			return commentary, true
		}
	}
	return module.Commentary{}, false
}

// loadCommentary reads the Active commentary's rows for the current Chapter. It
// is the only commentary read the reading pane does; the text stays out of
// memory until a Chapter needs it.
func (m Model) loadCommentary() Model {
	m.commentary = commentaryChapter{book: m.book.Number, chapter: m.chapter, active: m.activeCommentary}
	commentary, ok := m.activeCommentaryModule()
	if !ok || m.text == nil || m.book.Number == 0 || m.chapter < 1 {
		return m
	}
	rows, err := commentary.Rows(m.book.Number, m.chapter)
	if err != nil {
		m.warnings = append(m.warnings, fmt.Sprintf("%s: %v", commentary.Abbreviation, err))
		return m
	}
	m.commentary.rows = rows
	return m
}

// commentaryFresh reports whether the cached rows belong to the Chapter and
// commentary being read.
func (m Model) commentaryFresh() bool {
	return m.commentary.book == m.book.Number &&
		m.commentary.chapter == m.chapter &&
		m.commentary.active == m.activeCommentary
}

// commentaryLabels is the set of Verses of the current Chapter whose text ends
// with the Active commentary's label. A Chapter-level entry labels Verse 1 only.
func (m Model) commentaryLabels() map[int]bool {
	if m.activeCommentary == "" || !m.commentaryFresh() {
		return nil
	}
	labels := make(map[int]bool)
	verses := m.chapterVerses()
	for _, row := range m.commentary.rows {
		if row.ChapterLevel() {
			if row.ChapterFrom == m.chapter {
				labels[1] = true
			}
			continue
		}
		for _, verse := range verses {
			if row.Covers(m.chapter, verse.Number) {
				labels[verse.Number] = true
			}
		}
	}
	return labels
}

// activeCommentaryCovers reports whether the Active commentary has something to
// say about one Verse of the current Chapter.
func (m Model) activeCommentaryCovers(verse int) bool {
	if m.activeCommentary == "" || !m.commentaryFresh() {
		return false
	}
	for _, row := range m.commentary.rows {
		if row.ChapterLevel() {
			if verse == 1 && row.ChapterFrom == m.chapter {
				return true
			}
			continue
		}
		if row.Covers(m.chapter, verse) {
			return true
		}
	}
	return false
}

// appendCommentaryLabel writes the Active commentary's label at the end of a
// Verse's last line, or on an indented line of its own when it does not fit.
func (m Model) appendCommentaryLabel(lines []string, indent int) []string {
	if len(lines) == 0 {
		return lines
	}
	label := m.styles.hint.Render(" [" + m.activeCommentary + "]")
	last := len(lines) - 1
	if lipgloss.Width(lines[last])+lipgloss.Width(label) <= m.width {
		lines[last] += label
		return lines
	}
	return append(lines, strings.Repeat(" ", indent)+label)
}

// openCommentary opens the Commentary modal for the cursor's Verse, when the
// Active commentary speaks to it.
func (m Model) openCommentary() Model {
	reference, ok := m.currentReference()
	if !ok || !m.activeCommentaryCovers(reference.Verse) {
		return m
	}
	return m.openCommentaryAt(reference)
}

// openCommentaryAt builds the modal for a Reference, reading every Commentary
// module on demand and keeping only the rows that speak to that Verse.
func (m Model) openCommentaryAt(reference module.Reference) Model {
	tabs, warnings := m.commentaryTabs(reference)
	m.warnings = append(m.warnings, warnings...)
	if len(tabs) == 0 {
		m.commentaryModal = commentaryModal{}
		return m
	}
	cursor := 0
	for i, tab := range tabs {
		if tab.abbreviation == m.activeCommentary {
			cursor = i
			break
		}
	}
	m.commentaryModal = commentaryModal{open: true, reference: reference, tabs: tabs, cursor: cursor}
	return m
}

// commentaryTabs reads every installed Commentary module for the Reference's
// Chapter and keeps the modules that speak to its Verse, active one included.
func (m Model) commentaryTabs(reference module.Reference) ([]commentaryTab, []string) {
	width := m.commentaryWidth()
	tabs := make([]commentaryTab, 0, len(m.commentaries))
	var warnings []string
	for _, commentary := range m.commentaries {
		rows, err := commentary.Rows(reference.Book.Number, reference.Chapter)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", commentary.Abbreviation, err))
			continue
		}
		var blocks []module.CommentaryRow
		seen := make(map[string]bool)
		for _, row := range rows {
			if strings.TrimSpace(row.Text) == "" || !commentaryRowCovers(row, reference.Chapter, reference.Verse) || seen[row.Text] {
				continue
			}
			seen[row.Text] = true
			blocks = append(blocks, row)
		}
		if len(blocks) == 0 {
			continue
		}
		tabs = append(tabs, commentaryTab{
			abbreviation: commentary.Abbreviation,
			lines:        m.commentaryLines(blocks, width),
		})
	}
	return tabs, warnings
}

// commentaryRowCovers reports whether a row belongs to the Verse being read. A
// Chapter-level row belongs to Verse 1, the Verse the Chapter opens on.
func commentaryRowCovers(row module.CommentaryRow, chapter, verse int) bool {
	if row.ChapterLevel() {
		return chapter == row.ChapterFrom && verse == 1
	}
	return row.Covers(chapter, verse)
}

// commentaryLines renders a module's rows for one Verse: a quiet range heading
// where the row spans more than its own Verse, then the text wrapped to width.
func (m Model) commentaryLines(rows []module.CommentaryRow, width int) []string {
	var lines []string
	for i, row := range rows {
		if i > 0 {
			lines = append(lines, "")
		}
		if heading := m.commentaryRange(row); heading != "" {
			lines = append(lines, m.styles.hint.Render(heading))
		}
		for _, line := range splitVerseLines(module.ParseCommentary(row.Text)) {
			for _, wrapped := range wrapRuns(line, width) {
				lines = append(lines, m.renderCommentaryRuns(wrapped))
			}
		}
	}
	return lines
}

// renderCommentaryRuns renders one line of Commentary Runs: bold text in the
// theme's white and verse links in the theme's green, underlined.
func (m Model) renderCommentaryRuns(runs []module.Run) string {
	var b strings.Builder
	for _, run := range runs {
		b.WriteString(m.commentaryRunStyle(run).Render(run.Text))
	}
	return b.String()
}

// commentaryRunStyle styles one Commentary Run. A link keeps the boldness of
// the text around it and adds the link style's green underline.
func (m Model) commentaryRunStyle(run module.Run) lipgloss.Style {
	style := lipgloss.NewStyle()
	if run.Bold {
		style = m.styles.commentaryBold
	}
	if run.Link {
		style = m.styles.link
		if run.Bold {
			style = style.Bold(true)
		}
	}
	if run.Italics {
		style = style.Italic(true)
	}
	return style
}

// commentaryRange is the heading of one row: its Verse range, or the note that
// the row is a Chapter overview. A row about the Verse alone has no heading.
func (m Model) commentaryRange(row module.CommentaryRow) string {
	if row.ChapterLevel() {
		return m.translate("Chapter overview")
	}
	fromChapter, fromVerse := row.ChapterFrom, row.VerseFrom
	toChapter, toVerse := row.ChapterTo, row.VerseTo
	if toChapter < 1 {
		toChapter, toVerse = fromChapter, fromVerse
	}
	switch {
	case toChapter == fromChapter && toVerse == fromVerse:
		return ""
	case toChapter == fromChapter && toVerse < 1:
		return fmt.Sprintf("%d–%s", fromVerse, m.translate("end"))
	case toChapter == fromChapter:
		return fmt.Sprintf("%d–%d", fromVerse, toVerse)
	case toVerse < 1:
		return fmt.Sprintf("%d:%d–%d:%s", fromChapter, fromVerse, toChapter, m.translate("end"))
	default:
		return fmt.Sprintf("%d:%d–%d:%d", fromChapter, fromVerse, toChapter, toVerse)
	}
}

// updateCommentary handles keystrokes while the Commentary modal is open: Tab
// and h/l switch modules, ↑/↓ or j/k and the page keys scroll, Esc closes.
func (m Model) updateCommentary(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	modal := &m.commentaryModal
	if !modal.open || len(modal.tabs) == 0 {
		return m, nil
	}
	switch msg.String() {
	case keyClose:
		*modal = commentaryModal{}
		return m, nil
	case keyTab, keyPrevTab, "h", "l":
		delta := 1
		if msg.String() == keyPrevTab || msg.String() == "h" {
			delta = -1
		}
		modal.cursor = (modal.cursor + delta + len(modal.tabs)) % len(modal.tabs)
		modal.scroll = 0
	case keyDown, keyVerseDown:
		modal.scroll++
	case keyUp, keyVerseUp:
		modal.scroll--
	case keyPageDown:
		modal.scroll += m.commentaryRows()
	case keyPageUp:
		modal.scroll -= m.commentaryRows()
	}
	modal.scroll = clampCommentaryScroll(modal.scroll, len(modal.tabs[modal.cursor].lines), m.commentaryRows())
	return m, nil
}

// commentaryView is the open Commentary modal: the tab strip, the visible part
// of the selected module's text, and the key hint.
func (m Model) commentaryView() string {
	modal := m.commentaryModal
	if len(modal.tabs) == 0 {
		return ""
	}
	tab := modal.tabs[modal.cursor]
	rows := m.commentaryRows()
	scroll := clampCommentaryScroll(modal.scroll, len(tab.lines), rows)
	visible := tab.lines
	if scroll < len(visible) {
		visible = visible[scroll:]
	}
	if len(visible) > rows {
		visible = visible[:rows]
	}

	body := m.commentaryTabStrip()
	if body != "" {
		body += "\n\n"
	}
	body += strings.Join(visible, "\n")
	body += "\n" + m.styles.hint.Render(m.translate("tab to switch · ↑/↓, j/k to scroll"))
	return m.styles.frameBox(m.translate(commentaryTitle)+" · "+m.displayReference(modal.reference), keyClose, body)
}

// commentaryTabStrip is the row of module names that heads the modal, the
// selected one in the key color and the rest in the quiet hint color.
func (m Model) commentaryTabStrip() string {
	parts := make([]string, 0, len(m.commentaryModal.tabs))
	for i, tab := range m.commentaryModal.tabs {
		style := m.styles.hint
		if i == m.commentaryModal.cursor {
			style = m.styles.key
		}
		parts = append(parts, style.Render(" "+tab.abbreviation+" "))
	}
	return strings.Join(wrapParts(parts, m.commentaryWidth()), "\n")
}

// commentaryWidth is the width the modal wraps its text to: the screen less the
// frame, capped so a long line stays readable on a wide terminal.
func (m Model) commentaryWidth() int {
	width := m.width - 8
	if width < 20 {
		width = 20
	}
	if width > 100 {
		width = 100
	}
	return width
}

// commentaryRows is how many text lines fit between the tab strip and the hint.
func (m Model) commentaryRows() int {
	rows := m.height - 10 - lipgloss.Height(m.commentaryTabStrip())
	if rows < 1 {
		rows = 1
	}
	return rows
}

// clampCommentaryScroll keeps a scroll offset inside the text it scrolls.
func clampCommentaryScroll(scroll, total, rows int) int {
	if rows < 1 {
		rows = 1
	}
	if scroll > total-rows {
		scroll = total - rows
	}
	if scroll < 0 {
		scroll = 0
	}
	return scroll
}
