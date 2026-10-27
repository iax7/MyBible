package tui

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/iax7/mybible/internal/config"
	"github.com/iax7/mybible/internal/modtest"
)

// abbreviationFixture builds two Spanish Translations that abbreviate the same
// Books differently, so a Preferred set can be seen fixing the disagreement.
func abbreviationFixture(t *testing.T, cfg config.Config) (moduleDir, configDir string) {
	t.Helper()

	books := func(genesis, exodus, john string) []modtest.Book {
		return []modtest.Book{
			{Number: 10, ShortName: genesis, LongName: "Génesis", Present: true},
			{Number: 20, ShortName: exodus, LongName: "Éxodo", Present: true},
			{Number: 500, ShortName: john, LongName: "Juan", Present: true},
		}
	}
	verses := []modtest.Verse{
		{Book: 10, Chapter: 1, Verse: 1, Text: "En el principio"},
		{Book: 20, Chapter: 1, Verse: 1, Text: "Éxodo"},
		{Book: 500, Chapter: 1, Verse: 1, Text: "Juan"},
	}

	moduleDir = t.TempDir()
	modtest.BibleAll(t, moduleDir, "NBLA.SQLite3",
		map[string]string{"description": "Nueva Biblia de las Américas", "language": "es"},
		books("Gn", "Éx", "Jn"), verses)
	modtest.BibleAll(t, moduleDir, "RVR60.SQLite3",
		map[string]string{"description": "Reina-Valera 1960", "language": "es"},
		books("Gen", "Ex", "Jn"), verses)

	configDir = t.TempDir()
	cfg.ModuleDir = moduleDir
	cfg.ActiveTranslation = "NBLA"
	if err := config.Save(configDir, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}
	return moduleDir, configDir
}

// abbreviationReader builds the fixture over a Preferred set, when one is given.
func abbreviationReader(t *testing.T, cfg config.Config, set *config.Abbreviations) (Model, string) {
	t.Helper()
	moduleDir, configDir := abbreviationFixture(t, cfg)
	if set != nil {
		if err := config.SaveAbbreviations(configDir, *set); err != nil {
			t.Fatalf("SaveAbbreviations: %v", err)
		}
	}
	return newReader(t, configDir, moduleDir), configDir
}

func TestPreferredAbbreviationsReplaceTheModulesInTheBookList(t *testing.T) {
	m, _ := abbreviationReader(t, config.Config{PreferredAbbreviations: true},
		&config.Abbreviations{ByBookNumber: map[int]string{10: "GenX"}})

	m, _ = send(t, m, press('b', "b", 0))
	got := modalText(t, m)
	if !strings.Contains(got, "GenX · Génesis") {
		t.Errorf("the Book list did not show the Preferred abbreviation\n%s", got)
	}
	if strings.Contains(got, "Gn · Génesis") {
		t.Errorf("the Book list still shows the Module's abbreviation\n%s", got)
	}
	if !strings.Contains(got, "Éx") || !strings.Contains(got, "Éxodo") {
		t.Errorf("a Book without a Preferred abbreviation lost its Module abbreviation\n%s", got)
	}
}

func TestPreferredAbbreviationsAreAcceptedAsInput(t *testing.T) {
	m, _ := abbreviationReader(t, config.Config{PreferredAbbreviations: true},
		&config.Abbreviations{ByBookNumber: map[int]string{10: "GenX"}})

	m = goTo(t, m, "GenX 1:1")
	assertReference(t, m, "Genesis 1:1")

	m = goTo(t, m, "Gn 1:1") // the Module's own abbreviation still works
	assertReference(t, m, "Genesis 1:1")
}

func TestPreferredAbbreviationsAreScopedByLanguage(t *testing.T) {
	m, _ := abbreviationReader(t, config.Config{PreferredAbbreviations: true},
		&config.Abbreviations{Language: "en", ByBookNumber: map[int]string{10: "GenX"}})
	m, _ = send(t, m, press('b', "b", 0))
	if got := modalText(t, m); strings.Contains(got, "GenX") {
		t.Errorf("a set for another language was applied\n%s", got)
	}

	m, _ = abbreviationReader(t, config.Config{PreferredAbbreviations: true},
		&config.Abbreviations{Language: "es", ByBookNumber: map[int]string{10: "GenX"}})
	m, _ = send(t, m, press('b', "b", 0))
	if got := modalText(t, m); !strings.Contains(got, "GenX") {
		t.Errorf("a set for the Active translation's language was ignored\n%s", got)
	}
}

func TestPreferredAbbreviationsAreIgnoredWhenTheOptionIsOff(t *testing.T) {
	m, _ := abbreviationReader(t, config.Config{},
		&config.Abbreviations{ByBookNumber: map[int]string{10: "GenX"}})
	m, _ = send(t, m, press('b', "b", 0))
	if got := modalText(t, m); strings.Contains(got, "GenX") {
		t.Errorf("the reader used the set with the option off\n%s", got)
	}
}

func TestSeedingCopiesTheChosenTranslationsAbbreviations(t *testing.T) {
	m, configDir := abbreviationReader(t, config.Config{PreferredAbbreviations: true}, nil)

	if m.picker.kind != pickerPreferredAbbreviations {
		t.Fatalf("the reader did not ask for an abbreviation source, picker = %v", m.picker.kind)
	}
	if got := modalText(t, m); !strings.Contains(got, "Select Abbreviation Source") {
		t.Errorf("the abbreviation source modal is missing its title\n%s", got)
	}

	m, _ = send(t, m, press(tea.KeyDown, "", 0)) // RVR60, after NBLA
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))
	if m.picker.isOpen() {
		t.Error("the abbreviation source modal stayed open after a choice")
	}

	got, err := config.LoadAbbreviations(configDir)
	if err != nil {
		t.Fatalf("LoadAbbreviations: %v", err)
	}
	want := map[int]string{10: "Gen", 20: "Ex", 500: "Jn"}
	if !reflect.DeepEqual(got.ByBookNumber, want) {
		t.Errorf("seeded abbreviations = %v, want %v", got.ByBookNumber, want)
	}
	if got.Language != "" {
		t.Errorf("the seeded set names the language %q, want a set for every Translation", got.Language)
	}

	m, _ = send(t, m, press('b', "b", 0))
	if text := modalText(t, m); !strings.Contains(text, "Gen · Génesis") {
		t.Errorf("the seeded abbreviations did not reach the Active translation\n%s", text)
	}
}

func TestCancellingTheSeedWritesNoFile(t *testing.T) {
	m, configDir := abbreviationReader(t, config.Config{PreferredAbbreviations: true}, nil)

	m, _ = send(t, m, press(tea.KeyEscape, "", 0))
	if m.picker.isOpen() {
		t.Error("the abbreviation source modal stayed open after Esc")
	}
	path := filepath.Join(configDir, "preferred.abbreviations.json")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("cancelling the seed wrote %s", path)
	}
}

func TestAMalformedPreferredSetIsReportedAndLeftAlone(t *testing.T) {
	moduleDir, configDir := abbreviationFixture(t, config.Config{PreferredAbbreviations: true})
	path := filepath.Join(configDir, "preferred.abbreviations.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := newReader(t, configDir, moduleDir)
	if m.picker.isOpen() {
		t.Error("the reader asked to seed over a malformed file")
	}
	if got := strings.Join(m.warnings, "\n"); !strings.Contains(got, "preferred abbreviations") {
		t.Errorf("the malformed file was not reported\n%s", got)
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != "{not json" {
		t.Errorf("the malformed file changed: %q, %v", data, err)
	}
}

func TestTheSeedPromptIsSkippedWithoutTranslations(t *testing.T) {
	configDir := t.TempDir()
	empty := t.TempDir()
	if err := config.Save(configDir, config.Config{ModuleDir: empty, PreferredAbbreviations: true}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	m := newReader(t, configDir, empty)
	if m.picker.isOpen() {
		t.Error("the reader asked to seed with no Translation to copy from")
	}
	if got := strings.Join(m.warnings, "\n"); !strings.Contains(got, "no Translation to copy from") {
		t.Errorf("the missing source was not reported\n%s", got)
	}
}

func TestTheAbbreviationSourceModalReadsInBothLanguages(t *testing.T) {
	for language, want := range map[string]string{
		"en": "Select Abbreviation Source",
		"es": "Seleccionar origen de abreviaturas",
	} {
		m, _ := abbreviationReader(t, config.Config{PreferredAbbreviations: true, Language: language}, nil)
		if got := modalText(t, m); !strings.Contains(got, want) {
			t.Errorf("the %s modal is missing %q\n%s", language, want, got)
		}
	}
}
