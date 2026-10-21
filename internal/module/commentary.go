package module

import (
	"database/sql"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// TypeCommentaries is the Module type holding one edition's Commentaries. It is
// never surfaced as a Translation; the reader reads it beside one.
const TypeCommentaries Type = "commentaries"

const tableCommentaries = "commentaries"

// Commentary is a Module of type commentaries.
type Commentary struct {
	Abbreviation string // the file name stem, which exists nowhere inside the Module
	Title        string // the human title, from info.description
	Language     string // the two-letter language, from info.language
	path         string
}

// CommentaryRow is one entry of a Commentary: the range of Verses it speaks to
// and its text, with the Module's markup still in it. A zero ChapterTo or
// VerseTo means the Module wrote no end; a zero VerseFrom means the entry is a
// summary of the whole Chapter.
type CommentaryRow struct {
	ChapterFrom int
	VerseFrom   int
	ChapterTo   int
	VerseTo     int
	Text        string
}

// ChapterLevel reports whether the entry is a summary of a Chapter rather than
// of particular Verses.
func (r CommentaryRow) ChapterLevel() bool { return r.VerseFrom < 1 }

// Covers reports whether the entry speaks to a Verse of a Chapter. A Chapter-
// level entry covers every Verse of its Chapter; an entry with no end covers its
// own Verse or Chapter only.
func (r CommentaryRow) Covers(chapter, verse int) bool {
	if r.ChapterFrom < 1 || chapter < 1 || verse < 1 {
		return false
	}
	fromChapter, fromVerse := r.ChapterFrom, r.VerseFrom
	if fromVerse < 1 {
		fromVerse = 1
	}
	toChapter, toVerse := r.ChapterTo, r.VerseTo
	if toChapter < 1 {
		toChapter, toVerse = r.ChapterFrom, r.VerseFrom
	}
	if toVerse < 1 {
		toVerse = int(^uint(0) >> 1) // through the end of the Chapter
	}
	if chapter < fromChapter || chapter > toChapter {
		return false
	}
	if chapter == fromChapter && verse < fromVerse {
		return false
	}
	if chapter == toChapter && verse > toVerse {
		return false
	}
	return true
}

// DiscoverCommentaries scans dir for Commentary Modules, sorted by
// abbreviation. A Module that cannot be read is skipped and reported in
// warnings; discovery never fails.
func DiscoverCommentaries(dir string) (commentaries []Commentary, warnings []string) {
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
		if TypeOf(name) != TypeCommentaries {
			continue
		}
		commentary, err := readCommentary(path)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", name, err))
			continue
		}
		commentaries = append(commentaries, commentary)
	}

	sort.Slice(commentaries, func(i, j int) bool {
		return commentaries[i].Abbreviation < commentaries[j].Abbreviation
	})
	return commentaries, warnings
}

// readCommentary opens one Commentary Module read-only and reads its info table.
func readCommentary(path string) (Commentary, error) {
	db, err := openReadOnly(path)
	if err != nil {
		return Commentary{}, err
	}
	defer func() { _ = db.Close() }()

	tables, err := moduleTables(db)
	if err != nil {
		return Commentary{}, err
	}
	if !tables[tableCommentaries] {
		return Commentary{}, fmt.Errorf("missing the required %s table", tableCommentaries)
	}

	info := make(map[string]string)
	if tables[tableInfo] {
		info, err = infoTable(db)
		if err != nil {
			return Commentary{}, err
		}
	}
	stem := strings.TrimSuffix(filepath.Base(path), extension)
	return Commentary{
		Abbreviation: strings.TrimSuffix(stem, "."+string(TypeCommentaries)),
		Title:        info["description"],
		Language:     info["language"],
		path:         path,
	}, nil
}

// Rows reads one Chapter's Commentary rows on demand, in range order. The text
// is never kept in memory by discovery.
func (c Commentary) Rows(book, chapter int) ([]CommentaryRow, error) {
	if c.path == "" || chapter < 1 {
		return nil, nil
	}
	db, err := openReadOnly(c.path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Close() }()

	numbers := bookNumbers(book)
	args := make([]any, 0, len(numbers)+2)
	placeholders := make([]string, 0, len(numbers))
	for _, number := range numbers {
		placeholders = append(placeholders, "?")
		args = append(args, number)
	}
	args = append(args, chapter, chapter)

	rows, err := db.Query(`SELECT chapter_number_from, verse_number_from,
			COALESCE(chapter_number_to, 0), COALESCE(verse_number_to, 0), text
		FROM commentaries
		WHERE book_number IN (`+strings.Join(placeholders, ",")+`)
			AND chapter_number_from >= 1 AND chapter_number_from <= ?
			AND (COALESCE(chapter_number_to, 0) = 0 OR chapter_number_to >= ?)
		ORDER BY chapter_number_from, verse_number_from,
			COALESCE(chapter_number_to, 0), COALESCE(verse_number_to, 0)`, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var commentary []CommentaryRow
	for rows.Next() {
		var row CommentaryRow
		var text sql.NullString
		if err := rows.Scan(&row.ChapterFrom, &row.VerseFrom, &row.ChapterTo, &row.VerseTo, &text); err != nil {
			return nil, err
		}
		if row.ChapterTo < 1 && row.ChapterFrom != chapter {
			continue // a single-Chapter entry from another Chapter
		}
		row.Text = text.String
		commentary = append(commentary, row)
	}
	return commentary, rows.Err()
}

// bookNumbers is the Module book numbers that name one Canonical Book: its own,
// plus any older number the format used.
func bookNumbers(number int) []int {
	numbers := []int{number}
	for legacy, current := range legacyNumbers {
		if current == number {
			numbers = append(numbers, legacy)
		}
	}
	sort.Ints(numbers)
	return numbers
}

// spanStyle is the styling a <span> turns on until its closing tag.
type spanStyle struct {
	italics bool
	bold    bool
}

// verseLinkTag reports whether an <a> tag points at a Verse, rather than at a
// Strong's entry or another commentary note.
func verseLinkTag(raw string) bool {
	tag := strings.ReplaceAll(strings.ToLower(raw), " ", "")
	return strings.Contains(tag, "href='b:") || strings.Contains(tag, `href="b:`)
}

// ParseCommentary turns a Commentary's HTML into display Runs. Paragraph, list
// and table tags become breaks, a list item becomes a bullet, <i> and <em>
// become italics, <b>, <strong> and a bold <span> become bold, and a verse link
// keeps its text as a Link Run. Every other tag is stripped while its content
// stays. Entities are decoded.
func ParseCommentary(text string) []Run {
	var runs []Run
	var italicDepth, boldDepth, linkDepth int
	var spanStyles []spanStyle
	var hidden string
	var b strings.Builder

	flush := func() {
		if b.Len() == 0 {
			return
		}
		part := strings.ReplaceAll(html.UnescapeString(b.String()), "\u00a0", " ")
		runs = append(runs, Run{Text: part, Italics: italicDepth > 0, Bold: boldDepth > 0, Link: linkDepth > 0})
		b.Reset()
	}
	// breakWith ends the current line and adds br, unless the line already ends
	// with as much break as br asks for.
	breakWith := func(br string) {
		flush()
		if len(runs) == 0 {
			runs = append(runs, Run{Text: br})
			return
		}
		last := &runs[len(runs)-1]
		if strings.TrimSpace(last.Text) != "" {
			last.Text = strings.TrimRight(last.Text, " \t\n")
			runs = append(runs, Run{Text: br})
			return
		}
		if strings.HasSuffix(last.Text, "\n\n") || br == "\n" {
			return
		}
		last.Text = br
	}

	for i := 0; i < len(text); {
		if text[i] != '<' {
			end := strings.IndexByte(text[i:], '<')
			if end < 0 {
				end = len(text) - i
			}
			if hidden == "" {
				b.WriteString(text[i : i+end])
			}
			i += end
			continue
		}

		end := strings.IndexByte(text[i:], '>')
		if end < 0 {
			if hidden == "" {
				b.WriteString(text[i:])
			}
			break
		}
		raw := text[i+1 : i+end]
		i += end + 1
		closing := strings.HasPrefix(raw, "/")
		name := strings.ToLower(strings.Trim(raw, "/ "))
		if sp := strings.IndexAny(name, " \t\r\n/"); sp >= 0 {
			name = name[:sp]
		}

		if hidden != "" {
			if closing && name == hidden {
				hidden = ""
			}
			continue
		}
		switch {
		case (name == "script" || name == "style") && !closing:
			hidden = name
		case name == "p" || name == "div" || name == "tr" || name == "ul" || name == "ol" || name == "table":
			breakWith("\n\n")
		case name == "br":
			breakWith("\n")
		case name == "li":
			breakWith("\n")
			if !closing {
				b.WriteString("• ")
			}
		case name == "i" || name == "em":
			flush()
			if closing {
				if italicDepth > 0 {
					italicDepth--
				}
			} else {
				italicDepth++
			}
		case name == "b" || name == "strong":
			flush()
			if closing {
				if boldDepth > 0 {
					boldDepth--
				}
			} else {
				boldDepth++
			}
		case name == "a":
			flush()
			if closing {
				if linkDepth > 0 {
					linkDepth--
				}
			} else if verseLinkTag(raw) {
				linkDepth++
			}
		case name == "span":
			flush()
			if closing {
				if n := len(spanStyles); n > 0 {
					span := spanStyles[n-1]
					if span.italics && italicDepth > 0 {
						italicDepth--
					}
					if span.bold && boldDepth > 0 {
						boldDepth--
					}
					spanStyles = spanStyles[:n-1]
				}
			} else {
				span := spanStyle{
					italics: strings.Contains(strings.ToLower(raw), `class="it"`) ||
						strings.Contains(strings.ToLower(raw), `class='it'`),
					bold: strings.Contains(strings.ReplaceAll(strings.ToLower(raw), " ", ""), "font-weight:bold"),
				}
				spanStyles = append(spanStyles, span)
				if span.italics {
					italicDepth++
				}
				if span.bold {
					boldDepth++
				}
			}
		}
	}
	flush()
	return trimCommentaryRuns(runs)
}

// trimCommentaryRuns drops the whitespace-only Runs a leading or trailing block
// tag leaves behind.
func trimCommentaryRuns(runs []Run) []Run {
	for len(runs) > 0 && strings.TrimSpace(runs[0].Text) == "" {
		runs = runs[1:]
	}
	for len(runs) > 0 && strings.TrimSpace(runs[len(runs)-1].Text) == "" {
		runs = runs[:len(runs)-1]
	}
	return runs
}
