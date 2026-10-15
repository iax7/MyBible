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

// ParseReference parses "John 3:16", "John 3", or "1 Samuel 3:16". The Book name
// is matched case-insensitively against the Canonical book list; a missing Verse
// defaults to 1. Anything else is rejected.
func ParseReference(text string) (Reference, bool) {
	book, location, ok := splitReference(strings.TrimSpace(text))
	if !ok {
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

// splitReference separates the trailing Chapter:Verse from the Book name at the
// last space, since Book names contain spaces and leading digits of their own.
func splitReference(text string) (book, location string, ok bool) {
	i := strings.LastIndexByte(text, ' ')
	if i <= 0 || i == len(text)-1 {
		return "", "", false
	}
	return strings.Join(strings.Fields(text[:i]), " "), text[i+1:], true
}

// parseLocation parses "3" or "3:16" into a Chapter and a Verse, both at least 1.
func parseLocation(location string) (chapter, verse int, ok bool) {
	verse = 1
	if i := strings.IndexByte(location, ':'); i >= 0 {
		value, err := strconv.Atoi(location[i+1:])
		if err != nil {
			return 0, 0, false
		}
		verse = value
		location = location[:i]
	}
	value, err := strconv.Atoi(location)
	if err != nil {
		return 0, 0, false
	}
	if value < 1 || verse < 1 {
		return 0, 0, false
	}
	return value, verse, true
}
