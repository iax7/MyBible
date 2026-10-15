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
