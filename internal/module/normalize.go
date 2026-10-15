package module

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Normalize lowercases text and strips accents, so a query ignores letter case
// and diacritics (jesus finds Jesús, and a Module's Génesis answers to genesis).
func Normalize(text string) string {
	decomposed := norm.NFD.String(strings.ToLower(text))
	var b strings.Builder
	b.Grow(len(decomposed))
	for _, r := range decomposed {
		if unicode.Is(unicode.Mn, r) { // a combining mark: the accent itself
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
