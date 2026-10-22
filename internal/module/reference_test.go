package module

import "testing"

func TestParseReference(t *testing.T) {
	cases := []struct {
		text string
		want Reference
		ok   bool
	}{
		{"John 3:16", Reference{Book: Book{Number: 500, Name: "John"}, Chapter: 3, Verse: 16}, true},
		{"john 3:16", Reference{Book: Book{Number: 500, Name: "John"}, Chapter: 3, Verse: 16}, true},
		{"1 Samuel 3", Reference{Book: Book{Number: 90, Name: "1 Samuel"}, Chapter: 3, Verse: 1}, true},
		{"Song of Songs 1:2", Reference{Book: Book{Number: 260, Name: "Song of Songs"}, Chapter: 1, Verse: 2}, true},
		{"  Genesis   2:3  ", Reference{Book: Book{Number: 10, Name: "Genesis"}, Chapter: 2, Verse: 3}, true},
		{"John", Reference{Book: Book{Number: 500, Name: "John"}, Chapter: 1, Verse: 1}, true},
		{"1 Samuel", Reference{Book: Book{Number: 90, Name: "1 Samuel"}, Chapter: 1, Verse: 1}, true},
		{"Song of Songs", Reference{Book: Book{Number: 260, Name: "Song of Songs"}, Chapter: 1, Verse: 1}, true},
		{"John 3 16", Reference{Book: Book{Number: 500, Name: "John"}, Chapter: 3, Verse: 16}, true},
		{"John 3,16", Reference{Book: Book{Number: 500, Name: "John"}, Chapter: 3, Verse: 16}, true},
		{"John 3.16", Reference{Book: Book{Number: 500, Name: "John"}, Chapter: 3, Verse: 16}, true},
		{"John 3, 16", Reference{Book: Book{Number: 500, Name: "John"}, Chapter: 3, Verse: 16}, true},
		{"1 Samuel 3 16", Reference{Book: Book{Number: 90, Name: "1 Samuel"}, Chapter: 3, Verse: 16}, true},
		{"John 3 16 2", Reference{}, false},
		{"3:16", Reference{}, false},
		{"Genesis 2:3 4", Reference{}, false},
		{"Nonexistent 1:1", Reference{}, false},
		{"John 0:1", Reference{}, false},
		{"John 3:0", Reference{}, false},
		{"John 3:", Reference{}, false},
		{"John 3:16:2", Reference{}, false},
		{"", Reference{}, false},
	}
	for _, tc := range cases {
		got, ok := ParseReference(tc.text)
		if ok != tc.ok || got != tc.want {
			t.Errorf("ParseReference(%q) = %+v, %v; want %+v, %v", tc.text, got, ok, tc.want, tc.ok)
		}
	}
}

func TestReferenceStringRoundTrips(t *testing.T) {
	ref := Reference{Book: Book{Number: 500, Name: "John"}, Chapter: 3, Verse: 16}
	if got := ref.String(); got != "John 3:16" {
		t.Errorf("Reference.String() = %q, want %q", got, "John 3:16")
	}
	back, ok := ParseReference(ref.String())
	if !ok || back != ref {
		t.Errorf("ParseReference(String()) = %+v, %v; want %+v", back, ok, ref)
	}
}

// A Reference keeps the Canonical name for state files and shows the one the
// Module itself gives the Book, so a Spanish Translation reads "Génesis" without
// changing what ParseReference accepts.
func TestReferenceDisplayStringUsesTheModulesName(t *testing.T) {
	book := Book{Number: 10, Name: "Genesis", ShortName: "Gn", LongName: "Génesis"}
	ref := Reference{Book: book, Chapter: 1, Verse: 1}
	if got := ref.String(); got != "Genesis 1:1" {
		t.Errorf("Reference.String() = %q, want the stored form %q", got, "Genesis 1:1")
	}
	if got := ref.DisplayString(); got != "Génesis 1:1" {
		t.Errorf("Reference.DisplayString() = %q, want the shown form %q", got, "Génesis 1:1")
	}
	if back, ok := ParseReference(ref.DisplayString()); ok {
		t.Errorf("ParseReference(DisplayString()) = %+v, true; a displayed name is not a stored form", back)
	}
}
