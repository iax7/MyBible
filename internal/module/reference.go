package module

import (
	"fmt"
	"strconv"
	"strings"
)

// Reference is the address of a Verse: its Book, Chapter, and Verse number.
type Reference struct {
	Book    Book
	Chapter int
	Verse   int
}

// String renders a Reference the way the reader shows it and accepts it back.
func (r Reference) String() string {
	return fmt.Sprintf("%s %d:%d", r.Book.Name, r.Chapter, r.Verse)
}

// ParseReference parses "John 3:16", "John 3", or "John". The Book name is
// matched case-insensitively against the Canonical book list; a missing Chapter
// or Verse defaults to 1, and the Chapter and Verse are separated by ':', a
// comma, a period, or a space. Anything else is rejected. A Book named the way a
// Module itself names it is not known here: a caller holding a loaded
// Translation resolves that name first and passes the Canonical one.
func ParseReference(text string) (Reference, bool) {
	text = strings.TrimSpace(text)
	for _, b := range CanonicalBooks {
		if strings.EqualFold(b.Name, text) { // a Book on its own: its first Verse
			return Reference{Book: b, Chapter: 1, Verse: 1}, true
		}
	}
	book, location := SplitReference(text)
	if book == "" {
		return Reference{}, false
	}
	chapter, verse, ok := parseLocation(location)
	if !ok {
		return Reference{}, false
	}
	for _, b := range CanonicalBooks {
		if strings.EqualFold(b.Name, book) {
			return Reference{Book: b, Chapter: chapter, Verse: verse}, true
		}
	}
	return Reference{}, false
}

// separators are the characters that divide a Chapter from a Verse.
const separators = ":,."

// SplitReference separates the Book a Reference names from the location written
// after it. The location is the run of trailing words holding nothing but digits
// and separators, so "1 Samuel 3" keeps its Book name whole and "John 3 16" reads
// as John with a Chapter and a Verse. Text that names no Book is a location when
// it carries a separator or spans more than one word — "3:16", "3 16" — while a
// lone "3" is still a Book name, the numbered Books.
func SplitReference(text string) (book, location string) {
	text = strings.TrimSpace(text)
	words := strings.Fields(text)
	at := len(words)
	for at > 0 && isLocation(words[at-1]) {
		at--
	}
	if at > 0 {
		return strings.Join(words[:at], " "), strings.Join(words[at:], " ")
	}
	if !strings.ContainsAny(text, separators) && len(words) < 2 {
		return text, ""
	}
	return "", text
}

// isLocation reports whether a word could be part of the location of a
// Reference: nothing but digits and separators.
func isLocation(word string) bool {
	if word == "" {
		return false
	}
	for _, r := range word {
		if r != ':' && r != ',' && r != '.' && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}

// parseLocation parses a Chapter, or a Chapter and a Verse divided by a
// separator or a space: "3", "3:16", "3,16", "3.16", "3 16". Both are at least 1.
func parseLocation(location string) (chapter, verse int, ok bool) {
	verse = 1
	if i := strings.IndexAny(location, separators); i >= 0 {
		value, err := strconv.Atoi(strings.TrimSpace(location[i+1:]))
		if err != nil {
			return 0, 0, false
		}
		verse, location = value, location[:i]
	} else if words := strings.Fields(location); len(words) > 1 {
		if len(words) != 2 {
			return 0, 0, false
		}
		value, err := strconv.Atoi(words[1])
		if err != nil {
			return 0, 0, false
		}
		verse, location = value, words[0]
	}
	value, err := strconv.Atoi(strings.TrimSpace(location))
	if err != nil {
		return 0, 0, false
	}
	if value < 1 || verse < 1 {
		return 0, 0, false
	}
	return value, verse, true
}
