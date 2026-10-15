package tui

import (
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

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
	m, _ := send(t, New(), tea.WindowSizeMsg{Width: 80, Height: 24})
	return m
}

func render(t *testing.T, m Model) string {
	t.Helper()
	return ansi.ReplaceAllString(m.View().Content, "")
}

func press(code rune, text string, mod tea.KeyMod) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code, Text: text, Mod: mod}
}

func TestShellShowsEveryCommandWithItsKey(t *testing.T) {
	got := render(t, sized(t))

	for _, c := range Commands {
		if !strings.Contains(got, c.Label) {
			t.Errorf("top bar is missing command %q\n%s", c.Label, got)
		}
	}
}

func TestShellShowsReferencePlaceholder(t *testing.T) {
	if got := render(t, sized(t)); !strings.Contains(got, emptyReference) {
		t.Errorf("top bar is missing the Reference placeholder %q\n%s", emptyReference, got)
	}
}

func TestToggleBarHidesAndShowsTheTopBar(t *testing.T) {
	m := sized(t)

	m, _ = send(t, m, press('b', "", tea.ModCtrl))
	if got := render(t, m); strings.Contains(got, "Translation") {
		t.Errorf("top bar still visible after Ctrl-B\n%s", got)
	}

	m, _ = send(t, m, press('b', "", tea.ModCtrl))
	if got := render(t, m); !strings.Contains(got, "Translation") {
		t.Errorf("top bar not restored after second Ctrl-B\n%s", got)
	}
}

func TestHelpOpensAndCloses(t *testing.T) {
	m := sized(t)

	m, _ = send(t, m, press('?', "?", 0))
	if got := render(t, m); !strings.Contains(got, "Toggle top bar") {
		t.Errorf("help modal did not open\n%s", got)
	}

	m, _ = send(t, m, press(tea.KeyEscape, "", 0))
	if got := render(t, m); strings.Contains(got, "Toggle top bar") {
		t.Errorf("help modal did not close on Esc\n%s", got)
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
	if got := render(t, m); !strings.Contains(got, "Toggle top bar") {
		t.Errorf("help modal closed by an unrelated key\n%s", got)
	}
}
