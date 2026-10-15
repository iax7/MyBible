package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
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

// assertReference checks the top bar's current Reference.
func assertReference(t *testing.T, m Model, want string) {
	t.Helper()
	if got := render(t, m); !strings.Contains(got, want) {
		t.Errorf("reader is not on %q\n%s", want, got)
	}
}

func TestBackAndForwardRetraceJumps(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press(']', "]", 0)) // Genesis 2:1
	m, _ = send(t, m, press(']', "]", 0)) // Genesis 3:1
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
	m, _ = send(t, m, press(']', "]", 0)) // Genesis 2:1: history [Genesis 1:1, Genesis 2:1]
	m, _ = send(t, m, press('j', "j", 0)) // Genesis 2:2: a bare cursor move
	assertReference(t, m, "Genesis 2:2")

	m = back(t, m)
	assertReference(t, m, "Genesis 1:1") // the jump's origin, not the cursor position

	m = forward(t, m)
	assertReference(t, m, "Genesis 2:1") // the visited Reference, not the cursor position
}

func TestHistoryModalListsJumpsMostRecentFirst(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press(']', "]", 0)) // Genesis 2:1      next Chapter
	m, _ = send(t, m, press('c', "c", 0))
	m = typeKey(t, m, "3")
	m, _ = send(t, m, press(tea.KeyEnter, "", 0)) // Genesis 3:1      Select Chapter
	m, _ = send(t, m, press('b', "b", 0))
	m = typeKey(t, m, "exo")
	m, _ = send(t, m, press(tea.KeyEnter, "", 0)) // Exodus 1:1       Select Book
	m, _ = send(t, m, press('v', "v", 0))
	m = typeKey(t, m, "Genesis 1:5")
	m, _ = send(t, m, press(tea.KeyEnter, "", 0)) // Genesis 1:5      Go to Reference
	m, _ = send(t, m, press('/', "/", 0))
	m = typeKey(t, m, "exo02")
	m, _ = send(t, m, press(tea.KeyEnter, "", 0)) // Exodus 2:1       Search result

	m, _ = send(t, m, press('H', "H", 0))
	got := render(t, m)
	if !strings.Contains(got, "Filter: ") {
		t.Fatalf("H did not open the History modal\n%s", got)
	}

	order := []string{"Exodus 2:1", "Genesis 1:5", "Exodus 1:1", "Genesis 3:1", "Genesis 2:1", "Genesis 1:1"}
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
	if got := render(t, m); strings.Contains(got, "Filter: ") {
		t.Errorf("the History modal stayed open after Enter\n%s", got)
	}
}

func TestHistoryModalIsFilterable(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press(']', "]", 0)) // Genesis 2:1
	m, _ = send(t, m, press('b', "b", 0))
	m = typeKey(t, m, "exo")
	m, _ = send(t, m, press(tea.KeyEnter, "", 0)) // Exodus 1:1

	m, _ = send(t, m, press('H', "H", 0))
	m = typeKey(t, m, "gen")

	got := render(t, m)
	if !strings.Contains(got, "Genesis 2:1") || !strings.Contains(got, "Genesis 1:1") {
		t.Errorf("the History filter dropped a matching Reference\n%s", got)
	}
	if strings.Contains(got, "Exodus") {
		t.Errorf("the History filter kept a non-matching Reference\n%s", got)
	}
}
