package module

import "strings"

// Result is one Search hit: the Reference of a matching Verse and a snippet of
// its text as the reader displays it.
type Result struct {
	Reference Reference
	Snippet   string
}

// snippetRunes is how much Verse text a result shows around the match.
const snippetRunes = 60

// Search returns the Verses of the Active translation whose text contains the
// query, ignoring letter case and accents, in Canonical order. An empty query
// matches nothing.
func (t *Text) Search(query string) []Result {
	query = searchable(query)
	if query == "" {
		return nil
	}

	var results []Result
	for _, book := range t.books { // books are already in Canonical order
		for chapter := 1; chapter < len(book.chapters); chapter++ {
			for i := range book.chapters[chapter] {
				verse := &book.chapters[chapter][i]
				if !strings.Contains(verse.normalized, query) {
					continue
				}
				results = append(results, Result{
					Reference: Reference{Book: book.Book, Chapter: chapter, Verse: verse.Number},
					Snippet:   snippet(verse.Text, query),
				})
			}
		}
	}
	return results
}

// searchable is the form Search matches against: display text normalized and
// with runs of whitespace collapsed to single spaces.
func searchable(text string) string {
	return normalize(collapse(displayText(text)))
}

// displayText is a Verse's text as the reader shows it, with inline markup
// rendered away.
func displayText(text string) string {
	var b strings.Builder
	for _, run := range ParseMarkup(text) {
		b.WriteString(run.Text)
	}
	return b.String()
}

// collapse turns any run of whitespace (including the breaks ParseMarkup
// emits) into a single space, so a result stays on one line.
func collapse(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

// snippet is a window of a Verse's display text around where the query matches,
// so a long Verse stays readable and the match stays visible.
func snippet(text, query string) string {
	runes := []rune(collapse(displayText(text)))
	if len(runes) <= snippetRunes {
		return string(runes)
	}

	start := matchIndex(runes, query) - snippetRunes/3
	if start < 0 {
		start = 0
	}
	end := start + snippetRunes
	if end > len(runes) {
		end = len(runes)
		start = end - snippetRunes
	}
	out := string(runes[start:end])
	if start > 0 {
		out = "…" + out
	}
	if end < len(runes) {
		out += "…"
	}
	return out
}

// matchIndex is the position in display at which query matches once both are
// normalized, or 0 when the match cannot be mapped back.
func matchIndex(display []rune, query string) int {
	want := []rune(query)
	if len(want) == 0 {
		return 0
	}

	normalized := make([]rune, 0, len(display))
	positions := make([]int, 0, len(display))
	for i, r := range display {
		for _, n := range normalize(string(r)) {
			normalized = append(normalized, n)
			positions = append(positions, i)
		}
	}
	for i := 0; i+len(want) <= len(normalized); i++ {
		if string(normalized[i:i+len(want)]) == query {
			return positions[i]
		}
	}
	return 0
}
