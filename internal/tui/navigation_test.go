package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

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

func TestArrowKeysMoveTheVerseCursorToo(t *testing.T) {
	m := navReader(t)

	m, _ = send(t, m, press(tea.KeyDown, "", 0))
	if got := on(m); got != "Genesis 1:2" {
		t.Errorf("down did not move the Verse cursor, the reader is on %q", got)
	}
	m, _ = send(t, m, press(tea.KeyDown, "", 0))
	if got := on(m); got != "Genesis 1:3" {
		t.Errorf("down did not move the Verse cursor again, the reader is on %q", got)
	}

	m, _ = send(t, m, press(tea.KeyUp, "", 0))
	if got := on(m); got != "Genesis 1:2" {
		t.Errorf("up did not move the Verse cursor back, the reader is on %q", got)
	}

	// At the Chapter's first Verse the arrow stops, exactly as k does.
	m, _ = send(t, m, press('g', "g", 0))
	m, _ = send(t, m, press(tea.KeyUp, "", 0))
	if got := on(m); got != "Genesis 1:1" {
		t.Errorf("up crossed the Chapter's start, the reader is on %q", got)
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
	if got := modalText(t, m); strings.Contains(got, "Genesis") {
		t.Errorf("the Book filter kept a non-matching Book\n%s", got)
	}

	m, _ = send(t, m, press(tea.KeyEnter, "", 0))
	if got := on(m); got != "Exodus 1:1" {
		t.Errorf("Select Book did not jump to Exodus, the reader is on %q", got)
	}
}

// The Select Book list shows the Module's own abbreviation beside the canonical
// name — both in the Book's color, with a quiet separator between them — and the
// input narrows the list by either field.
func TestTheBookListShowsAndFiltersByTheModulesAbbreviation(t *testing.T) {
	info, books, verses := navFixture()
	books[0].ShortName = "Gn" // deliberately not a prefix of the canonical name
	m := activeReader(t, info, books, verses)

	m, _ = send(t, m, press('b', "b", 0))
	if got := modalText(t, m); !strings.Contains(got, "Gn") || !strings.Contains(got, "Genesis") {
		t.Fatalf("the Book list does not show the Module's abbreviation\n%s", got)
	}
	if got := modalText(t, m); strings.Contains(got, "|") {
		t.Errorf("the Book list keeps a literal separator\n%s", got)
	}
	if got := m.pickerView(); !strings.Contains(got, m.styles.hint.Render(bookSeparator)) {
		t.Errorf("the Book list does not draw the separator in the quiet grey\n%s", ansi.ReplaceAllString(got, ""))
	}

	m = typeKey(t, m, "Gn")
	got := modalText(t, m)
	if !strings.Contains(got, "Genesis") {
		t.Errorf("the Book input does not match the abbreviation\n%s", got)
	}
	if strings.Contains(got, "Exodus") {
		t.Errorf("the Book input kept a Book the abbreviation does not name\n%s", got)
	}
}

// The Book names line up: the separator sits at the same column in every row.
func TestTheBookListAlignsTheAbbreviations(t *testing.T) {
	info, books, verses := navFixture()
	books[0].ShortName = "Gn" // shorter than the fixture's other abbreviation
	m := activeReader(t, info, books, verses)
	m, _ = send(t, m, press('b', "b", 0))

	columns := map[int]bool{}
	for _, line := range strings.Split(modalText(t, m), "\n") {
		if i := strings.Index(line, bookSeparator); i >= 0 {
			columns[i] = true
		}
	}
	if len(columns) != 1 {
		t.Errorf("the Book rows put the separator at %d different columns, want 1\n%s", len(columns), modalText(t, m))
	}
}

// A Book whose abbreviation is exactly what is typed floats to the top, ahead of
// the Book Canonical order would put first; the rest keep that order.
func TestAnExactAbbreviationMatchFloatsToTheTop(t *testing.T) {
	info := map[string]string{"description": "Order Test"}
	books := []modtest.Book{
		{Number: 150, ShortName: "Esd", LongName: "Ezra", Present: true},
		{Number: 330, ShortName: "Ez", LongName: "Ezekiel", Present: true},
	}
	verses := []modtest.Verse{
		{Book: 150, Chapter: 1, Verse: 1, Text: "ezra"},
		{Book: 330, Chapter: 1, Verse: 1, Text: "ezekiel"},
	}
	m := activeReader(t, info, books, verses)
	m, _ = send(t, m, press('b', "b", 0))
	m = typeKey(t, m, "ez")

	got := modalText(t, m)
	if !strings.Contains(got, "Ez  · Ezekiel") || !strings.Contains(got, "Esd · Ezra") {
		t.Fatalf("the Book list did not match both Books\n%s", got)
	}
	if strings.Index(got, "Ezekiel") > strings.Index(got, "Ezra") {
		t.Errorf("the exact abbreviation match did not float to the top\n%s", got)
	}
}

// Select Book's input is drawn like the Go to Reference one — a chevron on a
// filled band — with no filter label of its own.
func TestTheBookPickerInputLooksLikeTheReferenceField(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('b', "b", 0))

	got := modalText(t, m)
	if !strings.Contains(got, "❯") {
		t.Errorf("the Book input has no chevron\n%s", got)
	}
	if strings.Contains(got, "Filter") {
		t.Errorf("the Book input still carries a filter label\n%s", got)
	}
}

// Every modal that takes a query draws the same chevron on the filled input
// field, and none labels the line with text of its own.
func TestEveryInputModalDrawsTheSameInputField(t *testing.T) {
	for _, key := range []string{"t", "b", "c", "/", "H", "v"} {
		m := navReader(t)
		m, _ = send(t, m, press(rune(key[0]), key, 0))

		got := modalText(t, m)
		if !strings.Contains(got, "❯") {
			t.Errorf("%q: the modal has no input field\n%s", key, got)
		}
		if strings.Contains(got, "Filter: ") || strings.Contains(got, "Search: ") {
			t.Errorf("%q: the modal labels its input line\n%s", key, got)
		}
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
	if popup := m.pickerPopupView(); popup != "" {
		t.Fatalf("v listed every Book before anything was typed\n%s", ansi.ReplaceAllString(popup, ""))
	}

	m = typeKey(t, m, "exo")
	got := modalText(t, m)
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
	if got := render(t, m); !strings.Contains(got, "❯ Exodus ") {
		t.Fatalf("Tab did not write the highlighted Book into the input\n%s", got)
	}

	m = typeKey(t, m, "2:2")
	if got := render(t, m); !strings.Contains(got, "❯ Exodus 2:2") {
		t.Fatalf("the Chapter typed after Tab did not land beside the Book's name\n%s", got)
	}

	m, _ = send(t, m, press(tea.KeyEnter, "", 0))
	if got := on(m); got != "Exodus 2:2" {
		t.Errorf("the completed Reference did not jump, the reader is on %q", got)
	}
}

// Tab completes the Book and takes the completion menu away: the menu did its
// job. A fresh Book name — after backspacing — asks for it again.
func TestTabClosesTheCompletionMenu(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('v', "v", 0))
	m = typeKey(t, m, "exo")
	if popup := ansi.ReplaceAllString(m.pickerPopupView(), ""); !strings.Contains(popup, "Exodus") {
		t.Fatalf("the completion menu did not show the matching Book\n%s", popup)
	}

	m, _ = send(t, m, press(tea.KeyTab, "", 0))
	if popup := m.pickerPopupView(); popup != "" {
		t.Errorf("the completion menu stayed open after Tab completed the Book\n%s", ansi.ReplaceAllString(popup, ""))
	}
	if got := modalText(t, m); !strings.Contains(got, "Exodus 1..2") {
		t.Errorf("Tab did not leave the completed Book in the input\n%s", got)
	}

	m, _ = send(t, m, press(tea.KeyBackspace, "", 0))
	m, _ = send(t, m, press(tea.KeyBackspace, "", 0))
	if popup := ansi.ReplaceAllString(m.pickerPopupView(), ""); !strings.Contains(popup, "Exodus") {
		t.Errorf("a fresh Book name did not bring the completion menu back\n%s", popup)
	}
}

// The Books the input names hang from the field it types into, not from the
// bottom of the box: the line right under the input is the first Book.
func TestTheCompletionMenuHangsFromTheInput(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('v', "v", 0))
	m = typeKey(t, m, "ex")

	lines := strings.Split(render(t, m), "\n")
	field := -1
	for i, line := range lines {
		if strings.Contains(line, "❯ ex") {
			field = i
		}
	}
	if field < 0 || field+1 >= len(lines) {
		t.Fatalf("the Reader did not show the input field\n%s", render(t, m))
	}
	if next := lines[field+1]; !strings.Contains(next, "Exo · Exodus") {
		t.Errorf("the completion menu does not hang from the input: the line under it is %q\n%s", next, render(t, m))
	}
}

// The space is the completion menu's own shortcut: with exactly one Book on it,
// it does what Tab does and leaves the menu closed.
func TestSpaceCompletesTheOnlyBook(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('v', "v", 0))
	m = typeKey(t, m, "ex")
	m, _ = send(t, m, press(tea.KeySpace, " ", 0))

	if got := render(t, m); !strings.Contains(got, "❯ Exodus ") {
		t.Fatalf("space did not write the only Book into the input\n%s", got)
	}
	if popup := m.pickerPopupView(); popup != "" {
		t.Errorf("the completion menu stayed open after space completed the Book\n%s", ansi.ReplaceAllString(popup, ""))
	}
}

// A space with several Books on the menu is ordinary text: it is what lets the
// reader narrow "2" down to "2 Samuel" by writing a second word. Once the input
// names a single Book, the next space completes it.
func TestSpaceWaitsForASingleBook(t *testing.T) {
	books := []modtest.Book{
		{Number: 100, ShortName: "2Sa", LongName: "2 Samuel", Present: true},
		{Number: 120, ShortName: "2Ki", LongName: "2 Kings", Present: true},
	}
	var verses []modtest.Verse
	for _, book := range books {
		for n := 1; n <= 2; n++ {
			verses = append(verses, modtest.Verse{Book: book.Number, Chapter: 1, Verse: n, Text: "verse"})
		}
	}
	m := activeReader(t, map[string]string{"description": "Numbered Books"}, books, verses)
	m, _ = send(t, m, press('v', "v", 0))
	m = typeKey(t, m, "2")

	if popup := ansi.ReplaceAllString(m.pickerPopupView(), ""); !strings.Contains(popup, "2 Samuel") || !strings.Contains(popup, "2 Kings") {
		t.Fatalf("two Books did not leave two matches on the completion menu\n%s", popup)
	}

	m = typeKey(t, m, " ")
	if got := render(t, m); !strings.Contains(got, "❯ 2 ") {
		t.Fatalf("the space with two matches was not typed into the input\n%s", got)
	}

	m = typeKey(t, m, "s")
	m, _ = send(t, m, press(tea.KeySpace, " ", 0))
	if got := render(t, m); !strings.Contains(got, "❯ 2 Samuel ") {
		t.Fatalf("the space did not complete the only Book\n%s", got)
	}
}

// Once the Book is completed the menu is closed, so later spaces are the ones
// the Reference accepts between a Chapter and a Verse.
func TestSpaceSeparatesTheVerseAfterCompletion(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('v', "v", 0))
	m = typeKey(t, m, "gen")
	m = typeKey(t, m, " ")
	m = typeKey(t, m, "2 1")
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))

	if got := on(m); got != "Genesis 2:1" {
		t.Errorf("the space between Chapter and Verse was not kept, the reader is on %q", got)
	}
}

// The example the modal accepts sits above the input, so the Book the input
// names is not written three times in the same view.
func TestTheReferenceExampleSitsAboveTheInput(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('v', "v", 0))

	got := modalText(t, m)
	if strings.Index(got, "For example") > strings.Index(got, "❯") {
		t.Errorf("the example does not sit above the input\n%s", got)
	}
}

// The modal grows to fit a line it has to show and never shrinks back under the
// reader's eyes: its size must not flicker as the reader types. The example line
// repeats the Book's name three times, so it is the widest line the modal needs.
func TestTheReferenceModalOnlyGrows(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('v', "v", 0))
	width := lipgloss.Width(m.modalView())

	// Tab replaces the example line with the Book's short range.
	m = typeKey(t, m, "exo")
	m, _ = send(t, m, press(tea.KeyTab, "", 0))
	if got := lipgloss.Width(m.modalView()); got != width {
		t.Errorf("the modal changed size to %d after Tab, from %d", got, width)
	}

	// A rejection adds a message and must not make the modal jump; a longer
	// message may widen it once, and it keeps that size afterwards.
	m = typeKey(t, m, "999")
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))
	width = max(width, lipgloss.Width(m.modalView()))

	// And it keeps that size: deleting the rejected input must not shrink it.
	m, _ = send(t, m, press(tea.KeyBackspace, "", 0))
	if got := lipgloss.Width(m.modalView()); got != width {
		t.Errorf("the modal changed size to %d after the message, from %d", got, width)
	}
}

func TestTabKeepsTheChapterAlreadyTyped(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('v', "v", 0))
	m = typeKey(t, m, "exo 2:2")
	m, _ = send(t, m, press(tea.KeyTab, "", 0))

	if got := render(t, m); !strings.Contains(got, "❯ Exodus 2:2") {
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
// prefix the list would have shown, is rejected. Completion is the way to use
// the list's highlight.
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
	if got := render(t, m); !strings.Contains(got, "For example: Genesis, Genesis 1, Genesis 1:1") {
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

// pickedOut is the ANSI sequence that opens a run the Reference modal picks out
// in red. The run sits on the input's filled band, so the band's background is
// already in effect; the compositor therefore emits the red foreground alone and
// the number's own text follows. The red is the theme's one red, the color the
// words of Jesus carry, which the picked-out style shares.
func pickedOut(m Model, run string) string {
	const probe = "\x00"
	rendered := m.styles.jesus.Render(probe)
	return rendered[:strings.IndexByte(rendered, 0)] + run
}

// A number the Book has no Chapter or Verse for is picked out, whole, in red as it
// is typed, without waiting for Enter, and the input otherwise reads as typed.
func TestTheReferenceViewPicksOutANumberTheBookLacks(t *testing.T) {
	cases := []struct{ typed, bad string }{
		{"Genesis 50", "50"}, // the fixture's Genesis has 3 Chapters
		{"Genesis 0", "0"},
		{"Genesis 3:999", "999"}, // Chapter 3 has 2 Verses
		{"Genesis 3:0", "0"},
	}
	for _, c := range cases {
		m := navReader(t)
		m, _ = send(t, m, press('v', "v", 0))
		m = typeKey(t, m, c.typed)

		raw := m.View().Content
		if !strings.Contains(raw, pickedOut(m, c.bad)) {
			t.Errorf("%q: the number %q was not picked out in red\n%s", c.typed, c.bad, ansi.ReplaceAllString(raw, ""))
		}
		if got := ansi.ReplaceAllString(raw, ""); !strings.Contains(got, "❯ "+c.typed) {
			t.Errorf("%q: the input reads as something else\n%s", c.typed, got)
		}
	}
}

func TestTheReferenceViewLeavesANumberTheBookHasAlone(t *testing.T) {
	cases := []struct {
		typed   string
		numbers []string
	}{
		{"gen 3", []string{"3"}},
		{"gen 3:2", []string{"3", "2"}},
		{"gen 1:20", []string{"1", "20"}},
	}
	for _, c := range cases {
		m := navReader(t)
		m, _ = send(t, m, press('v', "v", 0))
		m = typeKey(t, m, c.typed)

		raw := m.View().Content
		for _, number := range c.numbers {
			if strings.Contains(raw, pickedOut(m, number)) {
				t.Errorf("%q: the number %q was picked out though the Book has it\n%s", c.typed, number, ansi.ReplaceAllString(raw, ""))
			}
		}
	}
}

func TestTheReferencePickerRejectsMalformedInput(t *testing.T) {
	m := navReader(t)
	m, _ = send(t, m, press('v', "v", 0))
	m = typeKey(t, m, "banana")
	m, _ = send(t, m, press(tea.KeyTab, "", 0)) // no row to complete from
	if got := render(t, m); !strings.Contains(got, "❯ banana") {
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
	if !strings.Contains(m.barView(), "Genesis 1") {
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
