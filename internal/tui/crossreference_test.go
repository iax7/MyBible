package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/iax7/mybible/internal/config"
	"github.com/iax7/mybible/internal/modtest"
)

// crossReferenceReader builds a Bible whose first Chapter carries links from two
// sets: a target both sets share, a source range, and a target outside the
// Active translation.
func crossReferenceReader(t *testing.T) Model {
	t.Helper()

	dir := t.TempDir()
	modtest.BibleAll(t, dir, "RST.SQLite3",
		map[string]string{"description": "Reina-Valera 1909"},
		genesis(),
		[]modtest.Verse{
			{Book: 10, Chapter: 1, Verse: 1, Text: "verse01"},
			{Book: 10, Chapter: 1, Verse: 2, Text: "verse02"},
			{Book: 10, Chapter: 1, Verse: 3, Text: "verse03"},
			{Book: 10, Chapter: 1, Verse: 4, Text: "verse04"},
			{Book: 10, Chapter: 2, Verse: 1, Text: "chapter2"},
		})
	modtest.CrossReferences(t, dir, "TSK-x.crossreferences.SQLite3",
		map[string]string{"description": "The Treasury of Scripture Knowledge"},
		[]modtest.CrossReference{
			{Book: 10, Chapter: 1, Verse: 1, ToBook: 10, ToChapter: 2, ToVerse: 1, Votes: 9},
			{Book: 10, Chapter: 1, Verse: 1, ToBook: 500, ToChapter: 1, ToVerse: 1, Votes: 3},
			{Book: 10, Chapter: 1, Verse: 2, VerseEnd: 3, ToBook: 20, ToChapter: 1, ToVerse: 5, Votes: 1},
			{Book: 10, Chapter: 1, Verse: 3, ToBook: 10, ToChapter: 2, ToVerse: 1, Votes: 4},
		})
	modtest.CrossReferences(t, dir, "OBX.crossreferences.SQLite3",
		map[string]string{"description": "OpenBible"},
		[]modtest.CrossReference{
			// The same target as the Treasury's most voted link, with fewer votes.
			{Book: 10, Chapter: 1, Verse: 1, ToBook: 10, ToChapter: 2, ToVerse: 1, Votes: 2},
		})
	return configured(t, config.Config{
		ModuleDir:             dir,
		ActiveTranslation:     "RST",
		ActiveCrossReferences: []string{"OBX", "TSK-x"},
	})
}

func TestTheReadingPaneMarksVersesWithCrossReferences(t *testing.T) {
	got := render(t, crossReferenceReader(t))

	for _, want := range []string{
		"1 verse01 [⇅2]", // two targets, the shared one counted once
		"2 verse02 [⇅1]", // the source range covers it
		"3 verse03 [⇅2]", // the range and the link of its own
	} {
		if !strings.Contains(got, want) {
			t.Errorf("the reading pane is missing %q\n%s", want, got)
		}
	}
	if strings.Contains(got, "4 verse04 [") {
		t.Errorf("a Verse with no link carries a marker\n%s", got)
	}
}

func TestTheMarkerTakesTheLinkStyle(t *testing.T) {
	m := crossReferenceReader(t)
	marker := m.styles.link.Render("[⇅2]")
	if !strings.Contains(m.View().Content, marker) {
		t.Errorf("the marker is not drawn in the link style\n%q", m.View().Content)
	}
}

func TestXOpensTheReferencesWithTheirText(t *testing.T) {
	m := crossReferenceReader(t)

	m, _ = send(t, m, press('x', "x", 0))
	got := modalText(t, m)
	for _, want := range []string{
		"References · Genesis 1:1",
		"Genesis 2:1", // the most voted first
		"chapter2",    // the text it points to
		"John 1:1",    // a Book the Active translation lacks keeps its Canonical name
	} {
		if !strings.Contains(got, want) {
			t.Errorf("the References popup is missing %q\n%s", want, got)
		}
	}
	if strings.Index(got, "Genesis 2:1") > strings.Index(got, "John 1:1") {
		t.Errorf("the links are not ordered by votes\n%s", got)
	}

	m, _ = send(t, m, press(tea.KeyEscape, "", 0))
	if got := modalText(t, m); got != "" {
		t.Errorf("Esc did not close the References popup\n%s", got)
	}
}

func TestXExplainsTheEmptyStates(t *testing.T) {
	m := crossReferenceReader(t)
	m, _ = send(t, m, press('j', "j", 0))
	m, _ = send(t, m, press('j', "j", 0))
	m, _ = send(t, m, press('j', "j", 0))

	m, _ = send(t, m, press('x', "x", 0))
	if got := modalText(t, m); !strings.Contains(got, "No references for this verse.") {
		t.Errorf("the popup does not explain a Verse with no links\n%s", got)
	}
	m, _ = send(t, m, press(tea.KeyEscape, "", 0))

	m.activeCrossReferences = []string{}
	m = m.loadCrossReferences()
	m, _ = send(t, m, press('x', "x", 0))
	if got := modalText(t, m); !strings.Contains(got, "No cross-reference modules active.") {
		t.Errorf("the popup does not explain that no set is on\n%s", got)
	}
}

func TestTheSourceRangeOfALinkIsShownInItsLabel(t *testing.T) {
	m := crossReferenceReader(t)
	m, _ = send(t, m, press('j', "j", 0)) // verse 2, covered by the range
	m, _ = send(t, m, press('x', "x", 0))

	if got := modalText(t, m); !strings.Contains(got, "(2–3) Exodus 1:5") {
		t.Errorf("the popup does not name the range the link speaks from\n%s", got)
	}
}

func TestSelectingSetsTurnsTheMarkersOnAndOff(t *testing.T) {
	m := crossReferenceReader(t)

	m, _ = send(t, m, press('X', "X", 0))
	got := modalText(t, m)
	for _, want := range []string{"Cross-reference modules", "[x] OBX", "[x] TSK-x"} {
		if !strings.Contains(got, want) {
			t.Errorf("the selector is missing %q\n%s", want, got)
		}
	}

	m, _ = send(t, m, press(' ', " ", 0)) // turn OBX off; the Treasury still marks verse 1
	if !strings.Contains(render(t, m), "1 verse01 [⇅2]") {
		t.Errorf("turning a set off dropped the marker another set still explains\n%s", render(t, m))
	}
	if strings.Contains(modalText(t, m), "[x] OBX") {
		t.Errorf("the selector still marks OBX as on\n%s", modalText(t, m))
	}

	m, _ = send(t, m, press(tea.KeyEscape, "", 0))
	if got := modalText(t, m); got != "" {
		t.Errorf("Esc did not close the selector\n%s", got)
	}
	if !m.crossReferenceActive("TSK-x") || m.crossReferenceActive("OBX") {
		t.Errorf("the active sets are %v, want TSK-x alone", m.activeCrossReferences)
	}
}

func TestTheCrossReferenceSetsSurviveARestart(t *testing.T) {
	dir := t.TempDir()
	modtest.BibleAll(t, dir, "RST.SQLite3",
		map[string]string{"description": "Reina-Valera 1909"},
		genesis(),
		[]modtest.Verse{{Book: 10, Chapter: 1, Verse: 1, Text: "verse01"}})
	modtest.CrossReferences(t, dir, "OBX.crossreferences.SQLite3",
		map[string]string{"description": "OpenBible"}, nil)
	modtest.CrossReferences(t, dir, "TSK-x.crossreferences.SQLite3",
		map[string]string{"description": "The Treasury of Scripture Knowledge"}, nil)

	configDir := t.TempDir()
	if err := config.Save(configDir, config.Config{ModuleDir: dir, ActiveTranslation: "RST"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	m := newReader(t, configDir, dir)
	if want := []string{"OBX"}; !equalStrings(m.activeCrossReferences, want) {
		t.Fatalf("a config that never chose starts with %v, want %v", m.activeCrossReferences, want)
	}

	m, _ = send(t, m, press('X', "X", 0))
	m, _ = send(t, m, press(tea.KeyEnter, "", 0)) // turn OBX off
	m, _ = send(t, m, press(tea.KeyDown, "", 0))
	m, _ = send(t, m, press(tea.KeyEnter, "", 0)) // turn TSK-x on
	m, _ = send(t, m, press(tea.KeyEscape, "", 0))

	restarted := New(Options{ConfigDir: configDir, ModuleDir: dir})
	if want := []string{"TSK-x"}; !equalStrings(restarted.activeCrossReferences, want) {
		t.Errorf("the restart holds %v, want %v", restarted.activeCrossReferences, want)
	}

	// Turning every set off is a choice, not a reason to bring the first one back.
	m.activeCrossReferences = []string{}
	m = m.persist()
	restarted = New(Options{ConfigDir: configDir, ModuleDir: dir})
	if len(restarted.activeCrossReferences) != 0 {
		t.Errorf("the restart brought back %v after the reader turned everything off", restarted.activeCrossReferences)
	}
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// A Footnotes Module is paired with its translation by abbreviation: the marker
// the Verse text carries opens its note.
func TestFootnoteMarkersAreShownAndOpenTheirNote(t *testing.T) {
	dir := t.TempDir()
	modtest.BibleAll(t, dir, "RST.SQLite3",
		map[string]string{"description": "Reina-Valera 1909", "add_space_before_footnote_marker": "true"},
		genesis(),
		[]modtest.Verse{
			{Book: 10, Chapter: 1, Verse: 1, Text: "la superficie<f>[1]</f> del abismo"},
		})
	modtest.Commentaries(t, dir, "RST.commentaries.SQLite3",
		map[string]string{"description": "Notas", "is_footnotes": "true"},
		[]modtest.Commentary{
			{Book: 10, ChapterFrom: 1, VerseFrom: 1, Marker: "[1]", Text: "Lit. faz."},
		})
	m := configured(t, config.Config{ModuleDir: dir, ActiveTranslation: "RST", ActiveCommentary: "RST"})

	got := render(t, m)
	if !strings.Contains(got, "la superficie [1] del abismo") {
		t.Errorf("the Footnote marker is not spaced and shown\n%s", got)
	}
	if !strings.Contains(m.View().Content, m.styles.link.Render("[1]")) {
		t.Errorf("the Footnote marker is not drawn in the link style\n%q", m.View().Content)
	}

	m, _ = send(t, m, press(tea.KeyEnter, "", 0))
	raw := m.modalView()
	if !strings.Contains(raw, m.styles.link.Render("[1] ")+"Lit. faz.") {
		t.Errorf("the note is not shown behind its marker\n%q", raw)
	}
}

func TestTheSelectorExplainsWhenNoSetIsInstalled(t *testing.T) {
	m := activeReader(t, map[string]string{"description": "Reina-Valera 1909"}, genesis(),
		[]modtest.Verse{{Book: 10, Chapter: 1, Verse: 1, Text: "verse01"}})

	m, _ = send(t, m, press('X', "X", 0))
	got := modalText(t, m)
	for _, want := range []string{
		"Cross-reference modules",
		"No cross-reference modules found.",
		"Put your *.crossreferences.SQLite3 modules there.",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("the selector is missing %q\n%s", want, got)
		}
	}
}
