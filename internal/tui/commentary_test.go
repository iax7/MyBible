package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/iax7/mybible/internal/config"
	"github.com/iax7/mybible/internal/modtest"
)

// commentaryReader builds a Bible with a commentary module active, covering a
// single Verse, a range, and the Chapter as a whole.
func commentaryReader(t *testing.T) Model {
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
	modtest.Commentaries(t, dir, "MAC.commentaries.SQLite3",
		map[string]string{"description": "MacArthur Study Bible"},
		[]modtest.Commentary{
			{Book: 10, ChapterFrom: 1, VerseFrom: 1, Text: `<p>First <b>comment</b> &quot;quoted&quot; <a href='B:10 1:2'>verse 02</a></p>`},
			{Book: 10, ChapterFrom: 1, VerseFrom: 2, ChapterTo: 1, VerseTo: 3, Text: "Second comment"},
			{Book: 10, ChapterFrom: 1, VerseFrom: 0, Text: "Chapter overview"},
		})
	return configured(t, config.Config{
		ModuleDir:         dir,
		ActiveTranslation: "RST",
		ActiveCommentary:  "MAC",
	})
}

func TestCommentaryLabelsMarkTheVersesTheActiveCommentaryCovers(t *testing.T) {
	m := commentaryReader(t)
	got := render(t, m)

	for _, want := range []string{"verse01 [MAC]", "verse02 [MAC]", "verse03 [MAC]"} {
		if !strings.Contains(got, want) {
			t.Errorf("the reading pane is missing %q\n%s", want, got)
		}
	}
	if strings.Contains(got, "verse04 [MAC]") {
		t.Errorf("a Verse with no commentary carries the label\n%s", got)
	}

	m, _ = send(t, m, press(']', "]", 0)) // next Chapter has no commentary
	if got := render(t, m); strings.Contains(got, "[MAC]") {
		t.Errorf("the label followed the reader into a Chapter with no commentary\n%s", got)
	}
}

func TestEnterOpensTheCommentaryAndEscCloses(t *testing.T) {
	m := commentaryReader(t)

	m, _ = send(t, m, press(tea.KeyEnter, "", 0))
	got := modalText(t, m)
	for _, want := range []string{"Commentary · Genesis 1:1", "MAC", "First comment \"quoted\"", "Chapter overview"} {
		if !strings.Contains(got, want) {
			t.Errorf("the commentary modal is missing %q\n%s", want, got)
		}
	}

	m, _ = send(t, m, press(tea.KeyEscape, "", 0))
	if got := modalText(t, m); got != "" {
		t.Errorf("Esc did not close the commentary modal\n%s", got)
	}
}

func TestCommentaryBoldAndVerseLinksTakeTheirStyles(t *testing.T) {
	m := commentaryReader(t)
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))
	raw := m.modalView()

	for _, want := range []string{
		m.styles.commentaryBold.Render("comment"),
		m.styles.link.Render("verse 02"),
	} {
		if !strings.Contains(raw, want) {
			t.Errorf("the commentary modal is missing %q\n%q", want, raw)
		}
	}
	if !m.styles.link.GetUnderline() {
		t.Error("the verse link is not underlined")
	}
	if strings.Contains(raw, m.styles.commentaryBold.Render("First")) {
		t.Errorf("the plain commentary text took the bold style\n%q", raw)
	}
}

func TestTheCommentaryModalHasOneTabPerModule(t *testing.T) {
	m := commentaryReader(t)
	modtest.Commentaries(t, m.moduleDir, "ALT.commentaries.SQLite3",
		map[string]string{"description": "Another commentary"},
		[]modtest.Commentary{{Book: 10, ChapterFrom: 1, VerseFrom: 1, Text: "Other comment"}})

	// Discovery happens at startup, so reload the reader with the new module.
	m = configured(t, config.Config{
		ModuleDir:         m.moduleDir,
		ActiveTranslation: "RST",
		ActiveCommentary:  "MAC",
	})
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))

	got := modalText(t, m)
	for _, want := range []string{"MAC", "ALT", "First comment"} {
		if !strings.Contains(got, want) {
			t.Fatalf("the commentary modal is missing %q\n%s", want, got)
		}
	}

	m, _ = send(t, m, press(tea.KeyTab, "", 0))
	got = modalText(t, m)
	if !strings.Contains(got, "Other comment") {
		t.Errorf("Tab did not switch to the other commentary\n%s", got)
	}
	if strings.Contains(got, "First comment") {
		t.Errorf("the previous tab's text is still shown\n%s", got)
	}

	m, _ = send(t, m, press(tea.KeyTab, "", tea.ModShift))
	if got := modalText(t, m); !strings.Contains(got, "First comment") {
		t.Errorf("Shift-Tab did not switch back\n%s", got)
	}
}

func TestCommentaryModalScrollsLongText(t *testing.T) {
	m := commentaryReader(t)
	var long strings.Builder
	for i := 1; i <= 60; i++ {
		fmt.Fprintf(&long, "line %02d\n", i)
	}
	modtest.Commentaries(t, m.moduleDir, "LONG.commentaries.SQLite3",
		map[string]string{"description": "Long commentary"},
		[]modtest.Commentary{{Book: 10, ChapterFrom: 1, VerseFrom: 1, Text: long.String()}})
	m = configured(t, config.Config{
		ModuleDir:         m.moduleDir,
		ActiveTranslation: "RST",
		ActiveCommentary:  "LONG",
	})
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))

	before := modalText(t, m)
	if !strings.Contains(before, "line 01") {
		t.Fatalf("the long commentary did not open at its start\n%s", before)
	}

	m, _ = send(t, m, press('j', "j", 0))
	after := modalText(t, m)
	if strings.Contains(after, "line 01") {
		t.Errorf("j did not scroll the commentary\n%s", after)
	}
}

func TestNoLabelOrModalWhenTheActiveCommentaryIsSilent(t *testing.T) {
	m := commentaryReader(t)
	m, _ = send(t, m, press('G', "G", 0)) // the Chapter's last Verse has no commentary

	if got := render(t, m); strings.Contains(got, "verse04 [MAC]") {
		t.Errorf("the last Verse carries a label it should not\n%s", got)
	}
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))
	if got := modalText(t, m); got != "" {
		t.Errorf("Enter opened a commentary modal for a Verse with no commentary\n%s", got)
	}
}

func TestChapterLevelCommentaryLabelsOnlyTheFirstVerse(t *testing.T) {
	dir := t.TempDir()
	modtest.BibleAll(t, dir, "RST.SQLite3",
		map[string]string{"description": "Reina-Valera 1909"},
		genesis(),
		[]modtest.Verse{
			{Book: 10, Chapter: 1, Verse: 1, Text: "verse01"},
			{Book: 10, Chapter: 1, Verse: 2, Text: "verse02"},
		})
	modtest.Commentaries(t, dir, "MAC.commentaries.SQLite3",
		map[string]string{"description": "MacArthur Study Bible"},
		[]modtest.Commentary{{Book: 10, ChapterFrom: 1, VerseFrom: 0, Text: "Overview"}})

	m := configured(t, config.Config{ModuleDir: dir, ActiveTranslation: "RST", ActiveCommentary: "MAC"})
	got := render(t, m)
	if !strings.Contains(got, "verse01 [MAC]") {
		t.Errorf("a Chapter-level commentary did not label Verse 1\n%s", got)
	}
	if strings.Contains(got, "verse02 [MAC]") {
		t.Errorf("a Chapter-level commentary labelled every Verse\n%s", got)
	}
}

func TestTheCommentaryPickerChangesTheActiveCommentary(t *testing.T) {
	m := commentaryReader(t)
	modtest.Commentaries(t, m.moduleDir, "ALT.commentaries.SQLite3",
		map[string]string{"description": "Another commentary"},
		[]modtest.Commentary{{Book: 10, ChapterFrom: 1, VerseFrom: 1, Text: "Other comment"}})
	m = configured(t, config.Config{
		ModuleDir:         m.moduleDir,
		ActiveTranslation: "RST",
		ActiveCommentary:  "MAC",
	})

	m, _ = send(t, m, press('m', "m", 0))
	m = typeKey(t, m, "ALT")
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))

	if m.activeCommentary != "ALT" {
		t.Errorf("activeCommentary = %q, want ALT", m.activeCommentary)
	}
	got := render(t, m)
	if !strings.Contains(got, "verse01 [ALT]") || strings.Contains(got, "verse01 [MAC]") {
		t.Errorf("the reading pane still shows the old commentary label\n%s", got)
	}
}
