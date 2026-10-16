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

// activeReader builds a reader over a generated Module that is already the
// Active translation, the way a returning user's config would leave it.
func activeReader(t *testing.T, info map[string]string, books []modtest.Book, verses []modtest.Verse) Model {
	t.Helper()

	dir := t.TempDir()
	modtest.BibleAll(t, dir, "RST.SQLite3", info, books, verses)

	configDir := t.TempDir()
	if err := config.Save(configDir, config.Config{ModuleDir: dir, ActiveTranslation: "RST"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	return newReader(t, configDir, dir)
}

func genesis() []modtest.Book {
	return []modtest.Book{{Number: 10, ShortName: "Gen", LongName: "Genesis", Present: true}}
}

// hasSequence reports whether want appears as consecutive rendered lines,
// ignoring the padding the reading pane adds to each line.
func hasSequence(got string, want ...string) bool {
	lines := strings.Split(got, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " ")
	}
	for i := 0; i+len(want) <= len(lines); i++ {
		match := true
		for j, w := range want {
			if lines[i+j] != w {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func TestReadingRendersOneChapterWithVerseNumbers(t *testing.T) {
	m := activeReader(t, map[string]string{"description": "Reina-Valera 1909"}, genesis(),
		[]modtest.Verse{
			{Book: 10, Chapter: 1, Verse: 1, Text: "In the beginning <i>God</i> created the heavens"},
			{Book: 10, Chapter: 1, Verse: 2, Text: "And the earth was <J>without form</J>"},
			{Book: 10, Chapter: 2, Verse: 1, Text: "Thus the heavens were finished"},
		})

	got := render(t, m)

	if !strings.Contains(got, "Genesis 1") {
		t.Errorf("the top bar does not show the current Reference\n%s", got)
	}
	for _, want := range []string{
		"1 In the beginning God created the heavens",
		"2 And the earth was without form",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("the Chapter is missing %q\n%s", want, got)
		}
	}
	if strings.Contains(got, "Thus the heavens were finished") {
		t.Errorf("only the current Chapter should be rendered\n%s", got)
	}
	for _, unwanted := range []string{"<i>", "</i>", "<J>", "</J>"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("markup %q was rendered literally\n%s", unwanted, got)
		}
	}
}

func TestReadingPreservesParagraphAndLineBreaks(t *testing.T) {
	m := activeReader(t, map[string]string{"description": "Reina-Valera 1909"}, genesis(),
		[]modtest.Verse{
			{Book: 10, Chapter: 1, Verse: 1, Text: "line one<br/>line two"},
			{Book: 10, Chapter: 1, Verse: 2, Text: "first<pb/>second"},
		})

	got := render(t, m)

	if !hasSequence(got, "1 line one", "  line two") {
		t.Errorf("<br/> did not produce a line break\n%q", got)
	}
	if !hasSequence(got, "2 first", "", "  second") {
		t.Errorf("<pb/> did not produce a paragraph break\n%q", got)
	}
}

// A Module marks the Verse that opens a paragraph by starting its text with a
// paragraph break; that break must not leave the Verse number on a line of its
// own.
func TestReadingDoesNotBreakAVerseThatStartsAParagraph(t *testing.T) {
	m := activeReader(t, map[string]string{"description": "Reina-Valera 1909"}, genesis(),
		[]modtest.Verse{
			{Book: 10, Chapter: 1, Verse: 1, Text: "<pb/>first"},
			{Book: 10, Chapter: 1, Verse: 2, Text: "<pb/><t>second</t> and <br/>on two lines"},
			{Book: 10, Chapter: 1, Verse: 3, Text: "third"},
		})

	got := render(t, m)

	if !hasSequence(got, "1 first", "2 second and") {
		t.Errorf("a leading paragraph break pushed the Verse number onto its own line\n%q", got)
	}
	if !hasSequence(got, "2 second and", "  on two lines") {
		t.Errorf("a leading paragraph break stopped the line break inside the Verse\n%q", got)
	}
	if !hasSequence(got, "  on two lines", "3 third") {
		t.Errorf("the Verse after a broken one did not render\n%q", got)
	}
}

func TestReadingAnEmptyVerseRendersWithoutError(t *testing.T) {
	m := activeReader(t, map[string]string{"description": "Reina-Valera 1909"}, genesis(),
		[]modtest.Verse{
			{Book: 10, Chapter: 1, Verse: 1, Text: "In the beginning"},
			{Book: 10, Chapter: 1, Verse: 2, Text: ""},
			{Book: 10, Chapter: 1, Verse: 3, Text: "And there was light"},
		})

	got := render(t, m)

	if !hasSequence(got, "2") {
		t.Errorf("an empty Verse did not render its number on its own line\n%q", got)
	}
	if !hasSequence(got, "3 And there was light") {
		t.Errorf("reading stopped after an empty Verse\n%q", got)
	}
}

func TestReadingWrapsAVerseThatOverflowsThePane(t *testing.T) {
	words := make([]string, 40)
	for i := range words {
		words[i] = fmt.Sprintf("word%02d", i)
	}
	m := activeReader(t, map[string]string{"description": "Reina-Valera 1909"}, genesis(),
		[]modtest.Verse{{Book: 10, Chapter: 1, Verse: 1, Text: strings.Join(words, " ")}})

	got := render(t, m)
	if !strings.Contains(got, "word39") {
		t.Errorf("the end of a long Verse was cut off\n%s", got)
	}

	wrapped := false
	for _, line := range strings.Split(got, "\n") {
		line = strings.TrimRight(line, " ")
		if lipgloss.Width(line) > 100 {
			t.Errorf("a rendered line is %d cells wide, want at most the pane's 100: %q", lipgloss.Width(line), line)
		}
		if strings.HasPrefix(line, "  word") {
			wrapped = true
		}
	}
	if !wrapped {
		t.Errorf("no continuation line is indented under the Verse's text\n%s", got)
	}
}

func TestReadingOpensWithACentredChapterHeading(t *testing.T) {
	m := activeReader(t, map[string]string{"description": "Reina-Valera 1909"}, genesis(),
		[]modtest.Verse{{Book: 10, Chapter: 1, Verse: 1, Text: "In the beginning"}})

	got := render(t, m)
	found := false
	for _, line := range strings.Split(got, "\n") {
		i := strings.Index(line, "Chapter 1")
		if i < 0 {
			continue
		}
		found = true
		if i < 40 || i > 50 {
			t.Errorf("the Chapter heading is not centred: it starts at column %d\n%s", i, got)
		}
	}
	if !found {
		t.Errorf("the Chapter has no heading\n%s", got)
	}
}

func TestReadingShowsASubheadingAboveItsVerse(t *testing.T) {
	dir := t.TempDir()
	modtest.BibleAll(t, dir, "RST.SQLite3", map[string]string{"description": "Reina-Valera 1909"}, genesis(),
		[]modtest.Verse{
			{Book: 10, Chapter: 1, Verse: 1, Text: "In the beginning"},
			{Book: 10, Chapter: 1, Verse: 2, Text: "And the earth was without form"},
		})
	modtest.Subheadings(t, dir, "RST-s.subheadings.SQLite3", map[string]string{"description": "Títulos"},
		[]modtest.Subheading{{Book: 10, Chapter: 1, Verse: 2, Order: 1, Text: "The creation"}})

	configDir := t.TempDir()
	if err := config.Save(configDir, config.Config{ModuleDir: dir, ActiveTranslation: "RST"}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got := render(t, newReader(t, configDir, dir))
	if !hasSequence(got, "The creation", "2 And the earth was without form") {
		t.Errorf("the Subheading is not on its own line above its Verse\n%s", got)
	}
}

func TestSelectingATranslationLoadsItsChapter(t *testing.T) {
	dir := t.TempDir()
	modtest.BibleAll(t, dir, "KJV.SQLite3", map[string]string{"description": "King James Version"}, genesis(),
		[]modtest.Verse{{Book: 10, Chapter: 1, Verse: 1, Text: "In the beginning God created"}})

	m := newReader(t, t.TempDir(), dir)
	m, _ = send(t, m, press('t', "t", 0))
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))

	if got := render(t, m); !strings.Contains(got, "1 In the beginning God created") {
		t.Errorf("selecting a Translation did not load its Chapter\n%s", got)
	}
}
