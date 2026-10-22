package tui

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/iax7/mybible/internal/modtest"
)

// TestMain keeps the reader's XDG directories inside a temporary folder, so no
// test reads the Modules or the config the developer happens to have.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "mybible-test")
	if err != nil {
		log.Fatal(err)
	}
	for _, name := range []string{"XDG_CONFIG_HOME", "XDG_DATA_HOME"} {
		if err := os.Setenv(name, dir); err != nil {
			log.Fatal(err)
		}
	}

	code := m.Run()

	if err := os.RemoveAll(dir); err != nil {
		log.Print(err)
	}
	os.Exit(code)
}

var ansi = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func send(t *testing.T, m Model, msg tea.Msg) (Model, tea.Cmd) {
	t.Helper()
	updated, cmd := m.Update(msg)
	next, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want tui.Model", updated)
	}
	return next, cmd
}

func sized(t *testing.T) Model {
	t.Helper()
	m, _ := send(t, New(Options{ConfigDir: t.TempDir()}), tea.WindowSizeMsg{Width: 80, Height: 24})
	return m
}

// on is the Reference the reader is on, as "Book Chapter:Verse", or "" when it
// has opened no Chapter.
func on(m Model) string {
	reference, ok := m.currentReference()
	if !ok {
		return ""
	}
	return reference.String()
}

func render(t *testing.T, m Model) string {
	t.Helper()
	return ansi.ReplaceAllString(m.View().Content, "")
}

// modalText is the open modal's box with its styles stripped, so a test can read
// the modal without the reading pane visible behind it. The Reference modal's
// completion menu is part of what it shows, so it is included.
func modalText(t *testing.T, m Model) string {
	t.Helper()
	text := m.modalView()
	if popup := m.pickerPopupView(); popup != "" {
		text += "\n" + popup
	}
	return ansi.ReplaceAllString(text, "")
}

func press(code rune, text string, mod tea.KeyMod) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code, Text: text, Mod: mod}
}

func TestShellShowsTheActiveStateAndTheBarCommands(t *testing.T) {
	got := render(t, sized(t))

	for _, c := range barCommands {
		if !strings.Contains(got, c.Key+" "+c.Label) {
			t.Errorf("top bar is missing command %q\n%s", c.Label, got)
		}
	}
	if !strings.Contains(got, keyTranslation+" "+emptyReference) {
		t.Errorf("top bar is missing the Translation placeholder\n%s", got)
	}
	if !strings.Contains(got, keyReference+" "+emptyReference) {
		t.Errorf("top bar is missing the Reference placeholder\n%s", got)
	}
}

func TestTheTopBarKeepsTheOtherCommandsOut(t *testing.T) {
	got := render(t, sized(t))

	// Everything a Command adds to the bar beyond barCommands.
	for _, label := range []string{"Book", "Chapter", "Translation", "Go To", "Quit", "Toggle top bar", "Back", "Forward"} {
		if strings.Contains(got, label) {
			t.Errorf("top bar still announces %q\n%s", label, got)
		}
	}
}

func TestToggleBarHidesAndShowsTheTopBar(t *testing.T) {
	m := sized(t)
	bar := barCommands[0].Key + " " + barCommands[0].Label

	m, _ = send(t, m, press('b', "", tea.ModCtrl))
	if got := render(t, m); strings.Contains(got, bar) {
		t.Errorf("top bar still visible after Ctrl-B\n%s", got)
	}

	m, _ = send(t, m, press('b', "", tea.ModCtrl))
	if got := render(t, m); !strings.Contains(got, bar) {
		t.Errorf("top bar not restored after second Ctrl-B\n%s", got)
	}
}

func TestHelpOpensAndCloses(t *testing.T) {
	m := sized(t)

	m, _ = send(t, m, press('?', "?", 0))
	if got := render(t, m); !strings.Contains(got, "Movement") {
		t.Errorf("help modal did not open\n%s", got)
	}

	m, _ = send(t, m, press(tea.KeyEscape, "", 0))
	if got := render(t, m); strings.Contains(got, "Movement") {
		t.Errorf("help modal did not close on Esc\n%s", got)
	}
}

func TestTheBarShowsTheLogOnlyWhenThereAreWarnings(t *testing.T) {
	if got := render(t, sized(t)); strings.Contains(got, "⚠") {
		t.Errorf("a clean reader shows the Log entry\n%s", got)
	}

	dir := t.TempDir()
	modtest.Bible(t, dir, "RST.SQLite3", map[string]string{"description": "Reina-Valera 1909"})
	if err := os.WriteFile(filepath.Join(dir, "Broken.SQLite3"), []byte("not a database"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := newReader(t, t.TempDir(), dir)
	bar := render(t, m)
	if !strings.Contains(bar, "⚠") || !strings.Contains(bar, "Log") {
		t.Fatalf("the Log entry is missing with warnings present\n%s", bar)
	}

	m, _ = send(t, m, press('L', "L", 0))
	got := render(t, m)
	if !strings.Contains(got, "• ") {
		t.Errorf("the Log does not list its warnings as bullets\n%s", got)
	}
	if strings.Contains(got, "❯") {
		t.Errorf("the Log shows a filter line\n%s", got)
	}
}

// A modal overlays the reading pane instead of replacing it: the Chapter stays
// visible around the box.
func TestAnOpenModalKeepsTheReadingPaneVisible(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('v', "v", 0))

	got := render(t, m)
	if !strings.Contains(got, "Go to Reference") {
		t.Fatalf("the modal did not open\n%s", got)
	}
	if !strings.Contains(got, "verse01") {
		t.Errorf("the modal erased the reading pane behind it\n%s", got)
	}
}

func TestHelpListsEveryMovementAndCommandKey(t *testing.T) {
	m, _ := send(t, New(Options{ConfigDir: t.TempDir()}), tea.WindowSizeMsg{Width: 120, Height: 50})
	m, _ = send(t, m, press('?', "?", 0))

	got := render(t, m)
	for _, c := range append(append([]Command{}, MovementCommands...), Commands...) {
		if !strings.Contains(got, pad(c.Key, 10)+c.Label) {
			t.Errorf("help is missing the keybinding %s %s\n%s", c.Key, c.Label, got)
		}
	}

	key := m.styles.subheading.Render(pad(MovementCommands[0].Key, 10))
	if raw := m.modalView(); !strings.Contains(raw, key) {
		t.Errorf("the help keys are not drawn in cyan\n%s", ansi.ReplaceAllString(raw, ""))
	}
}

func TestQuitLeavesTheProgram(t *testing.T) {
	_, cmd := send(t, sized(t), press('q', "q", 0))
	if cmd == nil {
		t.Fatal("q returned no command")
	}
	if msg := cmd(); msg != tea.Quit() {
		t.Errorf("q returned %v, want tea.Quit", msg)
	}
}

func TestKeysAreIgnoredWhileHelpIsOpen(t *testing.T) {
	m := sized(t)

	m, _ = send(t, m, press('?', "?", 0))
	m, cmd := send(t, m, press('q', "q", 0))

	if cmd != nil {
		t.Error("q acted while the help modal was open")
	}
	if got := render(t, m); !strings.Contains(got, "Movement") {
		t.Errorf("help modal closed by an unrelated key\n%s", got)
	}
}
