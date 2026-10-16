package tui

import (
	"fmt"
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
	pickerHistory
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
// by, and the value the row selects. head is the leading text that names a Book,
// colored with style, so the Book stands out from the rest of the row.
type pickerItem struct {
	label  string
	search string
	value  int
	head   string
	style  lipgloss.Style
}

// pickerSpec is everything that differs between modals: the heading, where the
// rows come from, what an empty list says, and what confirming a row does. The
// generic modal code reads this table, so adding a modal is one entry here.
type pickerSpec struct {
	title         string                         // interface string key of the heading
	items         func(m Model) []pickerItem     // the full, unfiltered row list
	cursor        func(m Model) int              // initial cursor row, or nil for the first
	onQuery       func(m Model) Model            // extra work when the typed query changes
	empty         func(m Model) string           // message when no row matches, or nil for the default
	confirm       func(m Model, value int) Model // apply the highlighted row
	filter        bool                           // the typed query narrows the rows
	filterByTitle bool                           // the input line is titled with the modal's own heading
	typedOnly     bool                           // the rows appear only once the query names something
}

// bookItems are the Active translation's Books as modal rows, in Canonical
// order. The Select Book and Go to Reference modals list the same rows; only
// what confirming one does differs.
func bookItems(m Model) []pickerItem {
	books := m.books()
	items := make([]pickerItem, 0, len(books))
	for _, b := range books {
		items = append(items, pickerItem{
			label:  b.Name,
			search: strings.Join(b.Names(), " "),
			value:  b.Number,
			head:   b.Name,
			style:  m.styles.book(b.Number),
		})
	}
	return items
}

// pickers holds one entry per pickerKind, indexed by the kind. pickerNone is the
// zero entry and is never shown.
var pickers = [...]pickerSpec{
	pickerNone: {},
	pickerTranslation: {
		title:  selectTranslation,
		filter: true,
		items: func(m Model) []pickerItem {
			items := make([]pickerItem, 0, len(m.translations))
			for i, t := range m.translations {
				items = append(items, pickerItem{
					label:  pad(t.Abbreviation, 8) + t.Title,
					search: t.Abbreviation + " " + t.Title,
					value:  i,
				})
			}
			return items
		},
		confirm: func(m Model, value int) Model {
			if value >= 0 && value < len(m.translations) {
				m.active = m.translations[value].Abbreviation
				m = m.loadActive()
				m = m.persist()
			}
			return m
		},
	},
	pickerBook: {
		title:   selectBook,
		filter:  true,
		items:   bookItems,
		cursor:  func(m Model) int { return m.bookIndex(m.book.Number) },
		confirm: func(m Model, value int) Model { return m.gotoBook(value) },
	},
	pickerChapter: {
		title:  selectChapter,
		filter: true,
		items: func(m Model) []pickerItem {
			items := make([]pickerItem, 0, m.book.Chapters)
			for chapter := 1; chapter <= m.book.Chapters; chapter++ {
				items = append(items, pickerItem{label: strconv.Itoa(chapter), search: strconv.Itoa(chapter), value: chapter})
			}
			return items
		},
		cursor:  func(m Model) int { return m.chapter - 1 },
		confirm: func(m Model, value int) Model { return m.gotoChapter(value) },
	},
	pickerReference: {
		title:     selectReference,
		filter:    true,
		typedOnly: true, // typing is what asks for the Books; listing them all is what b does
		items:     bookItems,
	},
	pickerSearch: {
		title:         searchTitle,
		filterByTitle: true,
		items: func(m Model) []pickerItem {
			items := make([]pickerItem, 0, len(m.results))
			for i, result := range m.results {
				items = append(items, pickerItem{
					label: result.Reference.String() + "  " + result.Snippet,
					value: i,
					head:  result.Reference.Book.Name,
					style: m.styles.book(result.Reference.Book.Number),
				})
			}
			return items
		},
		onQuery: func(m Model) Model {
			if m.text == nil {
				m.results = nil
				return m
			}
			m.results = m.text.Search(m.picker.query)
			return m
		},
		empty: func(m Model) string {
			if m.picker.query == "" {
				return m.translate("Type to search")
			}
			return m.translate("No matches")
		},
		confirm: func(m Model, value int) Model { return m.confirmSearch(value) },
	},
	pickerHistory: {
		title:  historyTitle,
		filter: true,
		items: func(m Model) []pickerItem {
			items := make([]pickerItem, 0, len(m.history))
			for i := len(m.history) - 1; i >= 0; i-- { // most recent first
				reference := m.history[i].String()
				items = append(items, pickerItem{
					label:  reference,
					search: reference,
					value:  i,
					head:   m.history[i].Book.Name,
					style:  m.styles.book(m.history[i].Book.Number),
				})
			}
			return items
		},
		empty: func(m Model) string {
			if len(m.history) == 0 {
				return m.translate("Nothing visited yet")
			}
			return m.translate("No matches")
		},
		confirm: func(m Model, value int) Model { return m.jumpToHistory(value) },
	},
}

// openPicker starts a modal, preselecting the reader's current position where
// that makes sense.
func (m Model) openPicker(kind pickerKind) Model {
	m.picker = choice{kind: kind}
	m.results = nil
	if cursor := pickers[kind].cursor; cursor != nil {
		m.picker.cursor = cursor(m)
	}
	m.picker.clamp(len(m.pickerMatches()))
	return m
}

// updatePicker handles keystrokes while a modal is open: typing filters the
// list, Enter confirms, Tab completes the Reference modal, Esc closes. Every
// modal takes the same keys.
func (m Model) updatePicker(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyClose:
		m.picker = choice{}
	case keyEnter:
		m = m.pickerConfirm()
	case keyTab:
		m = m.pickerComplete()
	case keyBackspace:
		before := m.pickerFilterQuery()
		if runes := []rune(m.picker.query); len(runes) > 0 {
			m.picker.query = string(runes[:len(runes)-1])
		}
		m.picker.err = ""
		m = m.refilter(before)
	case keyUp:
		m.picker.cursor--
	case keyDown:
		m.picker.cursor++
	default:
		if msg.Text != "" {
			before := m.pickerFilterQuery()
			m.picker.query += msg.Text
			m.picker.err = ""
			m = m.refilter(before)
		}
	}
	if m.picker.isOpen() {
		m.picker.clamp(len(m.pickerMatches()))
	}
	return m, nil
}

// refilter recomputes a modal's rows after its query changed, and returns the
// cursor to the first row when the text the rows are filtered by changed, so a
// narrowed list starts at its top. The Reference modal filters by the Book its
// input names, so a Chapter typed after the Book leaves the highlight alone.
func (m Model) refilter(before string) Model {
	if before != m.pickerFilterQuery() {
		m.picker.cursor = 0
	}
	return m.refreshSearch()
}

// pickerFilterQuery is the text a list modal filters its rows by: the whole
// query, except in the Reference modal, which filters by the Book its input
// names and ignores the Chapter:Verse written after it.
func (m Model) pickerFilterQuery() string {
	if m.picker.kind == pickerReference {
		book, _ := module.SplitReference(m.picker.query)
		return book
	}
	return m.picker.query
}

// refreshSearch recomputes a modal's rows after its query changes, when its
// spec asks for it. It does nothing for the other modals.
func (m Model) refreshSearch() Model {
	if onQuery := pickers[m.picker.kind].onQuery; onQuery != nil {
		return onQuery(m)
	}
	return m
}

// pickerConfirm applies the modal's current selection and closes it, or reports
// why the Reference modal's input cannot be used.
func (m Model) pickerConfirm() Model {
	spec := pickers[m.picker.kind]
	if m.picker.kind == pickerReference {
		return m.confirmReference()
	}

	items := m.pickerMatches()
	if m.picker.cursor < 0 || m.picker.cursor >= len(items) {
		return m
	}
	value := items[m.picker.cursor].value
	m.picker = choice{}
	if spec.confirm == nil {
		return m
	}
	return spec.confirm(m, value)
}

// pickerComplete writes the highlighted row into the Reference modal's input,
// spaced so a Chapter typed next lands beside the Book's name rather than on it,
// and keeping the Chapter:Verse already written. It is the only way the list and
// the input meet: Enter never guesses from the highlight. A modal that only
// lists rows, and a list with no matching row, is left alone.
func (m Model) pickerComplete() Model {
	if m.picker.kind != pickerReference {
		return m
	}
	items := m.pickerMatches()
	if m.picker.cursor < 0 || m.picker.cursor >= len(items) {
		return m
	}
	_, location := module.SplitReference(m.picker.query)
	m.picker.query = items[m.picker.cursor].head + " " // the Book's canonical Name
	if location != "" {
		m.picker.query += location
	}
	m.picker.err = ""
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
	next, ok := m.jumpAndRecord(reference)
	if !ok {
		return m
	}
	next.picker = choice{}
	return next
}

// confirmReference resolves the Reference modal's input and jumps to it, keeping
// the modal open with a message when the input cannot be used.
func (m Model) confirmReference() Model {
	reference, ok := m.parseReferenceInput()
	if !ok {
		m.picker.err = m.translate("That is not a Reference. Try John 3:16.")
		return m
	}
	next, ok := m.jumpAndRecord(reference)
	if !ok {
		m.picker.err = m.translate("No such Reference in the Active translation.")
		return m
	}
	next.picker = choice{}
	return next
}

// parseReferenceInput reads what the Reference modal's input says. An input that
// names no Book is read as the translation's first Book, so an empty input lands
// on its first Verse, and a Book named the way the Module itself names it is
// replaced by its Canonical name, which is the only form the parser knows.
// Everything else is left to module.ParseReference, so a Reference the reader
// knows but the Active translation lacks is still read as one and reported by
// the jump rather than as a malformed input.
func (m Model) parseReferenceInput() (module.Reference, bool) {
	book, location := module.SplitReference(m.picker.query)
	text := book
	if book == "" {
		items := m.pickerMatches() // the translation's first Book
		if m.picker.cursor < 0 || m.picker.cursor >= len(items) {
			return module.Reference{}, false
		}
		text = items[m.picker.cursor].head
	}
	if named, ok := m.namedBook(text); ok {
		text = named.Name
	}
	if location != "" {
		text += " " + location
	}
	return module.ParseReference(text)
}

// namedBook is the one Book of the Active translation whose name is text, the
// canonical name or the Module's own. A name no Book has, or one that more than
// one Book claims, names nothing.
func (m Model) namedBook(text string) (module.Book, bool) {
	var named module.Book
	found := 0
	for _, b := range m.books() {
		if b.Named(text) {
			named, found = b, found+1
		}
	}
	return named, found == 1
}

// pickerItems is the full, unfiltered list a modal shows.
func (m Model) pickerItems() []pickerItem {
	if items := pickers[m.picker.kind].items; items != nil {
		return items(m)
	}
	return nil
}

// pickerMatches is the items that satisfy the modal's filter query, ignoring
// letter case and accents. The Search modal is already matched against the
// Active translation, so its items pass through unfiltered.
func (m Model) pickerMatches() []pickerItem {
	items := m.pickerItems()
	if !pickers[m.picker.kind].filter {
		return items
	}
	query := module.Normalize(m.pickerFilterQuery())
	if query == "" {
		return items
	}
	var matches []pickerItem
	for _, item := range items {
		if strings.Contains(module.Normalize(item.search), query) {
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
	return m.translate(pickers[m.picker.kind].title)
}

// pickerFilterLabel is the heading of the modal's input line.
func (m Model) pickerFilterLabel() string {
	spec := pickers[m.picker.kind]
	if spec.filterByTitle {
		return m.translate(spec.title) + ": "
	}
	return m.translate("Filter: ")
}

// pickerEmptyMessage explains an empty list, distinguishing "nothing typed yet"
// from "nothing matched".
func (m Model) pickerEmptyMessage() string {
	if empty := pickers[m.picker.kind].empty; empty != nil {
		return empty(m)
	}
	return m.translate("No matches")
}

// pickerView is a list modal: a filter line above the matching rows, windowed
// so a long list never overflows the screen, centred on the screen.
func (m Model) pickerView() string {
	if m.picker.kind == pickerReference {
		return m.referencePickerView()
	}

	var b strings.Builder
	b.WriteString(m.styles.label.Render(m.pickerTitle()))
	b.WriteString("\n\n")
	b.WriteString(m.styles.label.Render(m.pickerFilterLabel()) + m.picker.query)
	b.WriteString("\n\n")
	b.WriteString(m.pickerRowLines())
	b.WriteString("\n" + m.styles.label.Render(m.translate(keyEnter+" to select · "+keyClose+" to close")))

	box := m.styles.help.Render(strings.TrimRight(b.String(), "\n"))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// pickerRowLines is a modal's matching rows, windowed around the cursor, or the
// message explaining why the list is empty. A modal that lists only what the
// reader asks for shows nothing before the first keystroke.
func (m Model) pickerRowLines() string {
	if pickers[m.picker.kind].typedOnly && m.pickerFilterQuery() == "" {
		return ""
	}

	matches := m.pickerMatches()
	if len(matches) == 0 {
		return m.styles.label.Render(m.pickerEmptyMessage()) + "\n"
	}

	var b strings.Builder
	rows, start := m.pickerRows(), windowStart(m.picker.cursor, len(matches), m.pickerRows())
	for i := start; i < len(matches) && i < start+rows; i++ {
		line := truncate(matches[i].label, m.width-10)
		if i == m.picker.cursor {
			b.WriteString(m.styles.key.Render("> ") + m.pickerLabel(matches[i], line, true) + "\n")
		} else {
			b.WriteString("  " + m.pickerLabel(matches[i], line, false) + "\n")
		}
	}
	return b.String()
}

// pickerLabel renders one row, coloring the Book name it leads with. A row that
// names no Book, or whose Book name did not survive truncation, keeps the style
// it had before: the cursor's own on the highlighted row, none otherwise.
func (m Model) pickerLabel(item pickerItem, line string, cursor bool) string {
	if item.head == "" || !strings.HasPrefix(line, item.head) {
		if cursor {
			return m.styles.key.Render(line)
		}
		return line
	}
	style := item.style
	if cursor {
		style = style.Underline(true)
	}
	return style.Render(item.head) + strings.TrimPrefix(line, item.head)
}

// referencePickerView is the Go to Reference modal: an input line, the reason the
// last input was rejected and the range it should fall in, and the Books the
// input names.
func (m Model) referencePickerView() string {
	var b strings.Builder
	b.WriteString(m.styles.label.Render(m.translate(selectReference)))
	b.WriteString("\n\n")
	b.WriteString(m.styles.label.Render(m.translate("Reference: ")) + m.picker.query)
	b.WriteString("\n\n")
	if m.picker.err != "" {
		b.WriteString(m.styles.warning.Render(m.picker.err))
		b.WriteString("\n")
	}
	hint, outOfRange := m.referenceLine()
	style := m.styles.label
	if outOfRange { // flagged while it is typed, without waiting for Enter
		style = m.styles.warning
	}
	b.WriteString(style.Render(hint))
	if rows := m.pickerRowLines(); rows != "" {
		b.WriteString("\n\n" + rows)
	}
	b.WriteString("\n" + m.styles.label.Render(m.translate(keyEnter+" to go · "+keyTab+" to complete · "+keyClose+" to close")))

	box := m.styles.help.Render(strings.TrimRight(b.String(), "\n"))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// referenceLine is what the Reference modal says the input may be — the accepted
// shapes while no Book is named, the Book's range of Chapters, or a closed
// Chapter's range of Verses — and whether the input has written a number the Book
// does not have, which the modal flags while it is typed rather than on Enter.
func (m Model) referenceLine() (line string, outOfRange bool) {
	book, location := module.SplitReference(m.picker.query)
	named, ok := m.namedBook(book)
	if !ok {
		return m.translate("For example: John, John 3, John 3:16"), false
	}

	chapters := fmt.Sprintf("%s 1..%d", named.Name, named.Chapters)
	typed := readLocation(m.picker.query, location)
	if !typed.written {
		return chapters, false
	}
	if typed.chapter < 1 || typed.chapter > named.Chapters {
		return chapters, true
	}

	verses := m.text.Chapter(named.Number, typed.chapter)
	last := 0
	if len(verses) > 0 {
		last = verses[len(verses)-1].Number
	}
	versesOfChapter := fmt.Sprintf("%s %d:1..%d", named.Name, typed.chapter, last)
	switch {
	case last == 0: // a Chapter the Module carries no Verse for is nowhere to go
		return chapters, true
	case typed.hasVerse && (typed.verse < 1 || typed.verse > last):
		return versesOfChapter, true
	case typed.closed:
		return versesOfChapter, false
	}
	return chapters, false
}

// typedLocation is how far the Reference modal's input has read into a location:
// the Chapter it wrote, the Verse after it, and whether a separator closed the
// Chapter. The zero value is an input that wrote no location at all.
type typedLocation struct {
	chapter  int
	verse    int
	written  bool // the input wrote a Chapter
	closed   bool // a separator closed that Chapter, so it cannot grow
	hasVerse bool
}

// readLocation reads the location the input has written so far. It reads past a
// separator with nothing behind it, so "3:" closes Chapter 3 while the reader is
// still on their way to a Verse, and past a separator the split cannot keep: the
// space between a Chapter and its Verse.
func readLocation(query, location string) typedLocation {
	digits := 0
	for digits < len(location) && isDigit(location[digits]) {
		digits++
	}
	chapter, err := strconv.Atoi(location[:digits])
	if err != nil {
		return typedLocation{}
	}

	typed := typedLocation{chapter: chapter, written: true}
	rest := location[digits:]
	typed.closed = rest != "" || !endsWithDigit(query)
	if typed.closed {
		if verse, err := strconv.Atoi(strings.TrimLeft(rest, ":,. ")); err == nil {
			typed.verse, typed.hasVerse = verse, true
		}
	}
	return typed
}

// endsWithDigit reports whether the input ends on a digit, the one sign that its
// Chapter is still being written.
func endsWithDigit(text string) bool {
	return text != "" && isDigit(text[len(text)-1])
}

// isDigit reports whether b is one of the ASCII digits.
func isDigit(b byte) bool { return b >= '0' && b <= '9' }

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
