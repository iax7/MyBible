package tui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/iax7/mybible/internal/module"
)

// historyFile holds the Visits, one JSON object per line, beside config.toml.
const historyFile = "history.jsonl"

// visit is one Reference reached by a deliberate jump, at the moment it was
// reached. History keeps Visits across restarts; Trail keeps the session's.
type visit struct {
	Reference module.Reference
	At        time.Time
}

// visitRecord is a Visit as it is stored: the Book by number, so a History entry
// is independent of the Translation that was active when it was made.
type visitRecord struct {
	Book    int       `json:"book"`
	Chapter int       `json:"chapter"`
	Verse   int       `json:"verse"`
	At      time.Time `json:"at"`
}

// recordTrail appends the newly reached Reference to the session's Trail. A jump
// made after going back drops the entries that would be revisited, and the
// current Reference is never repeated.
func (m Model) recordTrail(previous module.Reference) Model {
	current, ok := m.currentReference()
	if !ok {
		return m
	}
	if len(m.trail) == 0 {
		m.trail, m.position = []module.Reference{previous}, 0
	}
	m.trail = m.trail[:m.position+1]
	if m.trail[len(m.trail)-1] != current {
		m.trail = append(m.trail, current)
		m.position = len(m.trail) - 1
	}
	return m
}

// recordVisit adds the current Reference to History, replacing that day's earlier
// Visit to the same Reference so History is one entry per Reference per day, and
// appends it to the file at once so an abrupt exit loses nothing.
func (m Model) recordVisit() Model {
	current, ok := m.currentReference()
	if !ok {
		return m
	}
	v := visit{Reference: current, At: m.clock()}
	m.history = addVisit(m.history, v)
	return m.saveVisit(v)
}

// addVisit keeps one Visit per Reference per day, with the latest Visit last.
func addVisit(history []visit, v visit) []visit {
	key := keyOf(v)
	kept := history[:0]
	for _, old := range history {
		if keyOf(old) != key {
			kept = append(kept, old)
		}
	}
	return append(kept, v)
}

// visitKey names the day and Reference a Visit belongs to.
type visitKey struct {
	day                  string
	book, chapter, verse int
}

func keyOf(v visit) visitKey {
	return visitKey{
		day:     v.At.In(time.Local).Format(time.DateOnly),
		book:    v.Reference.Book.Number,
		chapter: v.Reference.Chapter,
		verse:   v.Reference.Verse,
	}
}

// loadHistory reads the Visits kept from earlier runs. A missing file is a
// reader that has not visited anything yet; a damaged line is skipped and
// reported once rather than stopping the reader.
func (m Model) loadHistory() Model {
	if m.configDir == "" {
		return m
	}
	f, err := os.Open(filepath.Join(m.configDir, historyFile))
	if os.IsNotExist(err) {
		return m
	}
	if err != nil {
		return m.warnHistory(err)
	}
	defer func() { _ = f.Close() }()

	damaged := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var record visitRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			damaged = true
			continue
		}
		book, ok := module.BookByNumber(record.Book)
		if !ok || record.Chapter < 1 || record.Verse < 1 || record.At.IsZero() {
			damaged = true
			continue
		}
		m.history = addVisit(m.history, visit{
			Reference: module.Reference{Book: book, Chapter: record.Chapter, Verse: record.Verse},
			At:        record.At,
		})
	}
	if err := scanner.Err(); err != nil {
		return m.warnHistory(err)
	}
	if damaged {
		m.warnings = append(m.warnings, fmt.Sprintf("%s: some History entries could not be read", filepath.Join(m.configDir, historyFile)))
	}
	return m
}

// saveVisit appends one Visit to the History file.
func (m Model) saveVisit(v visit) Model {
	if m.configDir == "" {
		return m.warnHistory(fmt.Errorf("no config directory"))
	}
	if err := os.MkdirAll(m.configDir, 0o755); err != nil {
		return m.warnHistory(err)
	}
	f, err := os.OpenFile(filepath.Join(m.configDir, historyFile), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return m.warnHistory(err)
	}
	defer func() { _ = f.Close() }()

	err = json.NewEncoder(f).Encode(visitRecord{
		Book:    v.Reference.Book.Number,
		Chapter: v.Reference.Chapter,
		Verse:   v.Reference.Verse,
		At:      v.At,
	})
	if err != nil {
		return m.warnHistory(err)
	}
	return m
}

// warnHistory reports a History read or write failure once, so a failing disk
// does not fill the Log with one warning per jump.
func (m Model) warnHistory(err error) Model {
	if m.historyFailed {
		return m
	}
	m.historyFailed = true
	m.warnings = append(m.warnings, "History: "+err.Error())
	return m
}

// clock is the current time, or the wall clock when a Model was built without a
// clock of its own.
func (m Model) clock() time.Time {
	if m.now == nil {
		return time.Now()
	}
	return m.now()
}

// historyDay is the heading a Visit is grouped under: Today, Yesterday, or its
// calendar date.
func (m Model) historyDay(at time.Time) string {
	day := at.In(time.Local)
	today := m.clock().In(time.Local)
	switch {
	case sameDay(day, today):
		return m.translate("Today")
	case sameDay(day, today.AddDate(0, 0, -1)):
		return m.translate("Yesterday")
	}
	return day.Format(time.DateOnly)
}

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}

// back moves to the previous Visit in the Trail, adding no entry.
func (m Model) back() Model { return m.stepTrail(-1) }

// forward moves to the next Visit in the Trail, adding no entry.
func (m Model) forward() Model { return m.stepTrail(1) }

// stepTrail jumps to the Reference a delta away in the Trail without recording
// the move, staying put at either end.
func (m Model) stepTrail(delta int) Model {
	position := m.position + delta
	if position < 0 || position >= len(m.trail) {
		return m
	}
	next, ok := m.jumpTo(m.trail[position])
	if !ok {
		return m
	}
	next.position = position
	return next
}

// confirmHistory jumps to the highlighted History Visit, recording it as a jump
// the way any other deliberate move is. A Reference the Active translation lacks
// keeps the modal open and says so, so the entry is not lost silently.
func (m Model) confirmHistory() Model {
	items := m.pickerMatches()
	if m.picker.cursor < 0 || m.picker.cursor >= len(items) {
		return m
	}
	next, ok := m.jumpAndRecord(m.history[items[m.picker.cursor].value].Reference)
	if !ok {
		m.picker.err = m.translate("No such Reference in the Active translation.")
		return m
	}
	next.picker = choice{}
	return next
}
