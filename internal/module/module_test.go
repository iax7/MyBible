package module

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iax7/mybible/internal/modtest"
)

func TestTypeOfReadsTheTypeFromTheFileName(t *testing.T) {
	cases := []struct {
		file string
		want Type
	}{
		{"RST.SQLite3", TypeBible},
		{"KJV.SQLite3", TypeBible},
		{"Strong.dictionary.SQLite3", "dictionary"},
		{"RCAS.commentaries.SQLite3", "commentaries"},
	}
	for _, tc := range cases {
		if got := TypeOf(tc.file); got != tc.want {
			t.Errorf("TypeOf(%q) = %q, want %q", tc.file, got, tc.want)
		}
	}
}

func TestDiscoverReadsTheTitleAndLanguage(t *testing.T) {
	dir := t.TempDir()
	modtest.Bible(t, dir, "RST.SQLite3", map[string]string{
		"description": "Reina-Valera 1909",
		"language":    "es",
	})

	translations, warnings := Discover(dir)

	if len(warnings) != 0 {
		t.Fatalf("Discover warned about a valid Module: %v", warnings)
	}
	if len(translations) != 1 {
		t.Fatalf("Discover found %d Translations, want 1", len(translations))
	}
	got := translations[0]
	if got.Abbreviation != "RST" {
		t.Errorf("Abbreviation = %q, want RST", got.Abbreviation)
	}
	if got.Title != "Reina-Valera 1909" {
		t.Errorf("Title = %q, want Reina-Valera 1909", got.Title)
	}
	if got.Language != "es" {
		t.Errorf("Language = %q, want es", got.Language)
	}
}

func TestDiscoverIgnoresNonBibleModules(t *testing.T) {
	dir := t.TempDir()
	modtest.Bible(t, dir, "RST.SQLite3", map[string]string{"description": "Reina-Valera"})
	modtest.Create(t, dir, "Strong.dictionary.SQLite3",
		"CREATE TABLE info (name TEXT, value TEXT)",
		"CREATE TABLE dictionary (topic TEXT, definition TEXT NOT NULL)")

	translations, warnings := Discover(dir)

	if len(warnings) != 0 {
		t.Fatalf("Discover warned about an ignored Module: %v", warnings)
	}
	if len(translations) != 1 || translations[0].Abbreviation != "RST" {
		t.Fatalf("Discover = %+v, want only the RST Bible", translations)
	}
}

func TestDiscoverWarnsAboutACorruptModuleAndKeepsTheRest(t *testing.T) {
	dir := t.TempDir()
	modtest.Bible(t, dir, "RST.SQLite3", map[string]string{"description": "Reina-Valera"})
	if err := os.WriteFile(filepath.Join(dir, "Broken.SQLite3"), []byte("not a database"), 0o644); err != nil {
		t.Fatal(err)
	}

	translations, warnings := Discover(dir)

	if len(translations) != 1 {
		t.Errorf("Discover found %d Translations, want the one valid Module", len(translations))
	}
	if len(warnings) != 1 {
		t.Fatalf("Discover warned %d times, want 1: %v", len(warnings), warnings)
	}
	if !strings.Contains(warnings[0], "Broken.SQLite3") {
		t.Errorf("warning %q does not name the corrupt Module", warnings[0])
	}
}

func TestDiscoverWarnsAboutAModuleMissingARequiredTable(t *testing.T) {
	dir := t.TempDir()
	modtest.Create(t, dir, "Odd.SQLite3",
		"CREATE TABLE info (name TEXT, value TEXT)",
		"INSERT INTO info (name, value) VALUES ('description', 'Half a Module')")

	translations, warnings := Discover(dir)

	if len(translations) != 0 {
		t.Errorf("Discover found %d Translations, want none", len(translations))
	}
	if len(warnings) != 1 {
		t.Fatalf("Discover warned %d times, want 1: %v", len(warnings), warnings)
	}
	if !strings.Contains(warnings[0], "Odd.SQLite3") {
		t.Errorf("warning %q does not name the incomplete Module", warnings[0])
	}
}

func TestDiscoverLeavesModuleFilesByteIdentical(t *testing.T) {
	dir := t.TempDir()
	path := modtest.Bible(t, dir, "RST.SQLite3", map[string]string{"description": "Reina-Valera"})

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	Discover(dir)

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Error("the Module changed after the reader opened it; modules must be opened read-only")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("the Module directory gained files after a read: %v", entries)
	}
}
