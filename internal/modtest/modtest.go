// Package modtest builds synthetic MyBible Module files for tests.
//
// Fixtures are generated at test time with the same pure-Go SQLite driver the
// reader uses, so no binary module is ever committed to the repository.
package modtest

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

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
	for name, value := range info {
		statements = append(statements, insertInfo(name, value))
	}
	return Create(t, dir, filename, statements...)
}

// insertInfo builds an escaped INSERT for the info table.
func insertInfo(name, value string) string {
	return "INSERT INTO info (name, value) VALUES (" + quote(name) + ", " + quote(value) + ")"
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
