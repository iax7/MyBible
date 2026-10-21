package tui

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/iax7/mybible/internal/module"
)

// splitVerseLines separates a Verse's Runs at the breaks their text carries, one
// []Run per rendered line. A break that opens the Verse is dropped: Modules mark
// the Verse that begins a paragraph with one, and it says nothing where the Verse
// number already stands.
func splitVerseLines(runs []module.Run) [][]module.Run {
	lines := [][]module.Run{{}}
	for _, run := range runs {
		text := run.Text
		for {
			i := strings.IndexByte(text, '\n')
			if i < 0 {
				lines[len(lines)-1] = appendRun(lines[len(lines)-1], withText(run, text))
				break
			}
			lines[len(lines)-1] = appendRun(lines[len(lines)-1], withText(run, text[:i]))
			lines = append(lines, nil)
			text = text[i+1:]
		}
	}
	for len(lines) > 1 && len(lines[0]) == 0 {
		lines = lines[1:]
	}
	return lines
}

// wrapRuns breaks one line of a Verse into lines of at most width cells, between
// words, keeping each Run's style. A word wider than a line is left whole rather
// than cut. An empty line yields one empty line.
func wrapRuns(runs []module.Run, width int) [][]module.Run {
	if width < 1 {
		width = 1
	}
	var lines [][]module.Run
	var line []module.Run
	lineWidth := 0
	for _, word := range splitWords(runs) {
		wordWidth := runWidth(word.runs)
		if len(line) > 0 && lineWidth+1+wordWidth > width {
			lines = append(lines, line)
			line, lineWidth = nil, 0
		}
		if len(line) > 0 {
			line = appendRun(line, word.space)
			lineWidth++
		}
		for _, run := range word.runs {
			line = appendRun(line, run)
		}
		lineWidth += wordWidth
	}
	return append(lines, line)
}

// splitWord is one word of a line and the whitespace before it. The whitespace
// keeps its own style, so the spaces inside a link stay underlined when the
// line is rejoined.
type splitWord struct {
	space module.Run
	runs  []module.Run
}

// splitWords separates a line's Runs into words, dropping the whitespace between
// them: a wrapped line rejoins its words with single spaces.
func splitWords(runs []module.Run) []splitWord {
	var words []splitWord
	var word splitWord
	var space module.Run
	flush := func() {
		if len(word.runs) > 0 {
			word.space = space
			words = append(words, word)
			word = splitWord{}
			space = module.Run{}
		}
	}
	for _, run := range runs {
		start := 0
		for i, r := range run.Text {
			if r != ' ' && r != '\t' && r != '\n' {
				continue
			}
			word.runs = appendRun(word.runs, withText(run, run.Text[start:i]))
			flush()
			space = withText(run, " ")
			start = i + len(string(r))
		}
		word.runs = appendRun(word.runs, withText(run, run.Text[start:]))
	}
	flush()
	return words
}

// appendRun adds a Run to a line, merging it with the one before when they share
// a style, so wrapping does not fragment a line into one Run per word.
func appendRun(runs []module.Run, run module.Run) []module.Run {
	if run.Text == "" {
		return runs
	}
	if n := len(runs); n > 0 && runs[n-1].Italics == run.Italics && runs[n-1].Red == run.Red &&
		runs[n-1].Bold == run.Bold && runs[n-1].Link == run.Link {
		runs[n-1].Text += run.Text
		return runs
	}
	return append(runs, run)
}

// runWidth is the display width of a line of Runs.
func runWidth(runs []module.Run) int {
	width := 0
	for _, run := range runs {
		width += lipgloss.Width(run.Text)
	}
	return width
}

// withText is a Run's style carried onto a slice of its text.
func withText(run module.Run, text string) module.Run {
	return module.Run{Text: text, Italics: run.Italics, Red: run.Red, Bold: run.Bold, Link: run.Link}
}
