package module

import (
	"database/sql"
	"sort"
	"strings"
)

// Text is one Translation's complete text, loaded into memory. The reader
// navigates and searches this copy, so there are no per-scroll queries.
type Text struct {
	books []bookText
}

// bookText is one Book of a loaded Translation, with its Chapters indexed by
// Chapter number (index 0 is unused).
type bookText struct {
	Book
	chapters [][]Verse
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
	present, err := presentBookNumbers(db, tables)
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
		book, ok := canonicalBook(bookNumber)
		if !ok || !present[book.Number] || chapter < 1 {
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
	for _, b := range t.books {
		if b.Number == number {
			return b.Book, true
		}
	}
	return Book{}, false
}

// Chapter returns the Verses of one Chapter in Verse order.
func (t *Text) Chapter(number, chapter int) []Verse {
	for _, b := range t.books {
		if b.Number == number {
			if chapter > 0 && chapter < len(b.chapters) {
				return b.chapters[chapter]
			}
			return nil
		}
	}
	return nil
}

// presentBookNumbers reads the Module's book table and returns the Canonical
// numbers of the Books it marks present. books_all carries is_present; a book
// table without it lists only present Books.
func presentBookNumbers(db *sql.DB, tables map[string]bool) (map[int]bool, error) {
	table, presence := tableBooks, false
	if tables[tableBooksAll] {
		table, presence = tableBooksAll, true
	}
	columns := "book_number"
	if presence {
		columns += ", is_present"
	}

	rows, err := db.Query("SELECT " + columns + " FROM " + table)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	present := make(map[int]bool)
	for rows.Next() {
		var number int
		if presence {
			var isPresent sql.NullBool
			if err := rows.Scan(&number, &isPresent); err != nil {
				return nil, err
			}
			if !isPresent.Valid || !isPresent.Bool {
				continue
			}
		} else if err := rows.Scan(&number); err != nil {
			return nil, err
		}
		if book, ok := canonicalBook(number); ok {
			present[book.Number] = true
		}
	}
	return present, rows.Err()
}

// Run is a stretch of Verse text sharing one display style. A break tag becomes
// a Run whose Text is "\n" for <br/> or "\n\n" for <pb/>.
type Run struct {
	Text    string
	Italics bool
	Red     bool
}

// strippedTags enclose content that 0.0.1 hides: Strong's numbers, morphology,
// footnotes, and subheadings.
var strippedTags = map[string]bool{"s": true, "m": true, "f": true, "h": true}

// ParseMarkup turns a Verse's raw text into display Runs: <i> becomes italics,
// <J> becomes the red words-of-Jesus style, <pb/> and <br/> become breaks. All
// other tags are stripped, along with the content of <S>, <m>, <f>, and <h>.
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
