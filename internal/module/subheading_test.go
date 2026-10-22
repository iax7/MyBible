package module

import (
	"reflect"
	"testing"
)

func TestSubheadingReferenceLabels(t *testing.T) {
	cases := []struct {
		name  string
		raw   string
		text  string
		links []string
		runs  []Run // set when the Runs around a link matter
	}{
		{
			name:  "single verse",
			raw:   "<x>10 3:1</x>",
			text:  "Genesis 3:1",
			links: []string{"Genesis 3:1"},
		},
		{
			name:  "same chapter range",
			raw:   "<x>10 3:1-24</x>",
			text:  "Genesis 3:1–24",
			links: []string{"Genesis 3:1–24"},
		},
		{
			name:  "cross chapter range",
			raw:   "<x>500 1:3-2:10</x>",
			text:  "John 1:3–2:10",
			links: []string{"John 1:3–2:10"},
		},
		{
			name:  "comma list",
			raw:   "<x>10 1:1,3,5</x>",
			text:  "Genesis 1:1,3,5",
			links: []string{"Genesis 1:1,3,5"},
		},
		{
			name:  "chapter",
			raw:   "<x>10 1</x>",
			text:  "Genesis 1",
			links: []string{"Genesis 1"},
		},
		{
			name:  "chapter range",
			raw:   "<x>10 1-3</x>",
			text:  "Genesis 1–3",
			links: []string{"Genesis 1–3"},
		},
		{
			name:  "chapter range without a colon is still a chapter range",
			raw:   "<x>720 3-13</x>",
			text:  "Jude 3–13",
			links: []string{"Jude 3–13"},
		},
		{
			name:  "optional B prefix",
			raw:   "<x>B:10 3:1-24</x>",
			text:  "Genesis 3:1–24",
			links: []string{"Genesis 3:1–24"},
		},
		{
			name:  "multiple references",
			raw:   "(<x>20 34:18-26</x>; <x>50 16:1-17</x>)",
			text:  "(Exodus 34:18–26; Deuteronomy 16:1–17)",
			links: []string{"Exodus 34:18–26", "Deuteronomy 16:1–17"},
			runs: []Run{
				{Text: "("},
				{Text: "Exodus 34:18–26", Link: true},
				{Text: "; "},
				{Text: "Deuteronomy 16:1–17", Link: true},
				{Text: ")"},
			},
		},
		{
			name:  "surrounding words are kept",
			raw:   "The <x>10 1:1</x> fall",
			text:  "The Genesis 1:1 fall",
			links: []string{"Genesis 1:1"},
			runs: []Run{
				{Text: "The "},
				{Text: "Genesis 1:1", Link: true},
				{Text: " fall"},
			},
		},
		{
			name: "unknown book is hidden",
			raw:  "The <x>999 1:1</x>fall",
			text: "The fall",
		},
		{
			name: "a non-numeric target is hidden",
			raw:  "The <x>Gen 1:1</x>fall",
			text: "The fall",
		},
		{
			name: "a number below one is hidden",
			raw:  "<x>10 0:1</x>",
			text: "",
		},
		{
			name: "an unparsable target is hidden",
			raw:  "<x>10 one</x>",
			text: "",
		},
		{
			name: "an unclosed tag hides the rest",
			raw:  "before <x>10 1:1",
			text: "before",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runs := headingRuns(tc.raw, nil)
			if got := headingText(runs); got != tc.text {
				t.Errorf("headingRuns(%q) text = %q, want %q", tc.raw, got, tc.text)
			}
			var links []string
			for _, run := range runs {
				if run.Link {
					links = append(links, run.Text)
				}
			}
			if !reflect.DeepEqual(links, tc.links) {
				t.Errorf("headingRuns(%q) links = %q, want %q", tc.raw, links, tc.links)
			}
			if tc.runs != nil && !reflect.DeepEqual(runs, tc.runs) {
				t.Errorf("headingRuns(%q) = %+v, want %+v", tc.raw, runs, tc.runs)
			}
		})
	}
}

func TestHeadingRunsNamesTheBookTheWayTheTranslationDoes(t *testing.T) {
	bookName := func(number int) string {
		if number == 20 {
			return "Éxodo"
		}
		return ""
	}
	runs := headingRuns("La salida <x>b:20 3:1</x>", bookName)

	var links []string
	for _, run := range runs {
		if run.Link {
			links = append(links, run.Text)
		}
	}
	if want := []string{"Éxodo 3:1"}; !reflect.DeepEqual(links, want) {
		t.Errorf("headingRuns() links = %q, want %q", links, want)
	}
}
