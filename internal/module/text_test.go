package module

import (
	"reflect"
	"sort"
	"testing"

	"github.com/iax7/mybible/internal/modtest"
)

// loadFixture builds a Bible Module, discovers it, and loads it into memory.
func loadFixture(t *testing.T, name string, info map[string]string, books []modtest.Book, verses []modtest.Verse, all bool) *Text {
	t.Helper()

	dir := t.TempDir()
	if all {
		modtest.BibleAll(t, dir, name, info, books, verses)
	} else {
		modtest.BibleBooks(t, dir, name, info, books, verses)
	}

	translations, warnings := Discover(dir)
	if len(warnings) != 0 {
		t.Fatalf("Discover warned: %v", warnings)
	}
	if len(translations) != 1 {
		t.Fatalf("Discover found %d Translations, want 1", len(translations))
	}
	text, err := translations[0].Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return text
}

func TestCanonicalBookListUsesMyBibleNumbers(t *testing.T) {
	byName := make(map[string]int, len(CanonicalBooks))
	for _, b := range CanonicalBooks {
		byName[b.Name] = b.Number
	}

	cases := map[string]int{
		"Genesis":            10,
		"Revelation":         730,
		"Prayer of Manasseh": 145,
		"1 Maccabees":        462,
		"2 Esdras":           165,
		"Song of Songs":      260,
	}
	for name, want := range cases {
		if got := byName[name]; got != want {
			t.Errorf("Canonical number of %q = %d, want %d", name, got, want)
		}
	}

	if len(CanonicalBooks) != 84 {
		t.Errorf("Canonical book list has %d Books, want 84 (66 plus deuterocanonicals)", len(CanonicalBooks))
	}
	if !sort.SliceIsSorted(CanonicalBooks, func(i, j int) bool {
		return CanonicalBooks[i].Number < CanonicalBooks[j].Number
	}) {
		t.Error("Canonical book list is not ordered by book_number")
	}
}

func TestLoadMapsModuleBooksIntoTheCanonicalList(t *testing.T) {
	text := loadFixture(t, "RST.SQLite3", map[string]string{"description": "Reina-Valera"},
		[]modtest.Book{
			{Number: 10, ShortName: "Gen", LongName: "Genesis", Present: true},
			{Number: 790, ShortName: "Man", LongName: "Manasseh", Present: true}, // legacy number
			{Number: 200, ShortName: "1Ma", LongName: "1 Maccabees", Present: true},
			{Number: 730, ShortName: "Rev", LongName: "Revelation", Present: true},
			{Number: 220, ShortName: "Job", LongName: "Job", Present: false}, // absent
			{Number: 999, ShortName: "X", LongName: "Unknown", Present: true},
		},
		[]modtest.Verse{
			{Book: 10, Chapter: 1, Verse: 1, Text: "In the beginning"},
			{Book: 10, Chapter: 1, Verse: 2, Text: "And the earth"},
			{Book: 10, Chapter: 2, Verse: 1, Text: "Thus the heavens"},
			{Book: 790, Chapter: 1, Verse: 1, Text: "O Lord"},
			{Book: 200, Chapter: 1, Verse: 1, Text: "And it came to pass"},
			{Book: 730, Chapter: 1, Verse: 1, Text: "The revelation"},
			{Book: 220, Chapter: 1, Verse: 1, Text: "this Book is absent"},
			{Book: 999, Chapter: 1, Verse: 1, Text: "this Book is unknown"},
		}, true)

	numbers := make([]int, 0)
	for _, b := range text.Books() {
		numbers = append(numbers, b.Number)
	}
	if want := []int{10, 145, 462, 730}; !reflect.DeepEqual(numbers, want) {
		t.Fatalf("loaded Book numbers = %v, want %v", numbers, want)
	}

	genesis, ok := text.Book(10)
	if !ok || genesis.Name != "Genesis" {
		t.Errorf("Book(10) = %+v, %v; want the canonical Genesis", genesis, ok)
	}
	if genesis.Chapters != 2 {
		t.Errorf("Genesis has %d Chapters, want 2", genesis.Chapters)
	}

	manasseh, ok := text.Book(145)
	if !ok || manasseh.Name != "Prayer of Manasseh" {
		t.Errorf("legacy book 790 did not map onto canonical 145: %+v, %v", manasseh, ok)
	}
	if verses := text.Chapter(145, 1); len(verses) != 1 || verses[0].Text != "O Lord" {
		t.Errorf("Chapter(145, 1) = %+v, want the legacy-numbered Book's Verse", verses)
	}

	if _, ok := text.Book(462); !ok {
		t.Error("legacy book 200 did not map onto canonical 1 Maccabees (462)")
	}
	if _, ok := text.Book(220); ok {
		t.Error("a Book marked absent (is_present = false) was loaded")
	}
	if _, ok := text.Book(999); ok {
		t.Error("an unknown book_number was loaded")
	}

	verses := text.Chapter(10, 1)
	if len(verses) != 2 || verses[0].Number != 1 || verses[1].Number != 2 {
		t.Fatalf("Chapter(10, 1) = %+v, want Verses 1 and 2", verses)
	}
	if verses[0].Text != "In the beginning" {
		t.Errorf("Verse 1 text = %q, want %q", verses[0].Text, "In the beginning")
	}
	if got := len(text.Chapter(10, 3)); got != 0 {
		t.Errorf("an absent Chapter returned %d Verses, want 0", got)
	}
}

func TestLoadTreatsEveryListedBookAsPresentWithoutIsPresent(t *testing.T) {
	text := loadFixture(t, "RST.SQLite3", map[string]string{"description": "Reina-Valera"},
		[]modtest.Book{
			{Number: 10, ShortName: "Gen", LongName: "Genesis"},
			{Number: 730, ShortName: "Rev", LongName: "Revelation"},
		},
		[]modtest.Verse{
			{Book: 10, Chapter: 1, Verse: 1, Text: "In the beginning"},
			{Book: 730, Chapter: 1, Verse: 1, Text: "The revelation"},
		}, false)

	if got := len(text.Books()); got != 2 {
		t.Fatalf("loaded %d Books from a book table without is_present, want 2", got)
	}
}

func TestLoadKeepsAnEmptyVerseRow(t *testing.T) {
	text := loadFixture(t, "RST.SQLite3", map[string]string{"description": "Reina-Valera"},
		[]modtest.Book{{Number: 10, ShortName: "Gen", LongName: "Genesis", Present: true}},
		[]modtest.Verse{
			{Book: 10, Chapter: 1, Verse: 1, Text: "In the beginning"},
			{Book: 10, Chapter: 1, Verse: 2, Text: ""},
		}, true)

	verses := text.Chapter(10, 1)
	if len(verses) != 2 {
		t.Fatalf("Chapter(10, 1) has %d Verses, want the empty row kept", len(verses))
	}
	if verses[1].Number != 2 || verses[1].Text != "" {
		t.Errorf("the empty Verse row = %+v, want Verse 2 with no text", verses[1])
	}
	if runs := ParseMarkup(verses[1].Text); len(runs) != 0 {
		t.Errorf("ParseMarkup of an empty Verse = %+v, want no Runs", runs)
	}
}

func TestParseMarkup(t *testing.T) {
	cases := []struct {
		name string
		text string
		want []Run
	}{
		{"plain", "plain text", []Run{{Text: "plain text"}}},
		{"italics", "the <i>beginning</i>", []Run{{Text: "the "}, {Text: "beginning", Italics: true}}},
		{"jesus", "and <J>he said</J>", []Run{{Text: "and "}, {Text: "he said", Red: true}}},
		{"nested", "<J>a<i>b</i>c</J>", []Run{{Text: "a", Red: true}, {Text: "b", Italics: true, Red: true}, {Text: "c", Red: true}}},
		{"line break", "a<br/>b", []Run{{Text: "a"}, {Text: "\n"}, {Text: "b"}}},
		{"paragraph break", "a<pb/>b", []Run{{Text: "a"}, {Text: "\n\n"}, {Text: "b"}}},
		{"strongs stripped", "a<S>7225</S>b", []Run{{Text: "a"}, {Text: "b"}}},
		{"morphology stripped", "a<m>HV</m>b", []Run{{Text: "a"}, {Text: "b"}}},
		{"footnote stripped", "a<f>note</f>b", []Run{{Text: "a"}, {Text: "b"}}},
		{"subheading stripped", "a<h>heading</h>b", []Run{{Text: "a"}, {Text: "b"}}},
		{"unknown tag stripped", "a<x>weird</x>b", []Run{{Text: "a"}, {Text: "weird"}, {Text: "b"}}},
		{"empty", "", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ParseMarkup(tc.text); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ParseMarkup(%q) = %+v, want %+v", tc.text, got, tc.want)
			}
		})
	}
}
