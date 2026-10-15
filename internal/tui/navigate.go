package tui

import (
	"strconv"
	"strings"

	"github.com/iax7/mybible/internal/module"
)

// chapterLines is the current Chapter rendered as screen lines, together with
// the line span of the cursor Verse.
type chapterLines struct {
	lines []string
	from  int // first line of the cursor Verse
	to    int // last line of the cursor Verse, inclusive
}

// books is the Active translation's Books in Canonical order.
func (m Model) books() []module.Book {
	if m.text == nil {
		return nil
	}
	return m.text.Books()
}

// chapterVerses is the current Chapter's Verses in Verse order.
func (m Model) chapterVerses() []module.Verse {
	if m.text == nil {
		return nil
	}
	return m.text.Chapter(m.book.Number, m.chapter)
}

// bookIndex is the position of a Book in the Active translation, or 0 when it
// is not loaded.
func (m Model) bookIndex(number int) int {
	for i, b := range m.books() {
		if b.Number == number {
			return i
		}
	}
	return 0
}

// renderLines renders every Verse of the current Chapter, one rendered line per
// Verse line, remembering where the cursor Verse sits. The cursor Verse's number
// carries the cursor style.
func (m Model) renderLines() chapterLines {
	var out chapterLines
	for i, verse := range m.chapterVerses() {
		if i == m.cursor {
			out.from = len(out.lines)
		}
		for j, line := range strings.Split(renderMarkup(verse.Text), "\n") {
			if j == 0 {
				style := verseNumStyle
				if i == m.cursor {
					style = cursorNumStyle
				}
				line = style.Render(strconv.Itoa(verse.Number)) + " " + line
			}
			out.lines = append(out.lines, line)
		}
		if i == m.cursor {
			out.to = len(out.lines) - 1
		}
	}
	return out
}

// window is the lines to display and the first-line offset, clamped so the
// cursor Verse stays inside a height-line view.
func (c chapterLines) window(top, height int) ([]string, int) {
	if height <= 0 || len(c.lines) == 0 {
		return nil, 0
	}
	top = c.clampTop(top, height)
	end := top + height
	if end > len(c.lines) {
		end = len(c.lines)
	}
	return c.lines[top:end], top
}

// clampTop keeps the cursor Verse inside a height-line window.
func (c chapterLines) clampTop(top, height int) int {
	if c.from < top {
		top = c.from
	}
	if c.to >= top+height {
		top = c.to - height + 1
	}
	if last := len(c.lines) - height; top > last {
		top = last
	}
	if top < 0 {
		top = 0
	}
	return top
}

// ensureVisible re-clamps the Chapter view so the cursor Verse is on screen.
func (m Model) ensureVisible() Model {
	_, m.top = m.renderLines().window(m.top, m.paneHeight())
	return m
}

// chapterView renders the visible part of the current Chapter.
func (m Model) chapterView(height int) string {
	lines, _ := m.renderLines().window(m.top, height)
	return strings.Join(lines, "\n")
}

// pageSize is the number of Verses a page moves by: the visible height.
func (m Model) pageSize() int {
	if height := m.paneHeight(); height > 1 {
		return height
	}
	return 1
}

// setCursor moves the Verse cursor within the Chapter, clamping to its ends.
func (m Model) setCursor(cursor int) Model {
	verses := m.chapterVerses()
	if len(verses) == 0 {
		return m
	}
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= len(verses) {
		cursor = len(verses) - 1
	}
	m.cursor = cursor
	return m.ensureVisible()
}

// moveCursor moves the Verse cursor by delta Verses.
func (m Model) moveCursor(delta int) Model { return m.setCursor(m.cursor + delta) }

// moveChapter moves delta Chapters, crossing Book boundaries at the ends and
// staying put at the very beginning and end of the Canonical order.
func (m Model) moveChapter(delta int) Model {
	books := m.books()
	i := -1
	for j, b := range books {
		if b.Number == m.book.Number {
			i = j
			break
		}
	}
	if i < 0 {
		return m
	}

	chapter := m.chapter + delta
	for {
		if chapter >= 1 && chapter <= books[i].Chapters {
			m.book, m.chapter, m.cursor, m.top = books[i], chapter, 0, 0
			return m
		}
		i += delta
		if i < 0 || i >= len(books) {
			return m
		}
		if delta < 0 {
			chapter = books[i].Chapters
		} else {
			chapter = 1
		}
	}
}

// gotoBook points the reader at the first Chapter of a Book.
func (m Model) gotoBook(number int) Model {
	book, ok := m.text.Book(number)
	if !ok {
		return m
	}
	m.book, m.chapter, m.cursor, m.top = book, 1, 0, 0
	return m
}

// gotoChapter points the reader at the first Verse of a Chapter.
func (m Model) gotoChapter(chapter int) Model {
	if chapter < 1 || chapter > m.book.Chapters {
		return m
	}
	m.chapter, m.cursor, m.top = chapter, 0, 0
	return m
}

// jumpTo points the reader at a Reference, reporting whether the Active
// translation actually contains it.
func (m Model) jumpTo(reference module.Reference) (Model, bool) {
	if m.text == nil {
		return m, false
	}
	book, ok := m.text.Book(reference.Book.Number)
	if !ok || reference.Chapter < 1 || reference.Chapter > book.Chapters {
		return m, false
	}
	verses := m.text.Chapter(book.Number, reference.Chapter)
	cursor := 0
	if reference.Verse > 0 {
		found := false
		for i, verse := range verses {
			if verse.Number == reference.Verse {
				cursor, found = i, true
				break
			}
		}
		if !found {
			return m, false
		}
	}
	m.book, m.chapter, m.cursor, m.top = book, reference.Chapter, cursor, 0
	return m.ensureVisible(), true
}
