package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/iax7/mybible/internal/config"
	"github.com/iax7/mybible/internal/modtest"
)

// localized builds a reader whose Active translation is loaded and whose config
// names the interface language, so one state can be read in either language.
func localized(t *testing.T, language string) Model {
	t.Helper()
	dir := t.TempDir()
	modtest.BibleAll(t, dir, "RST.SQLite3",
		map[string]string{"description": "Reina-Valera 1909", "language": "es"},
		genesis(),
		[]modtest.Verse{{Book: 10, Chapter: 1, Verse: 1, Text: "En el principio <J>creó Dios</J>"}})
	return configured(t, config.Config{ModuleDir: dir, ActiveTranslation: "RST", Language: language})
}

// configured builds a reader over the given config. A config that names no
// Module directory leaves the reader on its default one, which is empty unless
// a test fills it.
func configured(t *testing.T, cfg config.Config) Model {
	t.Helper()
	configDir := t.TempDir()
	if err := config.Save(configDir, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}
	return newReader(t, configDir, cfg.ModuleDir)
}

// withoutDataDir removes every way the reader could reach a default Module
// directory, the only route left into the "not configured" state.
func withoutDataDir(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("HOME", "")
}

// assertSameState reaches one state in both languages and checks that the
// English reader shows english and the Spanish reader shows spanish.
func assertSameState(t *testing.T, keys []tea.KeyPressMsg, english, spanish string) {
	t.Helper()
	en, es := localized(t, "en"), localized(t, "es")
	for _, key := range keys {
		en, _ = send(t, en, key)
		es, _ = send(t, es, key)
	}
	if got := render(t, en); !strings.Contains(got, english) {
		t.Errorf("the English view is missing %q\n%s", english, got)
	}
	if got := render(t, es); !strings.Contains(got, spanish) {
		t.Errorf("the Spanish view is missing %q\n%s", spanish, got)
	}
}

func TestTheTopBarReadsInBothLanguages(t *testing.T) {
	for _, line := range []struct{ english, spanish string }{
		{"Search", "Buscar"},
		{"History", "Historial"},
		{"Help", "Ayuda"},
	} {
		assertSameState(t, nil, line.english, line.spanish)
	}
}

func TestEveryModalTitleReadsInBothLanguages(t *testing.T) {
	for _, modal := range []struct {
		key              tea.KeyPressMsg
		english, spanish string
	}{
		{press('b', "b", 0), selectBook, "Seleccionar libro"},
		{press('c', "c", 0), selectChapter, "Seleccionar capítulo"},
		{press('t', "t", 0), selectTranslation, "Seleccionar traducción"},
		{press('v', "v", 0), selectReference, "Ir a la referencia"},
		{press('/', "/", 0), searchTitle, "Buscar"},
		{press('H', "H", 0), historyTitle, "Historial"},
	} {
		assertSameState(t, []tea.KeyPressMsg{modal.key}, modal.english, modal.spanish)
	}
}

func TestTheHelpModalReadsInBothLanguages(t *testing.T) {
	help := []tea.KeyPressMsg{press('?', "?", 0)}
	for _, line := range []struct{ english, spanish string }{
		{"Keys", "Teclas"},
		{"Book", "Libro"},
		{"Chapter", "Capítulo"},
		{"Go To", "Ir a"},
		{"Translation", "Traducción"},
		{"Toggle top bar", "Mostrar u ocultar la barra"},
		{"Back", "Atrás"},
		{"Forward", "Adelante"},
		{"Quit", "Salir"},
		{"Close", "Cerrar"},
	} {
		assertSameState(t, help, line.english, line.spanish)
	}
}

func TestTheModalEmptyMessagesReadInBothLanguages(t *testing.T) {
	assertSameState(t, []tea.KeyPressMsg{press('/', "/", 0)}, "Type to search", "Escribe para buscar")
	assertSameState(t, []tea.KeyPressMsg{press('H', "H", 0)}, "Nothing visited yet", "Aún no has visitado nada")
	assertSameState(t, []tea.KeyPressMsg{press('v', "v", 0)}, "enter to go", "enter para ir")

	noMatch := []tea.KeyPressMsg{press('/', "/", 0), press('z', "z", 0)}
	assertSameState(t, noMatch, "No matches", "Sin resultados")
}

func TestTheEmptyStateReadsInBothLanguages(t *testing.T) {
	withoutDataDir(t)

	en := render(t, configured(t, config.Config{Language: "en"}))
	if !strings.Contains(en, "No Module directory configured.") {
		t.Errorf("the English empty state is missing its message\n%s", en)
	}
	es := render(t, configured(t, config.Config{Language: "es"}))
	if !strings.Contains(es, "No hay un directorio de módulos configurado.") {
		t.Errorf("the Spanish empty state is missing its message\n%s", es)
	}
}

func TestAnEmptyModuleDirectoryReadsInBothLanguages(t *testing.T) {
	en := render(t, configured(t, config.Config{ModuleDir: t.TempDir(), Language: "en"}))
	if !strings.Contains(en, "No Bible Modules found in") {
		t.Errorf("the English message for an empty Module directory is missing\n%s", en)
	}
	es := render(t, configured(t, config.Config{ModuleDir: t.TempDir(), Language: "es"}))
	if !strings.Contains(es, "No se encontraron módulos de la Biblia en") {
		t.Errorf("the Spanish message for an empty Module directory is missing\n%s", es)
	}
}

func TestLanguageInConfigBeatsTheSystemLocale(t *testing.T) {
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "es_ES.UTF-8")

	if got := render(t, localized(t, "en")); !strings.Contains(got, "Search") {
		t.Errorf("the config language did not win over the locale\n%s", got)
	}
}

func TestLanguageFallsBackToTheSystemLocale(t *testing.T) {
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "es_ES.UTF-8")

	got := render(t, configured(t, config.Config{}))
	if !strings.Contains(got, "Buscar") {
		t.Errorf("the system locale did not select Spanish\n%s", got)
	}
}

func TestLanguageFallsBackToEnglish(t *testing.T) {
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "fr_FR.UTF-8")

	got := render(t, configured(t, config.Config{}))
	if !strings.Contains(got, "Search") {
		t.Errorf("an unknown locale did not fall back to English\n%s", got)
	}
}
