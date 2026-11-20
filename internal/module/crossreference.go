package module

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// TypeCrossReferences is the Module type holding a set of Cross-references. It
// is never surfaced as a Translation; the reader reads it beside one.
const TypeCrossReferences Type = "crossreferences"

const tableCrossReferences = "cross_references"

// CrossReferenceSet is a Module of type crossreferences.
type CrossReferenceSet struct {
	Abbreviation string // the file name stem, which exists nowhere inside the Module
	Title        string // the human title, from info.description
	Language     string // the two-letter language, from info.language
	path         string
	reverse      bool // the Module asks for its links to be read in both directions
}

// CrossReference is one link of a Cross-reference set: the range of Verses it
// speaks from, the range of Verses it points to, and how many witnesses count it.
type CrossReference struct {
	Book, Chapter, Verse, VerseEnd int
	TargetBook                     int
	TargetChapter, TargetVerse     int
	TargetVerseEnd                 int
	Votes                          int
}

// Covers reports whether the link speaks from a Verse of its Chapter. A link
// written for one Verse has no end and covers only that Verse.
func (r CrossReference) Covers(verse int) bool {
	if r.Verse < 1 || verse < r.Verse {
		return false
	}
	if r.VerseEnd < 1 {
		return verse == r.Verse
	}
	return verse <= r.VerseEnd
}

// Label renders the link the way the reader shows it: the name the Active
// translation gives the target Book, its Chapter and Verse range, and, when the
// link speaks from a range, that range in parentheses. bookName names a Book the
// way the Translation does; nil keeps the Canonical names.
func (r CrossReference) Label(bookName func(int) string) string {
	name := ""
	if bookName != nil {
		name = bookName(r.TargetBook)
	}
	if name == "" {
		book, ok := canonicalBook(r.TargetBook)
		if !ok {
			return ""
		}
		name = book.Name
	}
	label := name + " " + strconv.Itoa(r.TargetChapter) + ":" + strconv.Itoa(r.TargetVerse)
	if r.TargetVerseEnd > r.TargetVerse {
		label += "–" + strconv.Itoa(r.TargetVerseEnd)
	}
	if r.VerseEnd > r.Verse {
		label = fmt.Sprintf("(%d–%d) %s", r.Verse, r.VerseEnd, label)
	}
	return label
}

// targetKey identifies the Verses a link points to, which is what makes two
// links from different sets the same Cross-reference.
type targetKey [4]int

// key is the link's target as a comparable value.
func (r CrossReference) key() targetKey {
	return targetKey{r.TargetBook, r.TargetChapter, r.TargetVerse, r.TargetVerseEnd}
}

// CrossReferencesFor returns the links that speak to one Verse, each target once:
// the most voted of the links that share a target wins, and the rest are sorted
// by votes descending, then in Canonical order.
func CrossReferencesFor(rows []CrossReference, verse int) []CrossReference {
	refs := make([]CrossReference, 0, len(rows))
	seen := make(map[targetKey]int, len(rows))
	for _, row := range rows {
		if !row.Covers(verse) {
			continue
		}
		if at, ok := seen[row.key()]; ok {
			if row.Votes > refs[at].Votes {
				refs[at] = row
			}
			continue
		}
		seen[row.key()] = len(refs)
		refs = append(refs, row)
	}
	sort.SliceStable(refs, func(i, j int) bool { return refs[i].before(refs[j]) })
	return refs
}

// before orders Cross-references by relevance: the most voted first, and a tie
// by the Canonical order of their targets.
func (r CrossReference) before(other CrossReference) bool {
	if r.Votes != other.Votes {
		return r.Votes > other.Votes
	}
	if r.TargetBook != other.TargetBook {
		return r.TargetBook < other.TargetBook
	}
	if r.TargetChapter != other.TargetChapter {
		return r.TargetChapter < other.TargetChapter
	}
	if r.TargetVerse != other.TargetVerse {
		return r.TargetVerse < other.TargetVerse
	}
	return r.TargetVerseEnd < other.TargetVerseEnd
}

// DiscoverCrossReferences scans dir for Cross-reference sets, sorted by
// abbreviation. A Module that cannot be read is skipped and reported in
// warnings; discovery never fails.
func DiscoverCrossReferences(dir string) (sets []CrossReferenceSet, warnings []string) {
	paths, err := filepath.Glob(filepath.Join(dir, "*"+extension))
	if err != nil {
		return nil, []string{fmt.Sprintf("cannot scan the Module directory %s: %v", dir, err)}
	}

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		name := filepath.Base(path)
		if TypeOf(name) != TypeCrossReferences {
			continue
		}
		set, err := readCrossReferenceSet(path)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", name, err))
			continue
		}
		sets = append(sets, set)
	}

	sort.Slice(sets, func(i, j int) bool {
		return sets[i].Abbreviation < sets[j].Abbreviation
	})
	return sets, warnings
}

// readCrossReferenceSet opens one Cross-reference set read-only and reads its
// info table.
func readCrossReferenceSet(path string) (CrossReferenceSet, error) {
	db, err := openReadOnly(path)
	if err != nil {
		return CrossReferenceSet{}, err
	}
	defer func() { _ = db.Close() }()

	tables, err := moduleTables(db)
	if err != nil {
		return CrossReferenceSet{}, err
	}
	if !tables[tableCrossReferences] {
		return CrossReferenceSet{}, fmt.Errorf("missing the required %s table", tableCrossReferences)
	}

	info := make(map[string]string)
	if tables[tableInfo] {
		info, err = infoTable(db)
		if err != nil {
			return CrossReferenceSet{}, err
		}
	}
	stem := strings.TrimSuffix(filepath.Base(path), extension)
	return CrossReferenceSet{
		Abbreviation: strings.TrimSuffix(stem, "."+string(TypeCrossReferences)),
		Title:        info["description"],
		Language:     info["language"],
		path:         path,
		reverse:      info["requires_reverse_processing"] == "true",
	}, nil
}

// Rows reads one Chapter's Cross-references on demand, in range order. A set
// that asks for reverse processing is read in both directions: the links that
// point at the Chapter are returned as links that speak from it.
func (s CrossReferenceSet) Rows(book, chapter int) ([]CrossReference, error) {
	if s.path == "" || chapter < 1 {
		return nil, nil
	}
	db, err := openReadOnly(s.path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Close() }()

	numbers := bookNumbers(book)
	placeholders := make([]string, 0, len(numbers))
	args := make([]any, 0, 2*len(numbers)+2)
	for _, number := range numbers {
		placeholders = append(placeholders, "?")
		args = append(args, number)
	}
	in := strings.Join(placeholders, ",")
	where := "book IN (" + in + ") AND chapter = ?"
	args = append(args, chapter)
	if s.reverse {
		where = "(" + where + ") OR (book_to IN (" + in + ") AND chapter_to = ?)"
		for _, number := range numbers {
			args = append(args, number)
		}
		args = append(args, chapter)
	}

	rows, err := db.Query(`SELECT book, chapter, verse, COALESCE(verse_end, 0),
			book_to, chapter_to, verse_to_start, COALESCE(verse_to_end, 0), COALESCE(votes, 0)
		FROM `+tableCrossReferences+`
		WHERE `+where+`
		ORDER BY book, chapter, verse, COALESCE(verse_end, 0)`, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var references []CrossReference
	for rows.Next() {
		var ref CrossReference
		var sourceBook, sourceChapter, targetBook, targetChapter int
		if err := rows.Scan(&sourceBook, &sourceChapter, &ref.Verse, &ref.VerseEnd,
			&targetBook, &targetChapter, &ref.TargetVerse, &ref.TargetVerseEnd, &ref.Votes); err != nil {
			return nil, err
		}
		if targetChapter < 1 || ref.TargetVerse < 1 {
			continue
		}
		source, ok := canonicalBook(sourceBook)
		if !ok {
			continue
		}
		target, ok := canonicalBook(targetBook)
		if !ok {
			continue
		}
		ref.TargetBook, ref.TargetChapter = target.Number, targetChapter
		if sourceChapter == chapter {
			ref.Book, ref.Chapter = source.Number, sourceChapter
			references = append(references, ref)
		}
		if !s.reverse || target.Number != book || targetChapter != chapter {
			continue
		}
		// The link points at this Chapter: read it backwards, from the target
		// it names to the Verse it was written from.
		references = append(references, CrossReference{
			Book:           target.Number,
			Chapter:        targetChapter,
			Verse:          ref.TargetVerse,
			VerseEnd:       ref.TargetVerseEnd,
			TargetBook:     source.Number,
			TargetChapter:  sourceChapter,
			TargetVerse:    ref.Verse,
			TargetVerseEnd: ref.VerseEnd,
			Votes:          ref.Votes,
		})
	}
	return references, rows.Err()
}
