package module

import "testing"

// TestEveryCanonicalBookHasAClassification guards classificationNumbers: a Book
// left out of it would silently read as NonCanonical.
func TestEveryCanonicalBookHasAClassification(t *testing.T) {
	for _, b := range CanonicalBooks {
		if _, ok := classificationByNumber[b.Number]; !ok {
			t.Errorf("Book %d (%s) is in no Classification", b.Number, b.Name)
		}
	}
}

func TestBookNamed(t *testing.T) {
	genesis := Book{Number: 10, Name: "Genesis", ShortName: "Gn", LongName: "Génesis"}
	cases := []struct {
		text string
		want bool
	}{
		{"Genesis", true},  // the reader's own name
		{"genesis", true},  // letter case is ignored
		{"Gn", true},       // the Module's short name
		{"Génesis", true},  // the Module's long name
		{"GENESIS", true},  // ...
		{"genesis ", true}, // the input is trimmed
		{"génesis", true},  // accents are ignored on both sides
		{"Gen", false},     // a prefix of a name is not a name
		{"John", false},
		{"", false},
	}
	for _, c := range cases {
		if got := genesis.Named(c.text); got != c.want {
			t.Errorf("Book.Named(%q) = %v, want %v", c.text, got, c.want)
		}
	}
}

func TestBookNamesFallsBackToTheCanonicalName(t *testing.T) {
	canonical := Book{Number: 500, Name: "John"} // a Canonical Book carries no Module names
	if got := canonical.Names(); len(got) != 1 || got[0] != "John" {
		t.Errorf("Canonical Book.Names() = %v, want just the reader's own name", got)
	}
	named := Book{Number: 10, Name: "Genesis", ShortName: "Gen", LongName: "Genesis"}
	if got := named.Names(); len(got) != 2 || got[1] != "Gen" {
		t.Errorf("Book.Names() = %v, want the two names it has, with no repetition", got)
	}
}

func TestClassificationOf(t *testing.T) {
	cases := []struct {
		number int
		want   Classification
	}{
		{10, Pentateuch},
		{50, Pentateuch},
		{60, HistoricalBooks},
		{230, PoeticBooks},
		{290, MajorProphets},
		{350, MinorProphets},
		{500, Gospels},
		{510, Acts},
		{520, PaulineEpistles},
		{660, GeneralEpistles},
		{730, Revelation},
		{165, HistoricalBooks},
		{280, PoeticBooks},
		{780, GeneralEpistles},
		{192, NonCanonical},
		{790, PoeticBooks},  // Prayer of Manasseh under its older book_number
		{200, NonCanonical}, // 1 Maccabees under its older book_number
		{9999, NonCanonical},
	}
	for _, c := range cases {
		if got := ClassificationOf(c.number); got != c.want {
			t.Errorf("ClassificationOf(%d) = %v, want %v", c.number, got, c.want)
		}
	}
}
