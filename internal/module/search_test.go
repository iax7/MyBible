package module

import (
	"strings"
	"testing"

	"github.com/iax7/mybible/internal/modtest"
)

// searchFixture is a two-Book Bible with accented text, so Search can be shown
// to ignore letter case and accents.
func searchFixture() (map[string]string, []modtest.Book, []modtest.Verse) {
	info := map[string]string{"description": "Reina-Valera"}
	books := []modtest.Book{
		{Number: 10, ShortName: "Gen", LongName: "Genesis", Present: true},
		{Number: 500, ShortName: "Jhn", LongName: "John", Present: true},
	}
	verses := []modtest.Verse{
		{Book: 10, Chapter: 1, Verse: 1, Text: "En el principio creó Dios los cielos"},
		{Book: 500, Chapter: 3, Verse: 16, Text: "Porque de tal manera amó Dios al mundo"},
		{Book: 500, Chapter: 11, Verse: 35, Text: "<J>Jesús</J> lloró."},
	}
	return info, books, verses
}

// address is the part of a Reference Search results are asserted on: the Book
// is narrowed to its Canonical number, since a loaded Book also carries its
// Chapter count.
func address(r Reference) [3]int { return [3]int{r.Book.Number, r.Chapter, r.Verse} }

func TestSearchIgnoresLetterCaseAndAccents(t *testing.T) {
	info, books, verses := searchFixture()
	text := loadFixture(t, "RST.SQLite3", info, books, verses, true)

	want := address(Reference{Book: Book{Number: 500, Name: "John"}, Chapter: 11, Verse: 35})
	for _, query := range []string{"jesus", "JESUS", "Jesús", "JESÚS"} {
		got := text.Search(query)
		if len(got) != 1 || address(got[0].Reference) != want {
			t.Errorf("Search(%q) = %+v, want the single hit %v", query, got, want)
		}
	}
}

func TestSearchReturnsResultsInCanonicalOrder(t *testing.T) {
	info, books, verses := searchFixture()
	text := loadFixture(t, "RST.SQLite3", info, books, verses, true)

	got := text.Search("dios")
	want := [][3]int{{10, 1, 1}, {500, 3, 16}}
	if len(got) != len(want) {
		t.Fatalf("Search(\"dios\") returned %d hits, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if address(got[i].Reference) != want[i] {
			t.Errorf("hit %d = %v, want %v", i, address(got[i].Reference), want[i])
		}
	}
}

func TestSearchSnippetIsDisplayText(t *testing.T) {
	info, books, verses := searchFixture()
	text := loadFixture(t, "RST.SQLite3", info, books, verses, true)

	got := text.Search("jesus")
	if len(got) != 1 {
		t.Fatalf("Search(\"jesus\") returned %d hits, want 1", len(got))
	}
	if strings.Contains(got[0].Snippet, "<J>") {
		t.Errorf("snippet kept markup: %q", got[0].Snippet)
	}
	if !strings.Contains(got[0].Snippet, "Jesús lloró") {
		t.Errorf("snippet = %q, want it to show the matched text", got[0].Snippet)
	}
}

func TestSearchSnippetKeepsTheMatchVisibleInALongVerse(t *testing.T) {
	long := strings.Repeat("palabra ", 40) + "Jesús lloró"
	text := loadFixture(t, "RST.SQLite3", map[string]string{"description": "Reina-Valera"},
		[]modtest.Book{{Number: 500, ShortName: "Jhn", LongName: "John", Present: true}},
		[]modtest.Verse{{Book: 500, Chapter: 11, Verse: 35, Text: long}}, true)

	got := text.Search("jesus")
	if len(got) != 1 {
		t.Fatalf("Search(\"jesus\") returned %d hits, want 1", len(got))
	}
	if !strings.Contains(got[0].Snippet, "Jesús") {
		t.Errorf("snippet = %q, want the match kept visible", got[0].Snippet)
	}
}

func TestSearchWithNoMatchesReturnsNothing(t *testing.T) {
	info, books, verses := searchFixture()
	text := loadFixture(t, "RST.SQLite3", info, books, verses, true)

	if got := text.Search("banana"); len(got) != 0 {
		t.Errorf("Search(\"banana\") = %+v, want no hits", got)
	}
	if got := text.Search(""); len(got) != 0 {
		t.Errorf("Search(\"\") = %+v, want no hits", got)
	}
}
