// Package module discovers and reads MyBible Modules from the Module directory.
//
// Modules are opened read-only and never modified. The package reads the
// filename to tell the Module type and the info table for the human title.
package module

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

// Type is a Module type, encoded as a suffix in its file name.
type Type string

// TypeBible is the Module type holding a complete Bible text; it is the only
// type the reader surfaces, and the only one with no file name suffix.
const TypeBible Type = "Bible"

// TypeSubheadings is the Module type holding a Bible's Subheadings. It is never
// surfaced as a Translation; it is read as a companion of the Bible it belongs to.
const TypeSubheadings Type = "subheadings"

const (
	extension        = ".SQLite3"
	tableInfo        = "info"
	tableBooks       = "books"
	tableBooksAll    = "books_all"
	tableStories     = "stories"
	tableVerses      = "verses"
	tableSubheadings = "subheadings"
)

// Translation is a Module of type Bible.
type Translation struct {
	Abbreviation string // the file name stem, which exists nowhere inside the Module
	Title        string // the human title, from info.description
	Language     string // the two-letter language, from info.language
	path         string
	subheadings  string // path of the companion Subheadings Module, empty when there is none
}

// TypeOf reports the Module type encoded in a Module file name.
func TypeOf(name string) Type {
	stem := strings.TrimSuffix(name, extension)
	if i := strings.LastIndex(stem, "."); i >= 0 {
		return Type(stem[i+1:])
	}
	return TypeBible
}

// Discover scans dir for Modules and returns its Bible Modules as Translations,
// sorted by abbreviation. Non-Bible Module types are ignored, except a
// Subheadings Module, which is paired with the Bible whose abbreviation its file
// name carries; the pairing is a companion of the Translation, never a
// Translation of its own. A Module that cannot be read is skipped and reported
// in warnings; discovery never fails.
func Discover(dir string) (translations []Translation, warnings []string) {
	paths, err := filepath.Glob(filepath.Join(dir, "*"+extension))
	if err != nil {
		return nil, []string{fmt.Sprintf("cannot scan the Module directory %s: %v", dir, err)}
	}

	companions := make(map[string]string) // Bible abbreviation -> Subheadings Module path
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		name := filepath.Base(path)
		switch TypeOf(name) {
		case TypeBible:
			translation, err := readTranslation(path)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("%s: %v", name, err))
				continue
			}
			translations = append(translations, translation)
		case TypeSubheadings:
			companions[companionBase(name)] = path
		}
	}

	for i := range translations {
		translations[i].subheadings = companions[translations[i].Abbreviation]
	}

	sort.Slice(translations, func(i, j int) bool {
		return translations[i].Abbreviation < translations[j].Abbreviation
	})
	return translations, warnings
}

// companionBase is the abbreviation of the Bible a companion Module belongs to:
// its name without the extension, the type suffix, and the trailing "-s" that
// marks Subheadings, so "RV60-s.subheadings.SQLite3" names "RV60".
//
// shortcut: only the "-s" Subheadings suffix is honoured; teach this the other
// companion suffixes when the reader reads more than one kind.
func companionBase(name string) string {
	stem := strings.TrimSuffix(name, extension)
	if i := strings.LastIndex(stem, "."); i >= 0 {
		stem = stem[:i]
	}
	return strings.TrimSuffix(stem, "-s")
}

// readTranslation opens one Bible Module read-only and reads its info table.
func readTranslation(path string) (Translation, error) {
	db, err := openReadOnly(path)
	if err != nil {
		return Translation{}, err
	}
	defer func() { _ = db.Close() }()

	tables, err := moduleTables(db)
	if err != nil {
		return Translation{}, err
	}
	for _, required := range []string{tableInfo, tableVerses} {
		if !tables[required] {
			return Translation{}, fmt.Errorf("missing the required %s table", required)
		}
	}
	if !tables[tableBooks] && !tables[tableBooksAll] {
		return Translation{}, fmt.Errorf("missing a %s or %s table", tableBooks, tableBooksAll)
	}

	info, err := infoTable(db)
	if err != nil {
		return Translation{}, err
	}
	return Translation{
		Abbreviation: strings.TrimSuffix(filepath.Base(path), extension),
		Title:        info["description"],
		Language:     info["language"],
		path:         path,
	}, nil
}

// openReadOnly opens a Module without ever writing to it. The URI form keeps
// the mode=ro parameter out of the file name itself.
func openReadOnly(path string) (*sql.DB, error) {
	dsn := url.URL{Scheme: "file", Path: path, RawQuery: "mode=ro"}
	return sql.Open("sqlite", dsn.String())
}

// moduleTables reports the tables present in a Module. Querying is also what
// surfaces a corrupt file, since opening a SQLite database is lazy.
func moduleTables(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type = 'table'")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	tables := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables[name] = true
	}
	return tables, rows.Err()
}

// infoTable reads every info key/value pair into a map.
func infoTable(db *sql.DB) (map[string]string, error) {
	rows, err := db.Query("SELECT name, value FROM info")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	info := make(map[string]string)
	for rows.Next() {
		var name string
		var value sql.NullString
		if err := rows.Scan(&name, &value); err != nil {
			return nil, err
		}
		info[name] = value.String
	}
	return info, rows.Err()
}
