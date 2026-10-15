package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

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

func TestMissingModuleDirectoryShowsHowToConfigureIt(t *testing.T) {
	got := render(t, newReader(t, t.TempDir(), ""))
	if !strings.Contains(got, "module_directory") {
		t.Errorf("empty state does not explain how to set the Module directory\n%s", got)
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

	got := render(t, m)
	if !strings.Contains(got, "Broken.SQLite3") {
		t.Errorf("the corrupt Module is not reported as a warning\n%s", got)
	}

	m, _ = send(t, m, press('t', "t", 0))
	if got := render(t, m); !strings.Contains(got, "Reina-Valera 1909") {
		t.Errorf("the remaining Translation did not load\n%s", got)
	}
}
