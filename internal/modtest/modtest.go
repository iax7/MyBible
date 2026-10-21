// Package modtest builds synthetic MyBible Module files for tests.
//
// Fixtures are generated at test time with the same pure-Go SQLite driver the
// reader uses, so no binary module is ever committed to the repository.
package modtest

import (
	"database/sql"
	"path/filepath"
	"strconv"
	"testing"

	_ "modernc.org/sqlite"
)

// Book is one row of a synthetic Module's book table.
type Book struct {
	Number    int
	ShortName string
	LongName  string
	Present   bool // only written to books_all, which carries is_present
}

// Verse is one row of a synthetic Module's verses table.
type Verse struct {
	Book    int
	Chapter int
	Verse   int
	Text    string
}

// Create makes a Module file named filename inside dir and runs statements
// against it. It returns the Module's path.
func Create(t testing.TB, dir, filename string, statements ...string) string {
	t.Helper()

	path := filepath.Join(dir, filename)
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("modtest: open %s: %v", path, err)
	}
	defer func() { _ = db.Close() }()

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("modtest: exec %q: %v", stmt, err)
		}
	}
	return path
}

// Bible makes a Bible Module named filename inside dir with the standard tables
// and the given info key/value pairs written into its info table.
func Bible(t testing.TB, dir, filename string, info map[string]string) string {
	t.Helper()

	statements := []string{
		`CREATE TABLE info (name TEXT, value TEXT)`,
		`CREATE TABLE books (book_number NUMERIC, book_color TEXT, short_name TEXT, long_name TEXT, sorting_order NUMERIC)`,
		`CREATE TABLE verses (book_number NUMERIC, chapter NUMERIC, verse NUMERIC, text TEXT)`,
	}
	statements = append(statements, infoStatements(info)...)
	return Create(t, dir, filename, statements...)
}

// BibleBooks makes a Bible Module whose book table is books, which carries no
// is_present flag, and inserts the given Books and Verses.
func BibleBooks(t testing.TB, dir, filename string, info map[string]string, books []Book, verses []Verse) string {
	t.Helper()

	statements := []string{
		`CREATE TABLE info (name TEXT, value TEXT)`,
		`CREATE TABLE books (book_number NUMERIC, book_color TEXT, short_name TEXT, long_name TEXT, sorting_order NUMERIC)`,
		`CREATE TABLE verses (book_number NUMERIC, chapter NUMERIC, verse NUMERIC, text TEXT)`,
	}
	statements = append(statements, infoStatements(info)...)
	statements = append(statements, insertBooks("books", false, books)...)
	statements = append(statements, insertVerses(verses)...)
	return Create(t, dir, filename, statements...)
}

// BibleAll makes a Bible Module whose book table is books_all, which carries
// is_present, and inserts the given Books and Verses.
func BibleAll(t testing.TB, dir, filename string, info map[string]string, books []Book, verses []Verse) string {
	t.Helper()

	statements := []string{
		`CREATE TABLE info (name TEXT, value TEXT)`,
		`CREATE TABLE books_all (book_number NUMERIC, book_color TEXT, short_name TEXT, title TEXT, long_name TEXT, is_present BOOLEAN, sorting_order NUMERIC)`,
		`CREATE TABLE verses (book_number NUMERIC, chapter NUMERIC, verse NUMERIC, text TEXT)`,
	}
	statements = append(statements, infoStatements(info)...)
	statements = append(statements, insertBooks("books_all", true, books)...)
	statements = append(statements, insertVerses(verses)...)
	return Create(t, dir, filename, statements...)
}

// Subheading is one row of a synthetic Subheadings Module.
type Subheading struct {
	Book    int
	Chapter int
	Verse   int
	Order   int
	Text    string
}

// Subheadings makes a Subheadings Module named filename inside dir with the
// standard tables, the given info key/value pairs, and the given Subheadings.
func Subheadings(t testing.TB, dir, filename string, info map[string]string, subheadings []Subheading) string {
	t.Helper()

	statements := []string{
		`CREATE TABLE info (name TEXT, value TEXT)`,
		`CREATE TABLE subheadings (book_number NUMERIC, chapter NUMERIC, verse NUMERIC, order_if_several NUMERIC, subheading TEXT, PRIMARY KEY (book_number, chapter, verse, order_if_several))`,
	}
	statements = append(statements, infoStatements(info)...)
	for _, s := range subheadings {
		statements = append(statements, "INSERT INTO subheadings (book_number, chapter, verse, order_if_several, subheading) VALUES ("+
			itoa(s.Book)+", "+itoa(s.Chapter)+", "+itoa(s.Verse)+", "+itoa(s.Order)+", "+quote(s.Text)+")")
	}
	return Create(t, dir, filename, statements...)
}

// Commentary is one row of a synthetic Commentary Module.
type Commentary struct {
	Book        int
	ChapterFrom int
	VerseFrom   int
	ChapterTo   int
	VerseTo     int
	Text        string
	NullEnd     bool // write the end as NULL, the way some Modules do
}

// Commentaries makes a Commentary Module named filename inside dir with the
// standard tables, the given info key/value pairs, and the given rows.
func Commentaries(t testing.TB, dir, filename string, info map[string]string, rows []Commentary) string {
	t.Helper()

	statements := []string{
		`CREATE TABLE info (name TEXT, value TEXT)`,
		`CREATE TABLE commentaries (book_number NUMERIC, chapter_number_from NUMERIC, verse_number_from NUMERIC, chapter_number_to NUMERIC, verse_number_to NUMERIC, marker TEXT, text TEXT NOT NULL DEFAULT '', PRIMARY KEY (book_number, chapter_number_from, verse_number_from, marker))`,
	}
	statements = append(statements, infoStatements(info)...)
	for i, row := range rows {
		end := itoa(row.ChapterTo) + ", " + itoa(row.VerseTo)
		if row.NullEnd {
			end = "NULL, NULL"
		}
		statements = append(statements, "INSERT INTO commentaries (book_number, chapter_number_from, verse_number_from, chapter_number_to, verse_number_to, marker, text) VALUES ("+
			itoa(row.Book)+", "+itoa(row.ChapterFrom)+", "+itoa(row.VerseFrom)+", "+end+", "+quote(itoa(i))+", "+quote(row.Text)+")")
	}
	return Create(t, dir, filename, statements...)
}

// infoStatements turns the info key/value pairs into INSERT statements.
func infoStatements(info map[string]string) []string {
	statements := make([]string, 0, len(info))
	for name, value := range info {
		statements = append(statements, insertInfo(name, value))
	}
	return statements
}

// insertBooks turns Book rows into INSERT statements for the named book table.
func insertBooks(table string, presence bool, books []Book) []string {
	columns := "book_number, short_name, long_name, sorting_order"
	if presence {
		columns = "book_number, short_name, long_name, is_present, sorting_order"
	}
	statements := make([]string, 0, len(books))
	for _, b := range books {
		values := itoa(b.Number) + ", " + quote(b.ShortName) + ", " + quote(b.LongName)
		if presence {
			values += ", " + boolLiteral(b.Present)
		}
		values += ", 0"
		statements = append(statements, "INSERT INTO "+table+" ("+columns+") VALUES ("+values+")")
	}
	return statements
}

// insertVerses turns Verse rows into INSERT statements.
func insertVerses(verses []Verse) []string {
	statements := make([]string, 0, len(verses))
	for _, v := range verses {
		statements = append(statements, "INSERT INTO verses (book_number, chapter, verse, text) VALUES ("+
			itoa(v.Book)+", "+itoa(v.Chapter)+", "+itoa(v.Verse)+", "+quote(v.Text)+")")
	}
	return statements
}

// insertInfo builds an escaped INSERT for the info table.
func insertInfo(name, value string) string {
	return "INSERT INTO info (name, value) VALUES (" + quote(name) + ", " + quote(value) + ")"
}

// boolLiteral renders a bool the way SQLite stores it.
func boolLiteral(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// itoa renders an int as a SQLite numeric literal.
func itoa(n int) string {
	return strconv.Itoa(n)
}

// quote turns a string into a SQLite string literal.
func quote(s string) string {
	out := make([]byte, 0, len(s)+2)
	out = append(out, '\'')
	for i := 0; i < len(s); i++ {
		if s[i] == '\'' {
			out = append(out, '\'')
		}
		out = append(out, s[i])
	}
	return string(append(out, '\''))
}
