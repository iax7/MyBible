package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/iax7/mybible/internal/config"
	"github.com/iax7/mybible/internal/modtest"
)

// back and forward are the history keys under test, driven as real key events.
func back(t *testing.T, m Model) Model {
	t.Helper()
	m, _ = send(t, m, press(tea.KeyLeft, "", tea.ModAlt))
	return m
}

func forward(t *testing.T, m Model) Model {
	t.Helper()
	m, _ = send(t, m, press(tea.KeyRight, "", tea.ModAlt))
	return m
}

// goTo jumps through the Go to Reference modal, a deliberate Visit.
func goTo(t *testing.T, m Model, reference string) Model {
	t.Helper()
	m, _ = send(t, m, press('v', "v", 0))
	m = typeKey(t, m, reference)
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))
	return m
}

// assertReference checks the Reference the reader is on. The top bar names the
// Book and the Chapter, so the Verse comes from the reader's own state.
func assertReference(t *testing.T, m Model, want string) {
	t.Helper()
	if got := on(m); got != want {
		t.Errorf("reader is on %q, want %q", got, want)
	}
}

func TestBackAndForwardRetraceJumps(t *testing.T) {
	m := navReader(t)
	m = goTo(t, m, "Genesis 2:1")
	m = goTo(t, m, "Genesis 3:1")
	assertReference(t, m, "Genesis 3:1")

	m = back(t, m)
	assertReference(t, m, "Genesis 2:1")
	m = back(t, m)
	assertReference(t, m, "Genesis 1:1")

	m = back(t, m) // nothing further back
	assertReference(t, m, "Genesis 1:1")

	m = forward(t, m)
	assertReference(t, m, "Genesis 2:1")
	m = forward(t, m)
	assertReference(t, m, "Genesis 3:1")

	m = forward(t, m) // nothing further forward
	assertReference(t, m, "Genesis 3:1")
}

func TestMovingTheVerseCursorAloneIsNotAJump(t *testing.T) {
	m := navReader(t)
	m = goTo(t, m, "Genesis 2:1")
	m, _ = send(t, m, press('j', "j", 0)) // Genesis 2:2: a bare cursor move
	assertReference(t, m, "Genesis 2:2")

	m = back(t, m)
	assertReference(t, m, "Genesis 1:1") // the jump's origin, not the cursor position

	m = forward(t, m)
	assertReference(t, m, "Genesis 2:1") // the visited Reference, not the cursor position
}

// Stepping Chapters is reading, not a Visit: it stays out of History and out of
// the Trail that back and forward walk.
func TestSteppingChaptersIsNotHistory(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press(']', "]", 0))
	assertReference(t, m, "Genesis 2:1")

	m = back(t, m)
	assertReference(t, m, "Genesis 2:1") // no Visit to retrace

	m, _ = send(t, m, press('H', "H", 0))
	if got := modalText(t, m); !strings.Contains(got, "Nothing visited yet") {
		t.Errorf("a Chapter step reached History\n%s", got)
	}
}

func TestHistoryModalListsJumpsMostRecentFirst(t *testing.T) {
	m := navReader(t)
	m = goTo(t, m, "Genesis 2:1")
	m, _ = send(t, m, press('c', "c", 0))
	m = typeKey(t, m, "3")
	m, _ = send(t, m, press(tea.KeyEnter, "", 0)) // Genesis 3:1      Select Chapter
	m, _ = send(t, m, press('b', "b", 0))
	m = typeKey(t, m, "exo")
	m, _ = send(t, m, press(tea.KeyEnter, "", 0)) // Exodus 1:1       Select Book
	m = goTo(t, m, "Genesis 1:5")                 // Genesis 1:5      Go to Reference
	m, _ = send(t, m, press('/', "/", 0))
	m = typeKey(t, m, "exo02")
	m, _ = send(t, m, press(tea.KeyEnter, "", 0)) // Exodus 2:1       Search result

	m, _ = send(t, m, press('H', "H", 0))
	if got := modalText(t, m); !strings.Contains(got, historyTitle) {
		t.Fatalf("H did not open the History modal\n%s", got)
	}
	got := render(t, m)

	order := []string{"Exodus 2:1", "Genesis 1:5", "Exodus 1:1", "Genesis 3:1", "Genesis 2:1"}
	for i := 0; i+1 < len(order); i++ {
		if !strings.Contains(got, order[i]) || !strings.Contains(got, order[i+1]) {
			t.Fatalf("History is missing %q or %q\n%s", order[i], order[i+1], got)
		}
		if strings.Index(got, order[i]) > strings.Index(got, order[i+1]) {
			t.Fatalf("History is not most-recent-first at %q\n%s", order[i], got)
		}
	}

	// Choosing a row jumps to that Reference and closes the modal.
	m, _ = send(t, m, press(tea.KeyDown, "", 0))
	m, _ = send(t, m, press(tea.KeyDown, "", 0))
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))

	assertReference(t, m, "Exodus 1:1")
	if got := modalText(t, m); strings.Contains(got, historyTitle) {
		t.Errorf("the History modal stayed open after Enter\n%s", got)
	}
}

func TestHistoryModalIsFilterable(t *testing.T) {
	m := navReader(t)
	m = goTo(t, m, "Genesis 2:1")
	m, _ = send(t, m, press('b', "b", 0))
	m = typeKey(t, m, "exo")
	m, _ = send(t, m, press(tea.KeyEnter, "", 0)) // Exodus 1:1

	m, _ = send(t, m, press('H', "H", 0))
	m = typeKey(t, m, "gen")

	got := modalText(t, m)
	if !strings.Contains(got, "Genesis 2:1") {
		t.Errorf("the History filter dropped a matching Reference\n%s", got)
	}
	if strings.Contains(got, "Exodus") {
		t.Errorf("the History filter kept a non-matching Reference\n%s", got)
	}
}

func TestHistoryKeepsOneVisitPerReferencePerDay(t *testing.T) {
	m := navReader(t)
	m = goTo(t, m, "Genesis 2:1")
	m = goTo(t, m, "Genesis 2:1")

	m, _ = send(t, m, press('H', "H", 0))
	if got := strings.Count(modalText(t, m), "Genesis 2:1"); got != 1 {
		t.Errorf("History kept %d Visits to Genesis 2:1 in one day, want 1", got)
	}
}

func TestHistoryIsGroupedByDayWithoutTimes(t *testing.T) {
	m := navReader(t)
	first := time.Date(2026, 10, 19, 10, 0, 0, 0, time.Local)
	m.now = func() time.Time { return first }
	m = goTo(t, m, "Genesis 2:1")
	m.now = func() time.Time { return first.AddDate(0, 0, 1) }
	m = goTo(t, m, "Exodus 1:1")

	m, _ = send(t, m, press('H', "H", 0))
	got := modalText(t, m)
	for _, want := range []string{"Today", "Yesterday", "Exodus 1:1", "Genesis 2:1"} {
		if !strings.Contains(got, want) {
			t.Errorf("History is missing %q\n%s", want, got)
		}
	}
	if strings.Index(got, "Today") > strings.Index(got, "Yesterday") {
		t.Errorf("History days are not most-recent-first\n%s", got)
	}
	if strings.Contains(got, "10:00") {
		t.Errorf("History shows a time\n%s", got)
	}
}

func TestHistorySurvivesARestart(t *testing.T) {
	info, books, verses := navFixture()
	dir := t.TempDir()
	modtest.BibleAll(t, dir, "RST.SQLite3", info, books, verses)

	configDir := t.TempDir()
	if err := config.Save(configDir, config.Config{ModuleDir: dir, ActiveTranslation: "RST"}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	m := newReader(t, configDir, dir)
	goTo(t, m, "Exodus 2:1")

	reopened := newReader(t, configDir, dir)
	reopened, _ = send(t, reopened, press('H', "H", 0))
	if got := modalText(t, reopened); !strings.Contains(got, "Exodus 2:1") {
		t.Errorf("History did not survive a restart\n%s", got)
	}
}

// A Reference the Active translation lacks stays in History and says why it
// cannot be reached, instead of closing the modal in silence.
func TestHistoryKeepsAMissingReferenceVisible(t *testing.T) {
	info, books, verses := navFixture() // Genesis and Exodus only
	dir := t.TempDir()
	modtest.BibleAll(t, dir, "RST.SQLite3", info, books, verses)

	configDir := t.TempDir()
	if err := config.Save(configDir, config.Config{ModuleDir: dir, ActiveTranslation: "RST"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	entry := `{"book":30,"chapter":1,"verse":1,"at":"2026-10-20T10:00:00Z"}` + "\n"
	if err := os.WriteFile(filepath.Join(configDir, historyFile), []byte(entry), 0o644); err != nil {
		t.Fatal(err)
	}

	m := newReader(t, configDir, dir)
	m, _ = send(t, m, press('H', "H", 0))
	if got := modalText(t, m); !strings.Contains(got, "Leviticus 1:1") {
		t.Fatalf("History lost a Reference the translation lacks\n%s", got)
	}
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))

	got := modalText(t, m)
	if !strings.Contains(got, historyTitle) || !strings.Contains(got, "No such Reference") {
		t.Errorf("History closed or stayed silent on a missing Reference\n%s", got)
	}
}
