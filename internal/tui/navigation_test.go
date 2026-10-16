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
	if got := on(m); got != "Genesis 1:1" {
		t.Fatalf("the reader did not start on Genesis 1:1, it is on %q", got)
	}

	m, _ = send(t, m, press('j', "j", 0))
	if got := on(m); got != "Genesis 1:2" {
		t.Errorf("j did not move the Verse cursor, the reader is on %q", got)
	}

	m, _ = send(t, m, press('k', "k", 0))
	if got := on(m); got != "Genesis 1:1" {
		t.Errorf("k did not move the Verse cursor back, the reader is on %q", got)
	}
}

func TestTheCursorKeepsTheCurrentVerseVisible(t *testing.T) {
	m := resized(t, navReader(t), 100, 8)
	m, _ = send(t, m, press('G', "G", 0))

	if got := on(m); got != "Genesis 1:20" {
		t.Fatalf("G did not move the cursor to the last Verse, the reader is on %q", got)
	}
	got := render(t, m)
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
	if got := on(m); got != "Genesis 1:12" {
		t.Errorf("Ctrl-D did not page down, the reader is on %q", got)
	}

	m, _ = send(t, m, press('u', "", tea.ModCtrl))
	if got := on(m); got != "Genesis 1:1" {
		t.Errorf("Ctrl-U did not page back up, the reader is on %q", got)
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
		if got := on(m); got != step.want {
			t.Errorf("%s reached %q, want %q", step.key, got, step.want)
		}
	}
}

func TestTopAndBottomJumpToTheChapterEnds(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('G', "G", 0))
	if got := on(m); got != "Genesis 1:20" {
		t.Errorf("G did not jump to the bottom of the Chapter, the reader is on %q", got)
	}

	m, _ = send(t, m, press('g', "g", 0))
	if got := on(m); got != "Genesis 1:1" {
		t.Errorf("g did not jump to the top of the Chapter, the reader is on %q", got)
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
	if got := on(m); got != "Exodus 1:1" {
		t.Errorf("Select Book did not jump to Exodus, the reader is on %q", got)
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
	if got := on(m); got != "Genesis 3:1" {
		t.Errorf("Select Chapter did not jump to Chapter 3, the reader is on %q", got)
	}
}

func TestTheReferencePickerJumpsToAVerse(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('v', "v", 0))
	m = typeKey(t, m, "Exodus 2:2")
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))

	if got := on(m); got != "Exodus 2:2" {
		t.Errorf("Go to Reference did not jump to Exodus 2:2, the reader is on %q", got)
	}
}

// The modal lists nothing until the input names a Book: listing them all is what
// b does.
func TestTheReferencePickerListsTheBooksTheInputNames(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('v', "v", 0))
	if got := render(t, m); strings.Contains(got, "Genesis") || strings.Contains(got, "Exodus") {
		t.Fatalf("v listed every Book before anything was typed\n%s", got)
	}

	m = typeKey(t, m, "exo")
	got := render(t, m)
	if strings.Contains(got, "Genesis") {
		t.Errorf("the Book list kept a Book the input does not name\n%s", got)
	}
	if !strings.Contains(got, "Exodus") {
		t.Errorf("the Book list does not show the Book the input names\n%s", got)
	}
}

func TestTheReferencePickerCompletesWithTab(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('v', "v", 0))
	m = typeKey(t, m, "exo")
	m, _ = send(t, m, press(tea.KeyTab, "", 0))

	// The trailing space keeps the Chapter off the Book's name.
	if got := render(t, m); !strings.Contains(got, "Reference: Exodus ") {
		t.Fatalf("Tab did not write the highlighted Book into the input\n%s", got)
	}

	m = typeKey(t, m, "2:2")
	if got := render(t, m); !strings.Contains(got, "Reference: Exodus 2:2") {
		t.Fatalf("the Chapter typed after Tab did not land beside the Book's name\n%s", got)
	}

	m, _ = send(t, m, press(tea.KeyEnter, "", 0))
	if got := on(m); got != "Exodus 2:2" {
		t.Errorf("the completed Reference did not jump, the reader is on %q", got)
	}
}

func TestTabKeepsTheChapterAlreadyTyped(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('v', "v", 0))
	m = typeKey(t, m, "exo 2:2")
	m, _ = send(t, m, press(tea.KeyTab, "", 0))

	if got := render(t, m); !strings.Contains(got, "Reference: Exodus 2:2") {
		t.Fatalf("Tab lost the Chapter:Verse already typed\n%s", got)
	}

	m, _ = send(t, m, press(tea.KeyEnter, "", 0))
	if got := on(m); got != "Exodus 2:2" {
		t.Errorf("the completed Reference did not jump, the reader is on %q", got)
	}
}

func TestTheReferencePickerTakesABookOnItsOwn(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('v', "v", 0))
	m = typeKey(t, m, "Exodus")
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))

	if got := on(m); got != "Exodus 1:1" {
		t.Errorf("a Book on its own did not jump to Chapter 1, Verse 1, the reader is on %q", got)
	}
}

func TestTheReferencePickerTakesAChapterWithoutAVerse(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('v', "v", 0))
	m = typeKey(t, m, "Exodus 2")
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))

	if got := on(m); got != "Exodus 2:1" {
		t.Errorf("a Chapter without a Verse did not jump to Verse 1, the reader is on %q", got)
	}
}

// A Book the Module names itself — "Gen" is the fixture's short name for
// Genesis — is resolved when a single Book answers to it.
func TestTheReferencePickerTakesTheModulesOwnBookName(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('v', "v", 0))
	m = typeKey(t, m, "Gen 2")
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))

	if got := on(m); got != "Genesis 2:1" {
		t.Errorf("the Module's own Book name was not resolved, the reader is on %q", got)
	}
}

// An input naming no Book is read as the translation's first Book.
func TestTheReferencePickerTakesTheFirstBookWhenTheInputNamesNone(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press(']', "]", 0)) // away from Genesis 1:1
	m, _ = send(t, m, press('v', "v", 0))
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))

	if got := on(m); got != "Genesis 1:1" {
		t.Errorf("an empty input did not read as the first Book, the reader is on %q", got)
	}

	m, _ = send(t, m, press('v', "v", 0))
	m = typeKey(t, m, "2:2")
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))
	if got := on(m); got != "Genesis 2:2" {
		t.Errorf("a Chapter:Verse with no Book did not read as the first Book's, the reader is on %q", got)
	}

	m, _ = send(t, m, press('v', "v", 0))
	m = typeKey(t, m, "2 2") // a space divides a Chapter from a Verse too
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))
	if got := on(m); got != "Genesis 2:2" {
		t.Errorf("a Chapter and Verse divided by a space did not read as the first Book's, the reader is on %q", got)
	}
}

// Enter never guesses from the list: a name no Book has to answer to, even a
// prefix the list would have shown, is rejected. Tab is the way to complete it.
func TestTheReferencePickerDoesNotGuessAPrefix(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('v', "v", 0))
	m = typeKey(t, m, "Ex")
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))

	if got := render(t, m); !strings.Contains(got, "not a Reference") {
		t.Errorf("a Book name the translation does not have was not rejected\n%s", got)
	}
	if got := on(m); got != "Genesis 1:1" {
		t.Errorf("the reader moved after a rejected input, it is on %q", got)
	}
}

// Two Books claiming the same Module name is no name at all: the reader reports
// it rather than picking one.
func TestTheReferencePickerRejectsAnAmbiguousModulesOwnName(t *testing.T) {
	info, books, verses := navFixture()
	for i := range books {
		books[i].ShortName = "Xx"
	}
	m := activeReader(t, info, books, verses)

	m, _ = send(t, m, press('v', "v", 0))
	m = typeKey(t, m, "Xx 2")
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))

	if got := render(t, m); !strings.Contains(got, "not a Reference") {
		t.Errorf("a Book name two Books answer to was not rejected\n%s", got)
	}
}

func TestTheReferencePickerTakesEveryChapterVerseSeparator(t *testing.T) {
	for _, typed := range []string{"Exodus 2:2", "Exodus 2,2", "Exodus 2.2", "Exodus 2 2"} {
		m := navReader(t)
		m, _ = send(t, m, press('v', "v", 0))
		m = typeKey(t, m, typed)
		m, _ = send(t, m, press(tea.KeyEnter, "", 0))

		if got := on(m); got != "Exodus 2:2" {
			t.Errorf("%q did not reach Exodus 2:2, the reader is on %q", typed, got)
		}
	}
}

// The modal says what it accepts: the shapes to type while no Book is named, and
// the range of Chapters the named Book has once one is.
func TestTheReferenceViewShowsTheChaptersTheNamedBookAccepts(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('v', "v", 0))
	if got := render(t, m); !strings.Contains(got, "For example: John, John 3, John 3:16") {
		t.Fatalf("the modal does not say what to type before a Book is named\n%s", got)
	}

	m = typeKey(t, m, "exo")
	m, _ = send(t, m, press(tea.KeyTab, "", 0))
	if got := render(t, m); !strings.Contains(got, "Exodus 1..2") {
		t.Errorf("the modal does not show the Chapters Exodus accepts\n%s", got)
	}
}

// A Chapter still being typed may grow into a longer one, so the Verses appear
// only once a separator closes it — the space included — and a rejected Reference
// keeps the range it should have fallen in beside the reason.
func TestTheReferenceViewShowsTheVersesTheClosedChapterAccepts(t *testing.T) {
	for _, closer := range []string{":", ",", ".", " "} {
		m := navReader(t)
		m, _ = send(t, m, press('v', "v", 0))
		m = typeKey(t, m, "gen 3")
		if got := render(t, m); !strings.Contains(got, "Genesis 1..3") {
			t.Fatalf("a Chapter still being typed did not keep the Book's Chapter range\n%s", got)
		}

		m = typeKey(t, m, closer)
		if got := render(t, m); !strings.Contains(got, "Genesis 3:1..2") {
			t.Fatalf("%q did not close the Chapter and show its Verses\n%s", closer, got)
		}
	}
}

func TestTheReferenceViewKeepsTheRangeOfARejectedVerse(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('v', "v", 0))
	m = typeKey(t, m, "gen 3:999")
	if got := render(t, m); !strings.Contains(got, "Genesis 3:1..2") {
		t.Fatalf("an impossible Verse hid the range it should fall in\n%s", got)
	}

	m, _ = send(t, m, press(tea.KeyEnter, "", 0))
	got := render(t, m)
	if !strings.Contains(got, "No such Reference") {
		t.Errorf("a Verse the translation lacks was not rejected\n%s", got)
	}
	if !strings.Contains(got, "Genesis 3:1..2") {
		t.Errorf("the rejection does not name the range the Verse should be in\n%s", got)
	}
}

func TestTheReferencePickerRejectsMalformedInput(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('v', "v", 0))
	m = typeKey(t, m, "banana")
	m, _ = send(t, m, press(tea.KeyTab, "", 0)) // no row to complete from
	if got := render(t, m); !strings.Contains(got, "Reference: banana") {
		t.Errorf("Tab changed the input though no Book matched it\n%s", got)
	}

	m, _ = send(t, m, press(tea.KeyEnter, "", 0))
	if got := render(t, m); !strings.Contains(got, "not a Reference") {
		t.Errorf("malformed input was not rejected\n%s", got)
	}

	m, _ = send(t, m, press(tea.KeyEscape, "", 0))
	if got := on(m); got != "Genesis 1:1" {
		t.Errorf("the reader did not stay on its Reference after a rejection, it is on %q", got)
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
	if !strings.Contains(got, keyReference+" Genesis 1") {
		t.Errorf("the Reference is not shown on a narrow terminal\n%s", got)
	}
	if !strings.Contains(got, barCommands[0].Label) {
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
	if got := on(m); got != "Genesis 2:2" {
		t.Fatalf("navigation did not reach Genesis 2:2, the reader is on %q", got)
	}

	send(t, m, press('q', "q", 0)) // the reader persists on quit

	if got := on(newReader(t, configDir, dir)); got != "Genesis 2:2" {
		t.Errorf("the last-read Reference was not restored, the reader is on %q", got)
	}
}
