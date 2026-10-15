package tui

import (
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/iax7/mybible/internal/module"
)

// pickerKind selects which modal is open and what it lists.
type pickerKind int

const (
	pickerNone pickerKind = iota
	pickerTranslation
	pickerBook
	pickerChapter
	pickerReference
	pickerSearch
)

// choice is the state shared by the modals. The list modals (Translation, Book,
// Chapter) filter a list and confirm a highlighted row; the Reference modal
// edits free text and keeps an error message when the input is rejected.
type choice struct {
	kind   pickerKind
	query  string
	cursor int
	err    string
}

func (c choice) isOpen() bool { return c.kind != pickerNone }

// pickerItem is one row of a list modal: the text to show, the text to filter
// by, and the value the row selects.
type pickerItem struct {
	label  string
	search string
	value  int
}

// openPicker starts a modal, preselecting the reader's current position where
// that makes sense.
func (m Model) openPicker(kind pickerKind) Model {
	m.picker = choice{kind: kind}
	m.results = nil
	switch kind {
	case pickerBook:
		m.picker.cursor = m.bookIndex(m.book.Number)
	case pickerChapter:
		m.picker.cursor = m.chapter - 1
	}
	m.picker.clamp(len(m.pickerMatches()))
	return m
}

// updatePicker handles keystrokes while a modal is open: typing filters the
// list, Enter confirms, Esc closes. The Reference modal takes the same keys.
func (m Model) updatePicker(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyClose:
		m.picker = choice{}
	case keyEnter:
		m = m.pickerConfirm()
	case keyBackspace:
		if runes := []rune(m.picker.query); len(runes) > 0 {
			m.picker.query = string(runes[:len(runes)-1])
		}
		m.picker.cursor = 0
		m.picker.err = ""
		m = m.refreshSearch()
	case keyUp:
		m.picker.cursor--
	case keyDown:
		m.picker.cursor++
	default:
		if msg.Text != "" {
			m.picker.query += msg.Text
			m.picker.cursor = 0
			m.picker.err = ""
			m = m.refreshSearch()
		}
	}
	if m.picker.isOpen() {
		m.picker.clamp(len(m.pickerMatches()))
	}
	return m, nil
}

// refreshSearch recomputes the Search modal's hits after its query changes. It
// does nothing for the other modals.
func (m Model) refreshSearch() Model {
	if m.picker.kind != pickerSearch {
		return m
	}
	if m.text == nil {
		m.results = nil
		return m
	}
	m.results = m.text.Search(m.picker.query)
	return m
}

// pickerConfirm applies the modal's current selection and closes it, or reports
// why the Reference modal's input cannot be used.
func (m Model) pickerConfirm() Model {
	kind := m.picker.kind
	if kind == pickerReference {
		return m.confirmReference()
	}

	items := m.pickerMatches()
	if m.picker.cursor < 0 || m.picker.cursor >= len(items) {
		return m
	}
	value := items[m.picker.cursor].value
	m.picker = choice{}

	switch kind {
	case pickerTranslation:
		if value >= 0 && value < len(m.translations) {
			m.active = m.translations[value].Abbreviation
			m = m.loadActive()
			m = m.persist()
		}
	case pickerBook:
		m = m.gotoBook(value)
	case pickerChapter:
		m = m.gotoChapter(value)
	case pickerSearch:
		m = m.confirmSearch(value)
	}
	return m
}

// confirmSearch jumps to the Search result at index i, closing the modal.
func (m Model) confirmSearch(i int) Model {
	if i < 0 || i >= len(m.results) {
		return m
	}
	return m.jumpToReference(m.results[i].Reference)
}

// jumpToReference jumps and closes the modal only when the Active translation
// actually contains the Reference, the way the other jumps behave.
func (m Model) jumpToReference(reference module.Reference) Model {
	next, ok := m.jumpTo(reference)
	if !ok {
		return m
	}
	next.picker = choice{}
	return next
}

// confirmReference parses the Reference modal's input and jumps to it, keeping
// the modal open with a message when the input cannot be used.
func (m Model) confirmReference() Model {
	reference, ok := module.ParseReference(m.picker.query)
	if !ok {
		m.picker.err = "That is not a Reference. Try John 3:16."
		return m
	}
	next, ok := m.jumpTo(reference)
	if !ok {
		m.picker.err = "No such Reference in the Active translation."
		return m
	}
	next.picker = choice{}
	return next
}

// pickerItems is the full, unfiltered list a modal shows.
func (m Model) pickerItems() []pickerItem {
	switch m.picker.kind {
	case pickerTranslation:
		items := make([]pickerItem, 0, len(m.translations))
		for i, t := range m.translations {
			items = append(items, pickerItem{
				label:  pad(t.Abbreviation, 8) + t.Title,
				search: t.Abbreviation + " " + t.Title,
				value:  i,
			})
		}
		return items
	case pickerBook:
		books := m.books()
		items := make([]pickerItem, 0, len(books))
		for _, b := range books {
			items = append(items, pickerItem{label: b.Name, search: b.Name, value: b.Number})
		}
		return items
	case pickerChapter:
		items := make([]pickerItem, 0, m.book.Chapters)
		for chapter := 1; chapter <= m.book.Chapters; chapter++ {
			items = append(items, pickerItem{label: strconv.Itoa(chapter), search: strconv.Itoa(chapter), value: chapter})
		}
		return items
	case pickerSearch:
		items := make([]pickerItem, 0, len(m.results))
		for i, result := range m.results {
			items = append(items, pickerItem{
				label: result.Reference.String() + "  " + result.Snippet,
				value: i,
			})
		}
		return items
	}
	return nil
}

// pickerMatches is the items that satisfy the modal's filter query. The Search
// modal is already matched against the Active translation, so its items pass
// through unfiltered.
func (m Model) pickerMatches() []pickerItem {
	items := m.pickerItems()
	if m.picker.kind == pickerSearch {
		return items
	}
	query := strings.ToLower(m.picker.query)
	if query == "" {
		return items
	}
	var matches []pickerItem
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.search), query) {
			matches = append(matches, item)
		}
	}
	return matches
}

// pickerRows is how many list rows fit above the modal's frame.
func (m Model) pickerRows() int {
	rows := m.height - 10
	if rows < 1 {
		rows = 1
	}
	return rows
}

// pickerTitle is the heading of the open list modal.
func (m Model) pickerTitle() string {
	switch m.picker.kind {
	case pickerBook:
		return selectBook
	case pickerChapter:
		return selectChapter
	case pickerSearch:
		return searchTitle
	default:
		return selectTranslation
	}
}

// pickerFilterLabel is the heading of the modal's input line.
func (m Model) pickerFilterLabel() string {
	if m.picker.kind == pickerSearch {
		return searchTitle + ": "
	}
	return "Filter: "
}

// pickerEmptyMessage explains an empty list, distinguishing "nothing typed yet"
// from "nothing matched".
func (m Model) pickerEmptyMessage() string {
	if m.picker.kind == pickerSearch && m.picker.query == "" {
		return "Type to search"
	}
	return "No matches"
}

// pickerView is a list modal: a filter line above the matching rows, windowed
// so a long list never overflows the screen, centred on the screen.
func (m Model) pickerView() string {
	if m.picker.kind == pickerReference {
		return m.referencePickerView()
	}

	var b strings.Builder
	b.WriteString(labelStyle.Render(m.pickerTitle()))
	b.WriteString("\n\n")
	b.WriteString(labelStyle.Render(m.pickerFilterLabel()) + m.picker.query)
	b.WriteString("\n\n")

	matches := m.pickerMatches()
	if len(matches) == 0 {
		b.WriteString(labelStyle.Render(m.pickerEmptyMessage()))
	}
	rows, start := m.pickerRows(), windowStart(m.picker.cursor, len(matches), m.pickerRows())
	for i := start; i < len(matches) && i < start+rows; i++ {
		line := truncate(matches[i].label, m.width-10)
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

// referencePickerView is the Go to Reference modal: a text field with a hint or
// the reason the last input was rejected.
func (m Model) referencePickerView() string {
	var b strings.Builder
	b.WriteString(labelStyle.Render(selectReference))
	b.WriteString("\n\n")
	b.WriteString(labelStyle.Render("Reference: ") + m.picker.query)
	b.WriteString("\n\n")
	if m.picker.err != "" {
		b.WriteString(warningStyle.Render(m.picker.err))
	} else {
		b.WriteString(labelStyle.Render("For example: John 3:16"))
	}
	b.WriteString("\n" + labelStyle.Render(keyEnter+" to go · "+keyClose+" to close"))

	box := helpStyle.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// windowStart is the first row of a window of rows lines that keeps the cursor
// visible inside a list of total rows.
func windowStart(cursor, total, rows int) int {
	if cursor < 0 {
		cursor = 0
	}
	start := cursor - rows/2
	if start < 0 {
		start = 0
	}
	if last := total - rows; start > last {
		start = last
	}
	if start < 0 {
		start = 0
	}
	return start
}

// clamp keeps the modal's cursor inside the list it currently matches.
func (c *choice) clamp(length int) {
	if c.cursor >= length {
		c.cursor = length - 1
	}
	if c.cursor < 0 {
		c.cursor = 0
	}
}
