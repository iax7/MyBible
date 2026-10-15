package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/iax7/mybible/internal/config"
	"github.com/iax7/mybible/internal/modtest"
)

// navFixture is a small two-Book Bible: Genesis with 3 Chapters and Exodus with
// 2, enough to exercise Chapter and Book boundaries.
func navFixture() (map[string]string, []modtest.Book, []modtest.Verse) {
	books := []modtest.Book{
		{Number: 10, ShortName: "Gen", LongName: "Genesis", Present: true},
		{Number: 20, ShortName: "Exo", LongName: "Exodus", Present: true},
	}
	verses := make([]modtest.Verse, 0, 26)
	for n := 1; n <= 20; n++ {
		verses = append(verses, modtest.Verse{Book: 10, Chapter: 1, Verse: n, Text: fmt.Sprintf("verse%02d", n)})
	}
	for _, ch := range []int{2, 3} {
		for n := 1; n <= 2; n++ {
			verses = append(verses, modtest.Verse{Book: 10, Chapter: ch, Verse: n, Text: fmt.Sprintf("gen%02d%02d", ch, n)})
		}
	}
	for _, ch := range []int{1, 2} {
		for n := 1; n <= 2; n++ {
			verses = append(verses, modtest.Verse{Book: 20, Chapter: ch, Verse: n, Text: fmt.Sprintf("exo%02d%02d", ch, n)})
		}
	}
	return map[string]string{"description": "Navigation Test"}, books, verses
}

func navReader(t *testing.T) Model {
	t.Helper()
	info, books, verses := navFixture()
	return activeReader(t, info, books, verses)
}

func resized(t *testing.T, m Model, width, height int) Model {
	t.Helper()
	m, _ = send(t, m, tea.WindowSizeMsg{Width: width, Height: height})
	return m
}

func TestVerseCursorMovesWithJAndK(t *testing.T) {
	m := navReader(t)
	if got := render(t, m); !strings.Contains(got, "Genesis 1:1") {
		t.Fatalf("the reader did not start on Genesis 1:1\n%s", got)
	}

	m, _ = send(t, m, press('j', "j", 0))
	if got := render(t, m); !strings.Contains(got, "Genesis 1:2") {
		t.Errorf("j did not move the Verse cursor\n%s", got)
	}

	m, _ = send(t, m, press('k', "k", 0))
	if got := render(t, m); !strings.Contains(got, "Genesis 1:1") {
		t.Errorf("k did not move the Verse cursor back\n%s", got)
	}
}

func TestTheCursorKeepsTheCurrentVerseVisible(t *testing.T) {
	m := resized(t, navReader(t), 100, 8)
	m, _ = send(t, m, press('G', "G", 0))

	got := render(t, m)
	if !strings.Contains(got, "Genesis 1:20") {
		t.Fatalf("G did not move the cursor to the last Verse\n%s", got)
	}
	if !strings.Contains(got, "verse20") {
		t.Errorf("the current Verse is not visible after G\n%s", got)
	}
	if strings.Contains(got, "verse01") {
		t.Errorf("the first Verse is still shown though the cursor moved past it\n%s", got)
	}
}

func TestPageDownAndPageUpMoveWithinTheChapter(t *testing.T) {
	// A width wide enough for the top bar to stay on one line, so the pane is
	// 11 rows and a page is 11 Verses.
	m := resized(t, navReader(t), 200, 12)
	m, _ = send(t, m, press('d', "", tea.ModCtrl))
	if got := render(t, m); !strings.Contains(got, "Genesis 1:12") {
		t.Errorf("Ctrl-D did not page down\n%s", got)
	}

	m, _ = send(t, m, press('u', "", tea.ModCtrl))
	if got := render(t, m); !strings.Contains(got, "Genesis 1:1") {
		t.Errorf("Ctrl-U did not page back up\n%s", got)
	}
}

func TestChapterMovementCrossesBookBoundaries(t *testing.T) {
	m := navReader(t)
	steps := []struct{ key, want string }{
		{"]", "Genesis 2:1"},
		{"]", "Genesis 3:1"},
		{"]", "Exodus 1:1"},
		{"[", "Genesis 3:1"},
		{"[", "Genesis 2:1"},
		{"[", "Genesis 1:1"},
		{"[", "Genesis 1:1"}, // already the first Chapter in the Canonical order
	}
	for _, step := range steps {
		m, _ = send(t, m, press(rune(step.key[0]), step.key, 0))
		if got := render(t, m); !strings.Contains(got, step.want) {
			t.Errorf("%s did not reach %q\n%s", step.key, step.want, got)
		}
	}
}

func TestTopAndBottomJumpToTheChapterEnds(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('G', "G", 0))
	if got := render(t, m); !strings.Contains(got, "Genesis 1:20") {
		t.Errorf("G did not jump to the bottom of the Chapter\n%s", got)
	}

	m, _ = send(t, m, press('g', "g", 0))
	if got := render(t, m); !strings.Contains(got, "Genesis 1:1") {
		t.Errorf("g did not jump to the top of the Chapter\n%s", got)
	}
}

func TestTheBookPickerJumpsToABook(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('b', "b", 0))
	if got := render(t, m); !strings.Contains(got, selectBook) {
		t.Fatalf("b did not open the Select Book modal\n%s", got)
	}

	m = typeKey(t, m, "exo")
	if got := render(t, m); strings.Contains(got, "Genesis") {
		t.Errorf("the Book filter kept a non-matching Book\n%s", got)
	}

	m, _ = send(t, m, press(tea.KeyEnter, "", 0))
	if got := render(t, m); !strings.Contains(got, "Exodus 1:1") {
		t.Errorf("Select Book did not jump to Exodus\n%s", got)
	}
}

func TestTheChapterPickerJumpsToAChapter(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('c', "c", 0))
	if got := render(t, m); !strings.Contains(got, selectChapter) {
		t.Fatalf("c did not open the Select Chapter modal\n%s", got)
	}

	m = typeKey(t, m, "3")
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))
	if got := render(t, m); !strings.Contains(got, "Genesis 3:1") {
		t.Errorf("Select Chapter did not jump to Chapter 3\n%s", got)
	}
}

func TestTheReferencePickerJumpsToAVerse(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('v', "v", 0))
	m = typeKey(t, m, "Exodus 2:2")
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))

	if got := render(t, m); !strings.Contains(got, "Exodus 2:2") {
		t.Errorf("Go to Reference did not jump to Exodus 2:2\n%s", got)
	}
}

func TestTheReferencePickerRejectsMalformedInput(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('v', "v", 0))
	m = typeKey(t, m, "banana")
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))

	if got := render(t, m); !strings.Contains(got, "not a Reference") {
		t.Errorf("malformed input was not rejected\n%s", got)
	}

	m, _ = send(t, m, press(tea.KeyEscape, "", 0))
	if got := render(t, m); !strings.Contains(got, "Genesis 1:1") {
		t.Errorf("the reader did not stay on its Reference after a rejection\n%s", got)
	}
}

func TestTheReferencePickerRejectsAVerseTheTranslationLacks(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('v', "v", 0))
	m = typeKey(t, m, "John 3:16")
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))

	if got := render(t, m); !strings.Contains(got, "No such Reference") {
		t.Errorf("a Reference outside the Active translation was not rejected\n%s", got)
	}
}

func TestTheBarShowsTheReferenceOnANarrowTerminal(t *testing.T) {
	m := resized(t, navReader(t), 80, 24)
	got := render(t, m)
	if !strings.Contains(got, "Genesis 1:1") {
		t.Errorf("the Reference is not shown on a narrow terminal\n%s", got)
	}
	if !strings.Contains(got, "Quit") {
		t.Errorf("the commands are not shown on a narrow terminal\n%s", got)
	}
}

func TestTheLastReadReferenceIsRestoredOnTheNextLaunch(t *testing.T) {
	info, books, verses := navFixture()
	dir := t.TempDir()
	modtest.BibleAll(t, dir, "RST.SQLite3", info, books, verses)

	configDir := t.TempDir()
	if err := config.Save(configDir, config.Config{ModuleDir: dir, ActiveTranslation: "RST"}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	m := newReader(t, configDir, dir)
	m, _ = send(t, m, press(']', "]", 0)) // Genesis 2:1
	m, _ = send(t, m, press('j', "j", 0)) // Genesis 2:2
	if got := render(t, m); !strings.Contains(got, "Genesis 2:2") {
		t.Fatalf("navigation did not reach Genesis 2:2\n%s", got)
	}

	send(t, m, press('q', "q", 0)) // the reader persists on quit

	if got := render(t, newReader(t, configDir, dir)); !strings.Contains(got, "Genesis 2:2") {
		t.Errorf("the last-read Reference was not restored\n%s", got)
	}
}
