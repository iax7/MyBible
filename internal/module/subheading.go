package module

import (
	"strconv"
	"strings"
	"unicode"
)

// headingRuns turns a Subheading's raw text into display Runs. An
// <x>target</x> tag carries its target as content rather than anchor text, so it
// becomes a Link Run holding a readable label. A malformed target is hidden like
// any other <x>. The text around a target goes through displayText, so the other
// markup a Subheading can carry stays hidden the way it always was.
func headingRuns(text string) []Run {
	var runs []Run
	for {
		start, after := nextTag(text, "x")
		if start < 0 {
			runs = appendHeadingText(runs, displayText(text))
			break
		}
		runs = appendHeadingText(runs, displayText(text[:start]))

		closeStart, closeEnd := nextClosingTag(text[after:], "x")
		if closeStart < 0 {
			break // an unclosed <x> hides the rest, as ParseMarkup does
		}
		if label, ok := headingLabel(text[after : after+closeStart]); ok {
			runs = append(runs, Run{Text: label, Link: true})
		}
		text = text[after+closeEnd:]
	}
	return trimHeadingRuns(runs)
}

// appendHeadingText adds literal heading text to runs, joining it onto the
// preceding plain Run so a tag boundary does not split the text needlessly.
func appendHeadingText(runs []Run, text string) []Run {
	if text == "" {
		return runs
	}
	if n := len(runs); n > 0 && !runs[n-1].Link {
		runs[n-1].Text += text
		return runs
	}
	return append(runs, Run{Text: text})
}

// trimHeadingRuns removes the whitespace a source leaves around a heading and
// drops the Runs that become empty.
func trimHeadingRuns(runs []Run) []Run {
	for len(runs) > 0 {
		runs[0].Text = strings.TrimLeftFunc(runs[0].Text, unicode.IsSpace)
		if runs[0].Text != "" {
			break
		}
		runs = runs[1:]
	}
	for len(runs) > 0 {
		last := len(runs) - 1
		runs[last].Text = strings.TrimRightFunc(runs[last].Text, unicode.IsSpace)
		if runs[last].Text != "" {
			break
		}
		runs = runs[:last]
	}
	return runs
}

// headingText is the flattened display text of a Subheading's Runs.
func headingText(runs []Run) string {
	var b strings.Builder
	for _, run := range runs {
		b.WriteString(run.Text)
	}
	return b.String()
}

// headingLabel turns an <x> target into the label the reader shows. The target
// is MyBible's dictionary hyperlink format: a Canonical book number, then a
// location, with an optional leading "B:". Nothing but a readable label comes
// out of it; the target is not used for navigation.
func headingLabel(target string) (string, bool) {
	target = strings.TrimSpace(target)
	if len(target) >= 2 && (target[0] == 'b' || target[0] == 'B') && target[1] == ':' {
		target = strings.TrimSpace(target[2:])
	}

	i := 0
	for i < len(target) && target[i] >= '0' && target[i] <= '9' {
		i++
	}
	if i == 0 {
		return "", false
	}
	number, ok := positiveNumber(target[:i])
	if !ok {
		return "", false
	}
	book, ok := canonicalBook(number)
	if !ok {
		return "", false
	}

	location := strings.TrimSpace(target[i:])
	if location == "" {
		return "", false
	}
	label, ok := headingLocationLabel(location)
	if !ok {
		return "", false
	}
	return book.Name + " " + label, true
}

// headingLocationLabel renders the location part of a target: a Chapter, a
// Chapter range, or a Chapter and its Verses, Verses range, cross-Chapter range,
// or comma-separated Verse list.
func headingLocationLabel(location string) (string, bool) {
	location = strings.TrimSpace(location)
	if location == "" {
		return "", false
	}

	if chapterPart, versePart, ok := splitColon(location); ok {
		chapter, ok := positiveNumber(chapterPart)
		if !ok {
			return "", false
		}
		verses, ok := headingVerseLabel(versePart)
		if !ok {
			return "", false
		}
		return strconv.Itoa(chapter) + ":" + verses, true
	}

	from, to, ranged, ok := splitRange(location)
	if !ok {
		return "", false
	}
	chapter, ok := positiveNumber(from)
	if !ok {
		return "", false
	}
	if !ranged {
		return strconv.Itoa(chapter), true
	}
	end, ok := positiveNumber(to)
	if !ok {
		return "", false
	}
	return strconv.Itoa(chapter) + "–" + strconv.Itoa(end), true
}

// headingVerseLabel renders the part after a Chapter's colon: one Verse, a
// same-Chapter range, a cross-Chapter range, or a comma-separated list.
func headingVerseLabel(verses string) (string, bool) {
	verses = strings.TrimSpace(verses)
	if verses == "" {
		return "", false
	}

	if from, to, ranged, ok := splitRange(verses); ok {
		if ranged {
			start, ok := positiveNumber(from)
			if !ok {
				return "", false
			}
			if endChapter, endVerse, ok := splitColon(to); ok {
				chapter, ok := positiveNumber(endChapter)
				if !ok {
					return "", false
				}
				verse, ok := positiveNumber(endVerse)
				if !ok {
					return "", false
				}
				return strconv.Itoa(start) + "–" + strconv.Itoa(chapter) + ":" + strconv.Itoa(verse), true
			}
			end, ok := positiveNumber(to)
			if !ok {
				return "", false
			}
			return strconv.Itoa(start) + "–" + strconv.Itoa(end), true
		}
	}

	if strings.Contains(verses, ",") {
		parts := strings.Split(verses, ",")
		numbers := make([]string, 0, len(parts))
		for _, part := range parts {
			number, ok := positiveNumber(part)
			if !ok {
				return "", false
			}
			numbers = append(numbers, strconv.Itoa(number))
		}
		return strings.Join(numbers, ","), true
	}

	number, ok := positiveNumber(verses)
	if !ok {
		return "", false
	}
	return strconv.Itoa(number), true
}

// splitRange divides a location at its one hyphen. A location with no hyphen is
// not a range; one with more than one is malformed.
func splitRange(text string) (from, to string, ranged, ok bool) {
	i := strings.IndexByte(text, '-')
	if i < 0 {
		return strings.TrimSpace(text), "", false, true
	}
	if strings.IndexByte(text[i+1:], '-') >= 0 {
		return "", "", false, false
	}
	return strings.TrimSpace(text[:i]), strings.TrimSpace(text[i+1:]), true, true
}

// splitColon divides a Chapter from its Verses at the first colon. A second
// colon belongs to a cross-Chapter end and is left to the Verse parser.
func splitColon(text string) (chapter, verses string, ok bool) {
	i := strings.IndexByte(text, ':')
	if i < 0 {
		return "", "", false
	}
	return strings.TrimSpace(text[:i]), strings.TrimSpace(text[i+1:]), true
}

// positiveNumber parses a run of digits, rejecting zero and anything else.
func positiveNumber(text string) (int, bool) {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0, false
	}
	for _, r := range text {
		if r < '0' || r > '9' {
			return 0, false
		}
	}
	number, err := strconv.Atoi(text)
	if err != nil || number < 1 {
		return 0, false
	}
	return number, true
}

// nextTag finds the next opening tag named name, returning where the tag starts
// and where it ends.
func nextTag(text, name string) (start, end int) {
	for i := 0; i < len(text); {
		lt := strings.IndexByte(text[i:], '<')
		if lt < 0 {
			return -1, -1
		}
		start = i + lt
		gt := strings.IndexByte(text[start:], '>')
		if gt < 0 {
			return -1, -1
		}
		raw := text[start+1 : start+gt]
		if !strings.HasPrefix(strings.TrimSpace(raw), "/") && tagName(raw) == name {
			return start, start + gt + 1
		}
		i = start + gt + 1
	}
	return -1, -1
}

// nextClosingTag finds the next closing tag named name, returning where it
// starts and ends.
func nextClosingTag(text, name string) (start, end int) {
	for i := 0; i < len(text); {
		lt := strings.IndexByte(text[i:], '<')
		if lt < 0 {
			return -1, -1
		}
		start = i + lt
		gt := strings.IndexByte(text[start:], '>')
		if gt < 0 {
			return -1, -1
		}
		raw := text[start+1 : start+gt]
		if strings.HasPrefix(strings.TrimSpace(raw), "/") && tagName(raw) == name {
			return start, start + gt + 1
		}
		i = start + gt + 1
	}
	return -1, -1
}

// tagName lowercases a tag and strips its slashes and surrounding spaces.
func tagName(raw string) string {
	return strings.ToLower(strings.Trim(raw, "/ "))
}
