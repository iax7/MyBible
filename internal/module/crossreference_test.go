package module

import (
	"reflect"
	"testing"

	"github.com/iax7/mybible/internal/modtest"
)

func TestDiscoverCrossReferencesFindsSetsAndSortsThem(t *testing.T) {
	dir := t.TempDir()
	modtest.CrossReferences(t, dir, "TSK-x.crossreferences.SQLite3",
		map[string]string{"description": "The Treasury of Scripture Knowledge", "language": "en"}, nil)
	modtest.CrossReferences(t, dir, "OBX.crossreferences.SQLite3",
		map[string]string{"description": "OpenBible", "requires_reverse_processing": "true"}, nil)
	modtest.Bible(t, dir, "RST.SQLite3", map[string]string{"description": "Synodal"})
	modtest.Commentaries(t, dir, "RST.commentaries.SQLite3", nil, nil)

	sets, warnings := DiscoverCrossReferences(dir)
	if len(warnings) != 0 {
		t.Fatalf("DiscoverCrossReferences reported %v", warnings)
	}
	if len(sets) != 2 {
		t.Fatalf("DiscoverCrossReferences found %d sets, want 2", len(sets))
	}
	if sets[0].Abbreviation != "OBX" || sets[0].Title != "OpenBible" || !sets[0].reverse {
		t.Errorf("sets[0] = %+v, want OBX asking for reverse processing", sets[0])
	}
	if sets[1].Abbreviation != "TSK-x" || sets[1].Title != "The Treasury of Scripture Knowledge" || sets[1].Language != "en" {
		t.Errorf("sets[1] = %+v, want the Treasury of Scripture Knowledge", sets[1])
	}
	if sets[0].reverse == sets[1].reverse {
		t.Error("the reverse flag was not read from info.requires_reverse_processing")
	}
}

func TestDiscoverCrossReferencesSkipsWhatItCannotRead(t *testing.T) {
	dir := t.TempDir()
	modtest.Create(t, dir, "BROKEN.crossreferences.SQLite3", `CREATE TABLE info (name TEXT, value TEXT)`)

	sets, warnings := DiscoverCrossReferences(dir)
	if len(sets) != 0 {
		t.Errorf("DiscoverCrossReferences returned %+v, want nothing", sets)
	}
	if len(warnings) != 1 {
		t.Errorf("DiscoverCrossReferences reported %v, want one warning", warnings)
	}
}

func TestCrossReferenceRowsReadsTheChapter(t *testing.T) {
	dir := t.TempDir()
	modtest.CrossReferences(t, dir, "TSK-x.crossreferences.SQLite3", nil, []modtest.CrossReference{
		{Book: 10, Chapter: 1, Verse: 2, ToBook: 500, ToChapter: 1, ToVerse: 3, ToVerseEnd: 5, Votes: 5},
		{Book: 10, Chapter: 2, Verse: 1, ToBook: 10, ToChapter: 1, ToVerse: 5, Votes: 2},
		{Book: 10, Chapter: 1, Verse: 4, ToBook: 9999, ToChapter: 1, ToVerse: 1},
	})
	sets, _ := DiscoverCrossReferences(dir)

	rows, err := sets[0].Rows(10, 1)
	if err != nil {
		t.Fatalf("Rows: %v", err)
	}
	want := []CrossReference{
		{Book: 10, Chapter: 1, Verse: 2, TargetBook: 500, TargetChapter: 1, TargetVerse: 3, TargetVerseEnd: 5, Votes: 5},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Errorf("Rows = %+v, want %+v", rows, want)
	}
}

// A set that asks for reverse processing is read both ways: the links that point
// at the Chapter become links that speak from it.
func TestCrossReferenceRowsReadBothDirections(t *testing.T) {
	dir := t.TempDir()
	modtest.CrossReferences(t, dir, "RST-x.crossreferences.SQLite3",
		map[string]string{"requires_reverse_processing": "true"},
		[]modtest.CrossReference{
			{Book: 10, Chapter: 1, Verse: 2, ToBook: 500, ToChapter: 1, ToVerse: 3, Votes: 5},
			{Book: 10, Chapter: 2, Verse: 1, ToBook: 10, ToChapter: 1, ToVerse: 5, Votes: 2},
			{Book: 10, Chapter: 3, Verse: 1, ToBook: 20, ToChapter: 2, ToVerse: 1},
		})
	sets, _ := DiscoverCrossReferences(dir)

	rows, err := sets[0].Rows(10, 1)
	if err != nil {
		t.Fatalf("Rows: %v", err)
	}
	want := []CrossReference{
		{Book: 10, Chapter: 1, Verse: 2, TargetBook: 500, TargetChapter: 1, TargetVerse: 3, Votes: 5},
		{Book: 10, Chapter: 1, Verse: 5, TargetBook: 10, TargetChapter: 2, TargetVerse: 1, Votes: 2},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Errorf("Rows = %+v, want %+v", rows, want)
	}
}

func TestCrossReferencesForCoversDeduplicatesAndSorts(t *testing.T) {
	rows := []CrossReference{
		{Book: 10, Chapter: 1, Verse: 1, VerseEnd: 3, TargetBook: 20, TargetChapter: 1, TargetVerse: 1, Votes: 2},
		{Book: 10, Chapter: 1, Verse: 2, TargetBook: 20, TargetChapter: 1, TargetVerse: 1, Votes: 7},
		{Book: 10, Chapter: 1, Verse: 2, TargetBook: 40, TargetChapter: 5, TargetVerse: 1, Votes: 7},
		{Book: 10, Chapter: 1, Verse: 2, TargetBook: 30, TargetChapter: 1, TargetVerse: 1, Votes: 7},
		{Book: 10, Chapter: 1, Verse: 5, TargetBook: 60, TargetChapter: 1, TargetVerse: 1, Votes: 9},
	}

	refs := CrossReferencesFor(rows, 2)
	if len(refs) != 3 {
		t.Fatalf("CrossReferencesFor returned %d links, want 3:\n%+v", len(refs), refs)
	}
	want := []targetKey{{20, 1, 1, 0}, {30, 1, 1, 0}, {40, 5, 1, 0}}
	for i, ref := range refs {
		if ref.key() != want[i] {
			t.Errorf("refs[%d] targets %v, want %v", i, ref.key(), want[i])
		}
	}
	if refs[0].Votes != 7 || refs[0].Verse != 2 {
		t.Errorf("the most voted link was not the one kept: %+v", refs[0])
	}
	if len(CrossReferencesFor(rows, 4)) != 0 {
		t.Error("a Verse with no link covering it got references")
	}
}

func TestCrossReferenceCoversItsSourceRange(t *testing.T) {
	link := CrossReference{Book: 10, Chapter: 1, Verse: 2, VerseEnd: 4}
	for verse, want := range map[int]bool{1: false, 2: true, 3: true, 4: true, 5: false} {
		if got := link.Covers(verse); got != want {
			t.Errorf("Covers(%d) = %v, want %v", verse, got, want)
		}
	}
	if !(CrossReference{Verse: 3}).Covers(3) || (CrossReference{Verse: 3}).Covers(4) {
		t.Error("a link without an end covered more than its own Verse")
	}
}

func TestCrossReferenceLabel(t *testing.T) {
	names := func(number int) string {
		if number == 500 {
			return "Juan"
		}
		return ""
	}
	for _, test := range []struct {
		link CrossReference
		want string
	}{
		{CrossReference{TargetBook: 500, TargetChapter: 3, TargetVerse: 16}, "Juan 3:16"},
		{CrossReference{TargetBook: 500, TargetChapter: 3, TargetVerse: 16, TargetVerseEnd: 18}, "Juan 3:16–18"},
		{CrossReference{Verse: 3, VerseEnd: 5, TargetBook: 10, TargetChapter: 1, TargetVerse: 2}, "(3–5) Genesis 1:2"},
	} {
		if got := test.link.Label(names); got != test.want {
			t.Errorf("Label = %q, want %q", got, test.want)
		}
	}
}
