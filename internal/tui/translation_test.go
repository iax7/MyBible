package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/iax7/mybible/internal/config"
	"github.com/iax7/mybible/internal/modtest"
)

// newReader builds a Model over temporary config and Module directories, the
// way main does for the XDG directories, and gives it a window size.
func newReader(t *testing.T, configDir, moduleDir string) Model {
	t.Helper()
	m := New(Options{ConfigDir: configDir, ModuleDir: moduleDir})
	m, _ = send(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	return m
}

func typeKey(t *testing.T, m Model, text string) Model {
	t.Helper()
	for _, r := range text {
		m, _ = send(t, m, press(r, string(r), 0))
	}
	return m
}

func TestTheDefaultModuleDirectoryIsUsedWhenConfigNamesNone(t *testing.T) {
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)
	dir := filepath.Join(dataHome, "mybible")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	modtest.Bible(t, dir, "RST.SQLite3", map[string]string{"description": "Reina-Valera 1909"})

	m := newReader(t, t.TempDir(), "")
	m, _ = send(t, m, press('t', "t", 0))

	if got := render(t, m); !strings.Contains(got, "Reina-Valera 1909") {
		t.Errorf("the default Module directory was not scanned\n%s", got)
	}
}

func TestAnEmptyModuleDirectoryExplainsItself(t *testing.T) {
	got := render(t, newReader(t, t.TempDir(), t.TempDir()))
	if !strings.Contains(got, "No Bible Modules") {
		t.Errorf("empty Module directory has no message\n%s", got)
	}
}

func TestTheModuleDirectoryIsReadFromConfig(t *testing.T) {
	configDir := t.TempDir()
	dir := t.TempDir()
	modtest.Bible(t, dir, "RST.SQLite3", map[string]string{"description": "Reina-Valera 1909"})
	if err := config.Save(configDir, config.Config{ModuleDir: dir}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	m := newReader(t, configDir, "")
	m, _ = send(t, m, press('t', "t", 0))

	if got := render(t, m); !strings.Contains(got, "Reina-Valera 1909") {
		t.Errorf("the configured Module directory was not scanned\n%s", got)
	}
}

func TestDiscoveredTranslationsAreOfferedInThePicker(t *testing.T) {
	dir := t.TempDir()
	modtest.Bible(t, dir, "RST.SQLite3", map[string]string{
		"description": "Reina-Valera 1909",
		"language":    "es",
	})
	modtest.Bible(t, dir, "KJV.SQLite3", map[string]string{
		"description": "King James Version",
		"language":    "en",
	})

	m := newReader(t, t.TempDir(), dir)
	m, _ = send(t, m, press('t', "t", 0))

	got := render(t, m)
	for _, title := range []string{"Reina-Valera 1909", "King James Version"} {
		if !strings.Contains(got, title) {
			t.Errorf("picker is missing the Translation %q\n%s", title, got)
		}
	}
}

func TestTheTranslationPickerIsFilterable(t *testing.T) {
	dir := t.TempDir()
	modtest.Bible(t, dir, "RST.SQLite3", map[string]string{"description": "Reina-Valera 1909"})
	modtest.Bible(t, dir, "KJV.SQLite3", map[string]string{"description": "King James Version"})

	m := newReader(t, t.TempDir(), dir)
	m, _ = send(t, m, press('t', "t", 0))
	m = typeKey(t, m, "king")

	got := render(t, m)
	if !strings.Contains(got, "King James Version") {
		t.Errorf("filter dropped the matching Translation\n%s", got)
	}
	if strings.Contains(got, "Reina-Valera") {
		t.Errorf("filter kept a non-matching Translation\n%s", got)
	}
}

// Filtering a list modal must not resize it: the box keeps the size it had when
// the whole list was shown, so it does not flicker under the reader's eyes.
func TestTheTranslationPickerOnlyGrows(t *testing.T) {
	dir := t.TempDir()
	modtest.Bible(t, dir, "RST.SQLite3", map[string]string{"description": "Reina-Valera 1909 Reina-Valera Actualizada"})
	modtest.Bible(t, dir, "KJV.SQLite3", map[string]string{"description": "King James Version"})

	m := newReader(t, t.TempDir(), dir)
	m, _ = send(t, m, press('t', "t", 0))
	width, height := lipgloss.Width(m.modalView()), lipgloss.Height(m.modalView())

	// "king" drops the wider row; the box must hold its size.
	m = typeKey(t, m, "king")
	if got := lipgloss.Width(m.modalView()); got != width {
		t.Errorf("the picker changed width to %d while typing, from %d", got, width)
	}
	if got := lipgloss.Height(m.modalView()); got != height {
		t.Errorf("the picker changed height to %d while typing, from %d", got, height)
	}
}

func TestTheTranslationPickerClosesOnEscape(t *testing.T) {
	dir := t.TempDir()
	modtest.Bible(t, dir, "RST.SQLite3", map[string]string{"description": "Reina-Valera 1909"})

	m := newReader(t, t.TempDir(), dir)
	m, _ = send(t, m, press('t', "t", 0))
	m, _ = send(t, m, press(tea.KeyEscape, "", 0))

	if got := render(t, m); strings.Contains(got, selectTranslation) {
		t.Errorf("picker stayed open after Esc\n%s", got)
	}
}

func TestSelectingATranslationMakesItActiveAndRemembersIt(t *testing.T) {
	configDir := t.TempDir()
	dir := t.TempDir()
	modtest.Bible(t, dir, "RST.SQLite3", map[string]string{"description": "Reina-Valera 1909"})
	modtest.Bible(t, dir, "KJV.SQLite3", map[string]string{"description": "King James Version"})

	m := newReader(t, configDir, dir)
	m, _ = send(t, m, press('t', "t", 0))
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))

	got := render(t, m)
	if strings.Contains(got, selectTranslation) {
		t.Errorf("picker stayed open after Enter\n%s", got)
	}
	if !strings.Contains(got, "King James Version") {
		t.Errorf("the picked Translation is not the Active one\n%s", got)
	}

	saved, err := config.Load(configDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if saved.ActiveTranslation != "KJV" {
		t.Errorf("saved Active translation = %q, want KJV", saved.ActiveTranslation)
	}

	if got := render(t, newReader(t, configDir, dir)); !strings.Contains(got, "King James Version") {
		t.Errorf("the Active translation was not restored on the next launch\n%s", got)
	}
}

func TestACorruptModuleIsSkippedWithAWarning(t *testing.T) {
	dir := t.TempDir()
	modtest.Bible(t, dir, "RST.SQLite3", map[string]string{"description": "Reina-Valera 1909"})
	if err := os.WriteFile(filepath.Join(dir, "Broken.SQLite3"), []byte("not a database"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := newReader(t, t.TempDir(), dir)

	// Warnings live in the Log, reached from the bar; the reading pane is clean.
	if got := render(t, m); strings.Contains(got, "Broken.SQLite3") {
		t.Errorf("the reading pane shows a warning\n%s", got)
	}
	m, _ = send(t, m, press('L', "L", 0))
	if got := render(t, m); !strings.Contains(got, "Broken.SQLite3") {
		t.Errorf("the corrupt Module is not reported in the Log\n%s", got)
	}
	m, _ = send(t, m, press(tea.KeyEscape, "", 0))

	m, _ = send(t, m, press('t', "t", 0))
	if got := render(t, m); !strings.Contains(got, "Reina-Valera 1909") {
		t.Errorf("the remaining Translation did not load\n%s", got)
	}
}

// switchReader is a reader on KJV holding two Translations that share Genesis
// but differ elsewhere: RST has no Exodus, so a switch can land or fall back.
func switchReader(t *testing.T) Model {
	t.Helper()
	dir := t.TempDir()
	genesis := []modtest.Verse{
		{Book: 10, Chapter: 1, Verse: 1, Text: "In the beginning"},
		{Book: 10, Chapter: 1, Verse: 2, Text: "And the earth was"},
		{Book: 10, Chapter: 1, Verse: 3, Text: "And God said"},
	}
	gen := []modtest.Book{{Number: 10, ShortName: "Gen", LongName: "Genesis"}}
	modtest.BibleBooks(t, dir, "RST.SQLite3", map[string]string{"description": "Reina-Valera 1909"}, gen, genesis)
	modtest.BibleBooks(t, dir, "KJV.SQLite3", map[string]string{"description": "King James Version"},
		append(gen, modtest.Book{Number: 20, ShortName: "Exo", LongName: "Exodus"}),
		append(genesis, modtest.Verse{Book: 20, Chapter: 1, Verse: 1, Text: "Now these are the names"}))

	configDir := t.TempDir()
	if err := config.Save(configDir, config.Config{ModuleDir: dir, ActiveTranslation: "KJV"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	return newReader(t, configDir, dir)
}

// switchTo picks a Translation through the picker's filter.
func switchTo(t *testing.T, m Model, filter string) Model {
	t.Helper()
	m, _ = send(t, m, press('t', "t", 0))
	m = typeKey(t, m, filter)
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))
	return m
}

func TestSwitchingTranslationKeepsTheReadingPosition(t *testing.T) {
	m := switchReader(t)
	m = goTo(t, m, "Genesis 1:3")

	m = switchTo(t, m, "reina")

	assertReference(t, m, "Genesis 1:3")
	if got := render(t, m); !strings.Contains(got, "And God said") {
		t.Errorf("the new Translation did not open at Genesis 1:3\n%s", got)
	}
}

func TestSwitchingTranslationFallsBackWhenThePositionIsMissing(t *testing.T) {
	m := switchReader(t)
	m = goTo(t, m, "Exodus 1:1")

	m = switchTo(t, m, "reina")

	assertReference(t, m, "Genesis 1:1") // RST has no Exodus
}

func TestSwitchingTranslationIsNotAVisit(t *testing.T) {
	m := switchReader(t)
	m = goTo(t, m, "Genesis 1:3")

	m = switchTo(t, m, "reina")
	m = back(t, m)

	assertReference(t, m, "Genesis 1:1") // the goTo's origin: the switch added no Visit
}

func TestABrokenStoriesTableIsReportedInTheLog(t *testing.T) {
	dir := t.TempDir()
	modtest.Create(t, dir, "RST.SQLite3",
		`CREATE TABLE info (name TEXT, value TEXT)`,
		`INSERT INTO info (name, value) VALUES ('description', 'Reina-Valera 1909')`,
		`CREATE TABLE books (book_number NUMERIC, book_color TEXT, short_name TEXT, long_name TEXT, sorting_order NUMERIC)`,
		`INSERT INTO books (book_number, short_name, long_name) VALUES (10, 'Gen', 'Genesis')`,
		`CREATE TABLE verses (book_number NUMERIC, chapter NUMERIC, verse NUMERIC, text TEXT)`,
		`INSERT INTO verses (book_number, chapter, verse, text) VALUES (10, 1, 1, 'In the beginning')`,
		`CREATE TABLE stories (book_number NUMERIC)`,
	)

	configDir := t.TempDir()
	if err := config.Save(configDir, config.Config{ModuleDir: dir, ActiveTranslation: "RST"}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	m := newReader(t, configDir, dir)
	if got := render(t, m); strings.Contains(got, "stories") {
		t.Errorf("the reading pane shows a warning\n%s", got)
	}
	m, _ = send(t, m, press('L', "L", 0))
	if got := render(t, m); !strings.Contains(got, "cannot read the stories table") {
		t.Errorf("the broken stories table is not reported in the Log\n%s", got)
	}
	m, _ = send(t, m, press(tea.KeyEscape, "", 0))

	if got := render(t, m); !strings.Contains(got, "1 In the beginning") {
		t.Errorf("the Bible did not stay readable\n%s", got)
	}
}
