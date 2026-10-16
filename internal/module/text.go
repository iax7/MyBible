package module

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// Text is one Translation's complete text, loaded into memory. The reader
// navigates and searches this copy, so there are no per-scroll queries.
type Text struct {
	books    []bookText
	index    map[int]int                  // Book number -> position in books, so lookup is not a walk
	headings map[int]map[int][]Subheading // Book number -> Chapter -> Subheadings, from the companion Module
}

// bookText is one Book of a loaded Translation, with its Chapters indexed by
// Chapter number (index 0 is unused).
type bookText struct {
	Book
	chapters [][]Verse
}

// Subheading is one section heading of a Chapter: the Verse it introduces, and
// its text with inline markup rendered away.
type Subheading struct {
	Verse int
	Text  string
}

// Verse is one numbered unit of text within a Chapter. Text keeps the Module's
// inline markup; ParseMarkup turns it into display Runs. normalized is the
// display text lowercased and stripped of accents, built once at load time so
// Search needs no case- or accent-insensitive SQL.
type Verse struct {
	Number     int
	Text       string
	normalized string
}

// Load reads the whole Translation into memory at once (a Bible is roughly
// 31,000 Verses). Books are mapped into the Canonical book list by book_number,
// honouring is_present and tolerating legacy book numbers. Books with no Verse
// rows are left out.
func (t Translation) Load() (*Text, error) {
	db, err := openReadOnly(t.path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Close() }()

	tables, err := moduleTables(db)
	if err != nil {
		return nil, err
	}
	books, err := moduleBooks(db, tables)
	if err != nil {
		return nil, err
	}

	byNumber := make(map[int]*bookText)
	rows, err := db.Query("SELECT book_number, chapter, verse, text FROM verses ORDER BY book_number, chapter, verse")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var bookNumber, chapter int
		var verse Verse
		var text sql.NullString
		if err := rows.Scan(&bookNumber, &chapter, &verse.Number, &text); err != nil {
			return nil, err
		}
		book, ok := books[bookNumber]
		if !ok || chapter < 1 {
			continue
		}
		entry := byNumber[book.Number]
		if entry == nil {
			entry = &bookText{Book: book}
			byNumber[book.Number] = entry
		}
		for len(entry.chapters) <= chapter {
			entry.chapters = append(entry.chapters, nil)
		}
		verse.Text = text.String
		verse.normalized = searchable(verse.Text)
		entry.chapters[chapter] = append(entry.chapters[chapter], verse)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	text := &Text{books: make([]bookText, 0, len(byNumber))}
	for _, book := range byNumber {
		for chapter := len(book.chapters) - 1; chapter > 0; chapter-- {
			if len(book.chapters[chapter]) > 0 {
				book.Chapters = chapter
				break
			}
		}
		if book.Chapters > 0 {
			text.books = append(text.books, *book)
		}
	}
	sort.Slice(text.books, func(i, j int) bool {
		return text.books[i].Number < text.books[j].Number
	})
	text.index = make(map[int]int, len(text.books))
	for i, book := range text.books {
		text.index[book.Number] = i
	}
	if t.subheadings != "" {
		// A companion Module that cannot be read is left out: it adds headings,
		// it is not what the reader came to read.
		//
		// shortcut: a broken Subheadings Module is ignored in silence; report it
		// once the reader has somewhere to report a companion's failure.
		text.headings, _ = loadSubheadings(t.subheadings)
	}
	return text, nil
}

// Books returns the loaded Translation's Books in Canonical order.
func (t *Text) Books() []Book {
	books := make([]Book, len(t.books))
	for i, b := range t.books {
		books[i] = b.Book
	}
	return books
}

// Book returns the loaded Book with the given Canonical number.
func (t *Text) Book(number int) (Book, bool) {
	if i, ok := t.index[number]; ok {
		return t.books[i].Book, true
	}
	return Book{}, false
}

// BookIndex returns the position of a Book in the loaded Translation's Canonical
// order, or -1 when it is not loaded.
func (t *Text) BookIndex(number int) int {
	if i, ok := t.index[number]; ok {
		return i
	}
	return -1
}

// Chapter returns the Verses of one Chapter in Verse order.
func (t *Text) Chapter(number, chapter int) []Verse {
	i, ok := t.index[number]
	if !ok {
		return nil
	}
	if chapter > 0 && chapter < len(t.books[i].chapters) {
		return t.books[i].chapters[chapter]
	}
	return nil
}

// Subheadings returns the Subheadings of one Chapter, in the order the companion
// Module writes them. A Translation with no companion returns none.
func (t *Text) Subheadings(number, chapter int) []Subheading {
	if chapter < 1 {
		return nil
	}
	return t.headings[number][chapter]
}

// loadSubheadings reads a Subheadings Module into memory, keyed by Canonical
// Book and Chapter. The Module carries no book table, so its own book numbers are
// mapped into the Canonical list on the way in, the way a Bible's are.
func loadSubheadings(path string) (map[int]map[int][]Subheading, error) {
	db, err := openReadOnly(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Close() }()

	tables, err := moduleTables(db)
	if err != nil {
		return nil, err
	}
	if !tables[tableSubheadings] {
		return nil, fmt.Errorf("missing the required %s table", tableSubheadings)
	}

	rows, err := db.Query("SELECT book_number, chapter, verse, subheading FROM subheadings ORDER BY book_number, chapter, verse, order_if_several")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	headings := make(map[int]map[int][]Subheading)
	for rows.Next() {
		var number, chapter, verse int
		var text sql.NullString
		if err := rows.Scan(&number, &chapter, &verse, &text); err != nil {
			return nil, err
		}
		book, ok := canonicalBook(number)
		if !ok || chapter < 1 {
			continue
		}
		heading := strings.TrimSpace(displayText(text.String))
		if !hasLetterOrDigit(heading) {
			continue // a Subheading that is nothing but a cross-reference says nothing
		}
		if headings[book.Number] == nil {
			headings[book.Number] = make(map[int][]Subheading)
		}
		headings[book.Number][chapter] = append(headings[book.Number][chapter], Subheading{Verse: verse, Text: heading})
	}
	return headings, rows.Err()
}

// hasLetterOrDigit reports whether text holds anything but punctuation and spaces.
func hasLetterOrDigit(text string) bool {
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// moduleBooks reads the Module's book table and returns the Canonical Books it
// holds, keyed by the book_number the Module writes them under, each carrying
// the Module's own short and long names. books_all carries is_present; a book
// table without it lists only present Books.
func moduleBooks(db *sql.DB, tables map[string]bool) (map[int]Book, error) {
	table, presence := tableBooks, false
	if tables[tableBooksAll] {
		table, presence = tableBooksAll, true
	}
	columns := "book_number, short_name, long_name"
	if presence {
		columns += ", is_present"
	}

	rows, err := db.Query("SELECT " + columns + " FROM " + table)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	books := make(map[int]Book)
	for rows.Next() {
		var number int
		var short, long sql.NullString
		if presence {
			var isPresent sql.NullBool
			if err := rows.Scan(&number, &short, &long, &isPresent); err != nil {
				return nil, err
			}
			if !isPresent.Valid || !isPresent.Bool {
				continue
			}
		} else if err := rows.Scan(&number, &short, &long); err != nil {
			return nil, err
		}
		book, ok := canonicalBook(number)
		if !ok {
			continue
		}
		book.ShortName, book.LongName = short.String, long.String
		books[number] = book
	}
	return books, rows.Err()
}

// Run is a stretch of Verse text sharing one display style. A break tag becomes
// a Run whose Text is "\n" for <br/> or "\n\n" for <pb/>.
type Run struct {
	Text    string
	Italics bool
	Red     bool
}

// strippedTags enclose content that 0.0.2 hides: Strong's numbers, morphology,
// footnotes, subheadings, and cross-reference markers.
var strippedTags = map[string]bool{"s": true, "m": true, "f": true, "h": true, "x": true}

// ParseMarkup turns a Verse's raw text into display Runs: <i> becomes italics,
// <J> becomes the red words-of-Jesus style, <pb/> and <br/> become breaks. All
// other tags are stripped, along with the content of <S>, <m>, <f>, <h>, and <x>.
func ParseMarkup(text string) []Run {
	var runs []Run
	var italics, red bool
	hidden := "" // the stripped tag whose content is being skipped

	for i := 0; i < len(text); {
		if text[i] != '<' {
			end := strings.IndexByte(text[i:], '<')
			if end < 0 {
				end = len(text) - i
			}
			if hidden == "" {
				runs = append(runs, Run{Text: text[i : i+end], Italics: italics, Red: red})
			}
			i += end
			continue
		}

		end := strings.IndexByte(text[i:], '>')
		if end < 0 { // an unmatched '<' is literal text
			if hidden == "" {
				runs = append(runs, Run{Text: text[i:], Italics: italics, Red: red})
			}
			break
		}
		raw := text[i+1 : i+end]
		i += end + 1

		closing := strings.HasPrefix(raw, "/")
		name := strings.ToLower(strings.Trim(raw, "/ "))
		if hidden != "" {
			if closing && name == hidden {
				hidden = ""
			}
			continue
		}
		switch {
		case name == "pb":
			runs = append(runs, Run{Text: "\n\n"})
		case name == "br":
			runs = append(runs, Run{Text: "\n"})
		case name == "i":
			italics = !closing
		case name == "j":
			red = !closing
		case strippedTags[name] && !closing:
			hidden = name
		}
	}
	return runs
}
