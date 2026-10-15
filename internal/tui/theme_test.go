package tui

import (
	"strings"
	"testing"

	"github.com/iax7/mybible/internal/config"
	"github.com/iax7/mybible/internal/modtest"
)

// themedReader builds a reader whose Verse contains the words of Jesus, read
// with the given theme settings.
func themedReader(t *testing.T, theme string, colors config.Colors) Model {
	t.Helper()
	dir := t.TempDir()
	modtest.BibleAll(t, dir, "RST.SQLite3",
		map[string]string{"description": "Reina-Valera 1909"},
		genesis(),
		[]modtest.Verse{{Book: 10, Chapter: 1, Verse: 1, Text: "and <J>Jesus wept</J>"}})
	return configured(t, config.Config{ModuleDir: dir, ActiveTranslation: "RST", Theme: theme, Colors: colors})
}

func TestTheDarkAndLightThemesRenderTheWordsOfJesusDifferently(t *testing.T) {
	dark := themedReader(t, "dark", config.Colors{}).View().Content
	if !strings.Contains(dark, "\x1b[38;5;196m") {
		t.Errorf("the dark theme does not use its red for the words of Jesus\n%q", dark)
	}

	light := themedReader(t, "light", config.Colors{}).View().Content
	if !strings.Contains(light, "\x1b[38;5;160m") {
		t.Errorf("the light theme does not use its red for the words of Jesus\n%q", light)
	}
}

func TestAThemeColorOverrideWins(t *testing.T) {
	got := themedReader(t, "dark", config.Colors{Jesus: "9"}).View().Content
	if !strings.Contains(got, "\x1b[91m") {
		t.Errorf("the words-of-Jesus override was ignored\n%q", got)
	}
	if strings.Contains(got, "\x1b[38;5;196m") {
		t.Errorf("the built-in red survived the override\n%q", got)
	}
}

func TestTheReaderKeepsThemeAndLanguageWhenItSavesState(t *testing.T) {
	dir := t.TempDir()
	modtest.BibleAll(t, dir, "RST.SQLite3", map[string]string{"description": "Reina-Valera 1909"},
		genesis(), []modtest.Verse{{Book: 10, Chapter: 1, Verse: 1, Text: "En el principio"}})

	configDir := t.TempDir()
	cfg := config.Config{ModuleDir: dir, ActiveTranslation: "RST", Theme: "light", Language: "es"}
	if err := config.Save(configDir, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	m := newReader(t, configDir, dir)
	send(t, m, press('q', "q", 0)) // the reader persists its state on quit

	saved, err := config.Load(configDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if saved.Theme != "light" || saved.Language != "es" {
		t.Errorf("saving state lost the theme or the language: %+v", saved)
	}
}
