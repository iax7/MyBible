package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/iax7/mybible/internal/config"
	"github.com/iax7/mybible/internal/modtest"
)

// searchBooks is the Book set the Search tests load.
func searchBooks() []modtest.Book {
	return []modtest.Book{
		{Number: 10, ShortName: "Gen", LongName: "Genesis", Present: true},
		{Number: 500, ShortName: "Jhn", LongName: "John", Present: true},
	}
}

func TestSearchOpensAndCloses(t *testing.T) {
	m := activeReader(t, map[string]string{"description": "Reina-Valera"}, searchBooks(),
		[]modtest.Verse{{Book: 10, Chapter: 1, Verse: 1, Text: "En el principio"}})

	m, _ = send(t, m, press('/', "/", 0))
	if got := render(t, m); !strings.Contains(got, searchTitle+": ") {
		t.Fatalf("/ did not open the Search modal\n%s", got)
	}

	m, _ = send(t, m, press(tea.KeyEscape, "", 0))
	if got := render(t, m); strings.Contains(got, searchTitle+": ") {
		t.Errorf("Esc did not close the Search modal\n%s", got)
	}
}

func TestSearchFindsAccentedTextAndEnterJumpsToIt(t *testing.T) {
	m := activeReader(t, map[string]string{"description": "Reina-Valera"}, searchBooks(),
		[]modtest.Verse{
			{Book: 10, Chapter: 1, Verse: 1, Text: "En el principio"},
			{Book: 500, Chapter: 11, Verse: 35, Text: "Jesús lloró."},
		})

	m, _ = send(t, m, press('/', "/", 0))
	m = typeKey(t, m, "jesus")

	got := render(t, m)
	if !strings.Contains(got, "John 11:35") {
		t.Errorf("the result is missing its Reference\n%s", got)
	}
	if !strings.Contains(got, "Jesús lloró") {
		t.Errorf("the result is missing its snippet\n%s", got)
	}

	m, _ = send(t, m, press(tea.KeyEnter, "", 0))
	got = render(t, m)
	if strings.Contains(got, searchTitle+": ") {
		t.Errorf("the Search modal stayed open after Enter\n%s", got)
	}
	if wanted := "John 11:35"; on(m) != wanted {
		t.Errorf("Enter did not jump to the result, the reader is on %q, want %q", on(m), wanted)
	}
}

func TestSearchWithNoMatchesSaysNoMatches(t *testing.T) {
	m := activeReader(t, map[string]string{"description": "Reina-Valera"}, searchBooks(),
		[]modtest.Verse{{Book: 10, Chapter: 1, Verse: 1, Text: "En el principio"}})

	m, _ = send(t, m, press('/', "/", 0))
	m = typeKey(t, m, "banana")

	if got := render(t, m); !strings.Contains(got, "No matches") {
		t.Errorf("an empty result was not reported\n%s", got)
	}
}

func TestSearchRunsAgainstTheActiveTranslationOnly(t *testing.T) {
	dir := t.TempDir()
	info := map[string]string{"description": "Bible"}
	genesis := []modtest.Book{{Number: 10, ShortName: "Gen", LongName: "Genesis", Present: true}}
	modtest.BibleAll(t, dir, "RST.SQLite3", info, genesis,
		[]modtest.Verse{{Book: 10, Chapter: 1, Verse: 1, Text: "En el principio creó Dios"}})
	modtest.BibleAll(t, dir, "KJV.SQLite3", info, genesis,
		[]modtest.Verse{{Book: 10, Chapter: 1, Verse: 1, Text: "In the beginning God created"}})

	configDir := t.TempDir()
	if err := config.Save(configDir, config.Config{ModuleDir: dir, ActiveTranslation: "RST"}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	m := newReader(t, configDir, dir)
	m, _ = send(t, m, press('/', "/", 0))
	m = typeKey(t, m, "god")

	if got := render(t, m); !strings.Contains(got, "No matches") {
		t.Errorf("Search read a Translation other than the Active one\n%s", got)
	}
}

func TestSearchResultsScroll(t *testing.T) {
	verses := make([]modtest.Verse, 0, 20)
	for n := 1; n <= 20; n++ {
		verses = append(verses, modtest.Verse{Book: 10, Chapter: 1, Verse: n, Text: fmt.Sprintf("palabra %d", n)})
	}
	m := activeReader(t, map[string]string{"description": "Reina-Valera"},
		[]modtest.Book{{Number: 10, ShortName: "Gen", LongName: "Genesis", Present: true}}, verses)
	m = resized(t, m, 100, 16)

	m, _ = send(t, m, press('/', "/", 0))
	m = typeKey(t, m, "palabra")
	for i := 0; i < 10; i++ {
		m, _ = send(t, m, press(tea.KeyDown, "", 0))
	}

	got := render(t, m)
	if !strings.Contains(got, "Genesis 1:11") {
		t.Errorf("the cursor did not scroll to a later result\n%s", got)
	}
	if strings.Contains(got, "Genesis 1:1 ") {
		t.Errorf("the first result is still shown after scrolling past it\n%s", got)
	}
}
