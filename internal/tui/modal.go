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
	pickerCommentary
	pickerBook
	pickerChapter
	pickerReference
	pickerSearch
	pickerHistory
	pickerLog
)

// choice is the state shared by the modals. The list modals (Translation, Book,
// Chapter) filter a list and confirm a highlighted row; the Reference modal
// edits free text and keeps an error message when the input is rejected.
type choice struct {
	kind      pickerKind
	query     string
	cursor    int
	err       string
	completed bool // the Reference modal's Tab already wrote a Book: the menu stays out of the way
	width     int  // the widest the Reference modal has had to be, so it grows but never shrinks
}

func (c choice) isOpen() bool { return c.kind != pickerNone }

// pickerItem is one row of a list modal: the text to show, the text to filter
// by, and the value the row selects. head is the text that names a Book, colored
// with style, so the Book stands out from the rest of the row; short is the
// Module's own abbreviation for that Book, shown before the name and colored the
// same way. A row that is not a Book has neither.
type pickerItem struct {
	label  string
	search string
	value  int
	head   string
	short  string
	style  lipgloss.Style
	group  string // the day heading this row belongs under, when the list is grouped by day
}

// pickerSpec is everything that differs between modals: the heading, where the
// rows come from, what an empty list says, and what confirming a row does. The
// generic modal code reads this table, so adding a modal is one entry here.
type pickerSpec struct {
	title      string                         // interface string key of the heading
	items      func(m Model) []pickerItem     // the full, unfiltered row list
	cursor     func(m Model) int              // initial cursor row, or nil for the first
	onQuery    func(m Model) Model            // extra work when the typed query changes
	empty      func(m Model) string           // message when no row matches, or nil for the default
	confirm    func(m Model, value int) Model // apply the highlighted row
	filter     bool                           // the typed query narrows the rows
	takesInput bool                           // the modal takes a query though it does not filter its own rows
	typedOnly  bool                           // the rows appear only once the query names something
}

// bookItems are the Active translation's Books as modal rows, in Canonical
// order. The Select Book and Go to Reference modals list the same rows; only
// what confirming one does differs.
func bookItems(m Model) []pickerItem {
	books := m.books()
	width := 0
	for _, b := range books {
		if w := lipgloss.Width(b.ShortName); w > width {
			width = w
		}
	}
	items := make([]pickerItem, 0, len(books))
	for _, b := range books {
		items = append(items, pickerItem{
			label:  bookLabel(b, width),
			search: strings.Join(b.Names(), " "),
			value:  b.Number,
			head:   b.Name,
			short:  b.ShortName,
			style:  m.styles.book(b.Number),
		})
	}
	return items
}

// bookSeparator is the quiet mark between a Book's abbreviation and its name: a
// middle dot, dark grey, so the two names of the same Book read as a pair.
const bookSeparator = " · "

// bookLabel is the row a Book shows in a list: the Module's own abbreviation
// padded so every name starts at the same column, the separator, and the
// canonical name — "Gn   · Genesis". A Module that names the Book only in full
// shows the name alone.
func bookLabel(b module.Book, width int) string {
	if b.ShortName == "" {
		return b.Name
	}
	return pad(b.ShortName, width) + bookSeparator + b.Name
}

// bookParts splits a Book's row into its abbreviation and its name, the
// separator between them already gone. A row without an abbreviation, or one
// whose separator did not survive truncation, has no parts.
func bookParts(item pickerItem, line string) (abbrev, name string, ok bool) {
	if item.short == "" {
		return "", "", false
	}
	if i := strings.Index(line, bookSeparator); i >= 0 {
		return line[:i], line[i+len(bookSeparator):], true
	}
	return "", "", false
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
	pickerCommentary: {
		title:  selectCommentary,
		filter: true,
		items: func(m Model) []pickerItem {
			items := make([]pickerItem, 0, len(m.commentaries))
			for i, commentary := range m.commentaries {
				items = append(items, pickerItem{
					label:  pad(commentary.Abbreviation, 10) + commentary.Title,
					search: commentary.Abbreviation + " " + commentary.Title,
					value:  i,
				})
			}
			return items
		},
		cursor: func(m Model) int {
			for i, commentary := range m.commentaries {
				if commentary.Abbreviation == m.activeCommentary {
					return i
				}
			}
			return 0
		},
		empty: func(m Model) string { return m.translate("No commentaries") },
		confirm: func(m Model, value int) Model {
			if value >= 0 && value < len(m.commentaries) {
				m.activeCommentary = m.commentaries[value].Abbreviation
				m = m.loadCommentary()
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
		title:      searchTitle,
		takesInput: true,
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
				reference := m.history[i].Reference
				items = append(items, pickerItem{
					label:  reference.String(),
					search: reference.String(),
					value:  i,
					head:   reference.Book.Name,
					style:  m.styles.book(reference.Book.Number),
					group:  m.historyDay(m.history[i].At),
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
	},
	pickerLog: {
		title: logTitle,
		items: func(m Model) []pickerItem {
			items := make([]pickerItem, 0, len(m.warnings))
			for i, warning := range m.warnings {
				items = append(items, pickerItem{label: "• " + warning, search: warning, value: i})
			}
			return items
		},
		empty: func(m Model) string { return m.translate("No warnings") },
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
	if kind == pickerReference {
		m.picker.width = m.referenceNeeds()
	}
	return m
}

// updatePicker handles keystrokes while a modal is open: typing filters the
// list, Enter confirms, Tab completes the Reference modal, Esc closes. Every
// modal takes the same keys.
func (m Model) updatePicker(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	types := m.pickerTypes()
	switch msg.String() {
	case keyClose:
		m.picker = choice{}
	case keyEnter:
		m = m.pickerConfirm()
	case keyTab:
		m = m.pickerComplete()
	case keyBackspace:
		if types {
			before := m.pickerFilterQuery()
			if runes := []rune(m.picker.query); len(runes) > 0 {
				m.picker.query = string(runes[:len(runes)-1])
			}
			m.picker.err = ""
			m = m.refilter(before)
		}
	case keyUp:
		m.picker.cursor--
	case keyDown:
		m.picker.cursor++
	default:
		if types && msg.Text != "" {
			before := m.pickerFilterQuery()
			m.picker.query += msg.Text
			m.picker.err = ""
			m = m.refilter(before)
		}
	}
	if m.picker.isOpen() {
		m.picker.clamp(len(m.pickerMatches()))
		if m.picker.kind == pickerReference {
			m.picker.width = max(m.picker.width, m.referenceNeeds())
		}
	}
	return m, nil
}

// pickerTypes reports whether keystrokes go into the open modal's query. The Log
// just lists warnings: it has nothing to type.
func (m Model) pickerTypes() bool {
	spec := pickers[m.picker.kind]
	return spec.filter || spec.takesInput
}

// refilter recomputes a modal's rows after its query changed, and returns the
// cursor to the first row when the text the rows are filtered by changed, so a
// narrowed list starts at its top. The Reference modal filters by the Book its
// input names, so a Chapter typed after the Book leaves the highlight alone; it
// also brings the completion menu back, since a fresh Book name is a fresh ask
// for it.
func (m Model) refilter(before string) Model {
	if before != m.pickerFilterQuery() {
		m.picker.cursor = 0
		m.picker.completed = false
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
	if m.picker.kind == pickerHistory {
		return m.confirmHistory()
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
	m.picker.completed = true // the menu did its job; it comes back only on a fresh name
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
	_, location := module.SplitReference(m.picker.query)
	text := m.picker.query
	if named, ok := m.resolvedBook(); ok {
		text = strings.TrimSpace(named.Name + " " + location)
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
// letter case and accents, in Canonical order. The Search modal is already
// matched against the Active translation, so its items pass through unfiltered.
// A Book whose abbreviation is exactly what is typed floats to the top, so "ez"
// lands on Ezekiel before Ezra.
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
	for i, item := range matches {
		if item.short == "" || module.Normalize(item.short) != query {
			continue
		}
		if i > 0 { // move it to the front, keeping the rest in order
			copy(matches[1:i+1], matches[:i])
			matches[0] = item
		}
		break
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

// pickerFilterLine is the line a list modal types into: the chevron on the
// theme's filled band that the Go to Reference modal uses, so every modal says
// where typing lands the same way.
func (m Model) pickerFilterLine() string {
	field := m.styles.field.Render("❯ " + m.picker.query)
	if pad := fieldMinWidth - lipgloss.Width(field); pad > 0 {
		field += m.styles.field.Render(strings.Repeat(" ", pad))
	}
	return field
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
	if m.pickerTypes() {
		b.WriteString(m.pickerFilterLine())
		b.WriteString("\n\n")
	}
	b.WriteString(m.pickerRowLines())
	if m.picker.err != "" {
		b.WriteString("\n" + m.styles.warning.Render(m.picker.err))
	}
	b.WriteString("\n" + m.styles.hint.Render(m.translate(keyEnter+" to select")))

	return m.styles.frameBox(m.pickerTitle(), keyClose, strings.TrimRight(b.String(), "\n"))
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
	if m.picker.kind == pickerHistory {
		return m.historyRowLines(matches)
	}

	var b strings.Builder
	rows, start := m.pickerRows(), windowStart(m.picker.cursor, len(matches), m.pickerRows())
	for i := start; i < len(matches) && i < start+rows; i++ {
		b.WriteString(m.pickerRow(matches[i], i == m.picker.cursor) + "\n")
	}
	return b.String()
}

// historyRowLines renders the History rows with a day heading above the first
// visible Visit of each day. Headings take lines too, so the window is measured
// in lines rather than rows; the cursor's own row always stays in view.
func (m Model) historyRowLines(matches []pickerItem) string {
	start, end := m.historyWindow(matches)
	var b strings.Builder
	last := ""
	for i := start; i < end; i++ {
		if matches[i].group != last {
			b.WriteString(m.styles.label.Render(truncate(matches[i].group, m.width-10)) + "\n")
			last = matches[i].group
		}
		b.WriteString(m.pickerRow(matches[i], i == m.picker.cursor) + "\n")
	}
	return b.String()
}

// historyWindow is the first and past-last row of History that fit above the
// modal's frame, day headings included, around the cursor.
func (m Model) historyWindow(matches []pickerItem) (int, int) {
	budget := m.pickerRows()
	start, end := m.picker.cursor, m.picker.cursor+1
	for start > 0 && historyLines(matches, start-1, end) <= budget {
		start--
	}
	for end < len(matches) && historyLines(matches, start, end+1) <= budget {
		end++
	}
	return start, end
}

// historyLines is how many screen lines the rows in [start, end) need, counting
// a heading wherever the day changes.
func historyLines(matches []pickerItem, start, end int) int {
	lines, last := 0, ""
	for i := start; i < end; i++ {
		if matches[i].group != last {
			lines++
			last = matches[i].group
		}
		lines++
	}
	return lines
}

// pickerRow is one rendered row, marked when it is the cursor's.
func (m Model) pickerRow(item pickerItem, cursor bool) string {
	line := truncate(item.label, m.width-10)
	if cursor {
		return m.styles.key.Render("> ") + m.pickerLabel(item, line, true)
	}
	return "  " + m.pickerLabel(item, line, false)
}

// pickerLabel renders one row. A Book row colors its abbreviation and its name
// with the Book's classification color, leaving the quiet separator between them
// in the theme's hint grey; any other row colors the Book name it carries. A row
// whose Book name did not survive truncation keeps the style it had before: the
// cursor's own on the highlighted row, none otherwise.
func (m Model) pickerLabel(item pickerItem, line string, cursor bool) string {
	style := item.style
	if cursor {
		style = style.Underline(true)
	}
	if abbrev, name, ok := bookParts(item, line); ok {
		return style.Render(abbrev) + m.styles.hint.Render(bookSeparator) + style.Render(name)
	}
	at := -1
	if item.head != "" {
		at = strings.Index(line, item.head)
	}
	if at < 0 {
		if cursor {
			return m.styles.key.Render(line)
		}
		return line
	}
	return line[:at] + style.Render(item.head) + line[at+len(item.head):]
}

// The Reference modal's measurements: the input band is wide enough to type into
// even when the range line is short; the completion menu stays a menu rather than
// a list, and is indented so its chevron lines up with the input field's.
const (
	fieldMinWidth    = 30
	popupRows        = 8
	popupWidth       = 40
	pickerPopupInset = 2
)

// referencePickerView is the Go to Reference modal: the input line the reader
// types into, a filled band marked with a chevron so it is obvious where typing
// lands, and, above it, the reason the last input was rejected or the range the
// input may use. The Books the input names are not part of the box: they float
// under the input as a menu (pickerPopupView), so a Book is never named three
// times in the same view.
func (m Model) referencePickerView() string {
	status, _ := m.referenceStatus()
	state := m.referenceState()
	footer := m.translate(keyEnter + " to go · " + keyTab + " to complete")
	width := max(m.picker.width, m.referenceNeeds())

	field := m.styles.field.Render("❯ ") + m.referenceInput(state)
	if pad := width - lipgloss.Width(field); pad > 0 {
		field += m.styles.field.Render(strings.Repeat(" ", pad))
	}

	body := strings.Join(status, "\n") + "\n" + field + "\n" + m.styles.hint.Render(footer)
	return m.styles.frameBox(m.translate(selectReference), keyClose, body)
}

// referenceStatus is the Reference modal's lines above the input — the reason the
// last input was rejected, and the range the input may use — with the width the
// widest of them needs to sit on the input's band.
func (m Model) referenceStatus() ([]string, int) {
	state := m.referenceState()
	lines := make([]string, 0, 2)
	if m.picker.err != "" {
		lines = append(lines, m.styles.warning.Render(m.picker.err))
	}
	lines = append(lines, m.styles.label.Render(state.rangeText))

	width := fieldMinWidth
	for _, line := range lines {
		if w := lipgloss.Width(line); w > width {
			width = w
		}
	}
	return lines, width
}

// referenceNeeds is the width the Reference modal's current content wants: the
// widest of its status lines, its input, and its footer. The modal keeps the
// widest it has needed (choice.width), so it grows to fit a line and never
// shrinks back under the reader's eyes.
func (m Model) referenceNeeds() int {
	_, width := m.referenceStatus()
	return max(width,
		lipgloss.Width("❯ ")+lipgloss.Width(m.picker.query),
		lipgloss.Width(m.translate(keyEnter+" to go · "+keyTab+" to complete")))
}

// referenceInput is the input as typed on the field's own background, with the
// number the Book has no Chapter or Verse for picked out, whole, in red.
func (m Model) referenceInput(state referenceState) string {
	if state.badStart < 0 || state.badEnd > len(m.picker.query) {
		return m.styles.field.Render(m.picker.query)
	}
	query := m.picker.query
	return m.styles.field.Render(query[:state.badStart]) +
		m.styles.invalid.Render(query[state.badStart:state.badEnd]) +
		m.styles.field.Render(query[state.badEnd:])
}

// pickerPopupView is the Reference modal's completion menu: the Books its input
// names, floating under the input. It is empty for every other modal, once Tab
// has completed a Book, before anything is typed, and when nothing matches.
func (m Model) pickerPopupView() string {
	if m.picker.kind != pickerReference || m.picker.completed || m.pickerFilterQuery() == "" {
		return ""
	}
	matches := m.pickerMatches()
	if len(matches) == 0 {
		return ""
	}

	rows := m.pickerRows()
	if rows > popupRows {
		rows = popupRows
	}
	start := windowStart(m.picker.cursor, len(matches), rows)

	width := 0
	for i := start; i < len(matches) && i < start+rows; i++ {
		if w := lipgloss.Width(matches[i].label); w > width {
			width = w
		}
	}
	if width > popupWidth {
		width = popupWidth
	}

	var b strings.Builder
	for i := start; i < len(matches) && i < start+rows; i++ {
		b.WriteString(m.popupRow(matches[i], i == m.picker.cursor, width))
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// popupRow paints one completion row on the menu's background: a Book's
// abbreviation and name in the Book's classification color with the quiet
// separator between them, or any other row's Book name the same way, and the
// cursor's row picked out with a chevron and the key color.
func (m Model) popupRow(item pickerItem, cursor bool, width int) string {
	marker, base := "  ", m.styles.popup
	if cursor {
		marker, base = "> ", m.styles.popupKey
	}
	book := item.style.Background(lipgloss.Color(m.styles.barColor))

	row := base.Render(" " + marker)
	label := truncate(item.label, width)
	if abbrev, name, ok := bookParts(item, label); ok {
		row += book.Render(abbrev) + m.styles.hint.Background(lipgloss.Color(m.styles.barColor)).Render(bookSeparator) + book.Render(name)
	} else if at := strings.Index(label, item.head); at >= 0 {
		row += base.Render(label[:at]) + book.Render(item.head) + base.Render(label[at+len(item.head):])
	} else {
		row += base.Render(label)
	}

	if pad := width + 4 - lipgloss.Width(row); pad > 0 {
		row += m.styles.popup.Render(strings.Repeat(" ", pad))
	}
	return row
}

// referenceState is what the Reference modal knows about its input: the range the
// input may use, and, when the input has written a number the Active translation
// has no Chapter or Verse for, where that number sits in the input so the modal
// can pick it out while it is typed. badStart is -1 when every number written is
// one the Book has.
type referenceState struct {
	rangeText        string
	badStart, badEnd int
}

// referenceState reads the input and the Book it resolves to, once, for both the
// range line and the number it may flag.
func (m Model) referenceState() referenceState {
	shapes := referenceState{
		rangeText: m.translate("For example: John, John 3, John 3:16"),
		badStart:  -1,
	}
	book, location := module.SplitReference(m.picker.query)
	typed := readLocation(m.picker.query, location)
	if book == "" && !typed.written { // nothing said yet: no Book, no number
		return shapes
	}
	named, ok := m.resolvedBook()
	if !ok {
		return shapes
	}

	chapters := fmt.Sprintf("%s 1..%d", named.Name, named.Chapters)
	state := referenceState{rangeText: chapters, badStart: -1}
	if !typed.written {
		return state
	}
	runs := lastNumbers(m.picker.query)

	if typed.chapter < 1 || typed.chapter > named.Chapters {
		return state.flag(runs, chapterNumber)
	}

	verses := m.text.Chapter(named.Number, typed.chapter)
	last := 0
	if len(verses) > 0 {
		last = verses[len(verses)-1].Number
	}
	if last == 0 { // a Chapter the Module carries no Verse for is nowhere to go
		return state.flag(runs, chapterNumber)
	}

	if typed.closed { // a Chapter still being typed may still grow into another
		state.rangeText = fmt.Sprintf("%s %d:1..%d", named.Name, typed.chapter, last)
	}
	if typed.hasVerse && (typed.verse < 1 || typed.verse > last) {
		return state.flag(runs, verseNumber)
	}
	return state
}

// The numbers an input writes, in the order lastNumbers reads them.
const (
	chapterNumber = 0
	verseNumber   = 1
)

// flag points the state at one of the numbers the input wrote, picked out of the
// spans lastNumbers read.
func (s referenceState) flag(runs [][2]int, which int) referenceState {
	if which < len(runs) {
		s.badStart, s.badEnd = runs[which][0], runs[which][1]
	}
	return s
}

// lastNumbers is the span of each run of digits the input ends with, the Chapter
// first and the Verse last: "Genesis 3:16" gives the spans of 3 and 16, while
// "1 Samuel 3" gives only the 3, because the Book's own number is not a location.
func lastNumbers(text string) [][2]int {
	text = strings.TrimRight(text, " ")
	var runs [][2]int
	for end := len(text); end > 0; {
		start := end
		for start > 0 && isDigit(text[start-1]) {
			start--
		}
		if start == end {
			break // no digit behind: the input ends on something else
		}
		if start > 0 && !strings.ContainsRune(":,. ", rune(text[start-1])) {
			break // a letter, so what is left is the Book's name
		}
		runs = append([][2]int{{start, end}}, runs...)
		end = start
		for end > 0 && strings.ContainsRune(":,. ", rune(text[end-1])) {
			end--
		}
	}
	return runs
}

// resolvedBook is the Book the input reads as: the one it names, or the
// translation's first Book — the highlighted row — when it names none.
func (m Model) resolvedBook() (module.Book, bool) {
	book, _ := module.SplitReference(m.picker.query)
	if book != "" {
		return m.namedBook(book)
	}
	items := m.pickerMatches()
	if m.picker.cursor < 0 || m.picker.cursor >= len(items) {
		return module.Book{}, false
	}
	return m.text.Book(items[m.picker.cursor].value)
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
