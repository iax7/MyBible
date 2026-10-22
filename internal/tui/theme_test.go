package tui

import (
	"strings"
	"testing"

	"github.com/iax7/mybible/internal/config"
	"github.com/iax7/mybible/internal/modtest"
)

// themedReader builds a reader whose Verse contains the words of Jesus, read
// with the given theme settings.
func themedReader(t *testing.T, theme string, colors config.Colors) Model {
	t.Helper()
	return translatorReader(t, theme, colors, "and <J>Jesus wept</J>")
}

// translatorReader builds a reader whose only Verse carries the given text, so
// a test can say what the translator's added words look like.
func translatorReader(t *testing.T, theme string, colors config.Colors, text string) Model {
	t.Helper()
	dir := t.TempDir()
	modtest.BibleAll(t, dir, "RST.SQLite3",
		map[string]string{"description": "Reina-Valera 1909"},
		genesis(),
		[]modtest.Verse{{Book: 10, Chapter: 1, Verse: 1, Text: text}})
	return configured(t, config.Config{ModuleDir: dir, ActiveTranslation: "RST", Theme: theme, Colors: colors})
}

// classifiedReader builds a reader with the given Books, in the given theme.
// Every Book carries a Verse, so the reader loads them all.
func classifiedReader(t *testing.T, theme string, books []modtest.Book) Model {
	t.Helper()
	dir := t.TempDir()
	verses := make([]modtest.Verse, 0, len(books))
	for _, b := range books {
		verses = append(verses, modtest.Verse{Book: b.Number, Chapter: 1, Verse: 1, Text: "In the beginning"})
	}
	modtest.BibleAll(t, dir, "RST.SQLite3", map[string]string{"description": "Reina-Valera 1909"}, books, verses)
	return configured(t, config.Config{ModuleDir: dir, ActiveTranslation: "RST", Theme: theme})
}

// oneBookPerClassification holds three Books of different classifications.
func oneBookPerClassification() []modtest.Book {
	return []modtest.Book{
		{Number: 10, ShortName: "Gen", LongName: "Genesis", Present: true},
		{Number: 230, ShortName: "Ps", LongName: "Psalms", Present: true},
		{Number: 500, ShortName: "Jn", LongName: "John", Present: true},
	}
}

func TestBooksTakeTheColourOfTheirClassification(t *testing.T) {
	m := classifiedReader(t, "dark", oneBookPerClassification())

	pentateuch := m.styles.book(10).Render("x")
	if got := m.styles.book(50).Render("x"); got != pentateuch {
		t.Errorf("two Books of one Classification differ: %q and %q", got, pentateuch)
	}
	for _, number := range []int{230, 500} {
		if got := m.styles.book(number).Render("x"); got == pentateuch {
			t.Errorf("Book %d is colored like the Pentateuch: %q", number, got)
		}
	}

	light := classifiedReader(t, "light", oneBookPerClassification())
	if got := light.styles.book(10).Render("x"); got == pentateuch {
		t.Errorf("the light theme reuses the dark classification colors: %q", got)
	}
}

// TestTheBookListAndTheTopBarCarryTheClassificationColour checks the color
// reaches both places a Book is shown.
func TestTheBookListAndTheTopBarCarryTheClassificationColour(t *testing.T) {
	m := classifiedReader(t, "dark", oneBookPerClassification())

	items := m.openPicker(pickerBook).pickerItems()
	if len(items) != 3 {
		t.Fatalf("the book list holds %d rows, want 3", len(items))
	}
	if got := items[0].label; got != "Gen · Genesis" {
		t.Errorf("the first row reads %q, want the Module's abbreviation, a separator, and the name", got)
	}
	for i, item := range items {
		if !strings.Contains(item.label, item.head) {
			t.Errorf("row %d colors %q, which its label %q does not carry", i, item.head, item.label)
		}
		if rendered := item.style.Render(item.short); !strings.Contains(rendered, "\x1b[") {
			t.Errorf("row %d does not color its abbreviation %q: %q", i, item.short, rendered)
		}
		if rendered := item.style.Render(item.head); !strings.Contains(rendered, "\x1b[") {
			t.Errorf("row %d (%s) is not colored: %q", i, item.label, rendered)
		}
	}
	if items[0].style.Render("x") == items[2].style.Render("x") {
		t.Error("the book list paints every Book alike")
	}

	m = m.gotoBook(10)
	if !strings.Contains(m.barView(), m.styles.bookTab(10).Render(" Genesis 1 ")) {
		t.Errorf("the top bar does not name the Book with its color\n%s", m.barView())
	}
}

func TestTheDarkAndLightThemesRenderTheWordsOfJesusDifferently(t *testing.T) {
	dark := themedReader(t, "dark", config.Colors{}).View().Content
	if !strings.Contains(dark, "\x1b[38;2;255;95;110m") {
		t.Errorf("the dark theme does not use its red for the words of Jesus\n%q", dark)
	}

	light := themedReader(t, "light", config.Colors{}).View().Content
	if !strings.Contains(light, "\x1b[38;5;160m") {
		t.Errorf("the light theme does not use its red for the words of Jesus\n%q", light)
	}
}

func TestAThemeColorOverrideWins(t *testing.T) {
	got := themedReader(t, "dark", config.Colors{Jesus: "9"}).View().Content
	if !strings.Contains(got, "\x1b[91m") {
		t.Errorf("the words-of-Jesus override was ignored\n%q", got)
	}
	if strings.Contains(got, "\x1b[38;2;255;95;110m") {
		t.Errorf("the built-in red survived the override\n%q", got)
	}
}

// The translator's added words are italic and quieter than the body text, in
// both themes.
func TestTranslatorWordsReadItalicAndQuieter(t *testing.T) {
	for _, theme := range []string{"dark", "light"} {
		m := translatorReader(t, theme, config.Colors{}, "the <i>added</i> word")
		if !strings.Contains(m.View().Content, m.styles.italics.Render("added")) {
			t.Errorf("the %s theme does not render translator words in its italics style\n%q", theme, m.View().Content)
		}
		if !m.styles.italics.GetItalic() {
			t.Errorf("the %s theme's translator words are not italic", theme)
		}
		if m.styles.italics.Render("x") == m.styles.jesus.Render("x") {
			t.Errorf("the %s theme paints translator words like the words of Jesus", theme)
		}
	}

	override := translatorReader(t, "dark", config.Colors{Italics: "9"}, "the <i>added</i> word")
	if !strings.Contains(override.styles.italics.Render("x"), ";91m") {
		t.Errorf("the italics color override was ignored: %q", override.styles.italics.Render("x"))
	}
}

// A word the translator added inside the words of Jesus keeps the quieter
// italics color: it is not the speaker's own word.
func TestTranslatorWordsInsideTheWordsOfJesusStayQuiet(t *testing.T) {
	m := translatorReader(t, "dark", config.Colors{}, "<J>a <i>added</i> word</J>")
	got := m.View().Content
	if !strings.Contains(got, m.styles.italics.Render("added")) {
		t.Errorf("the added word did not take the italics style\n%q", got)
	}
	if strings.Contains(got, m.styles.jesus.Render("added")) {
		t.Errorf("the added word was painted red like the words of Jesus\n%q", got)
	}
	if !strings.Contains(got, m.styles.jesus.Render("a ")) {
		t.Errorf("the words of Jesus around the added word lost their red\n%q", got)
	}
}

// The Commentary's bold text is the theme's white, and its verse links the
// theme's green underline; only the link color is configurable.
func TestTheCommentaryBoldAndLinksFollowTheTheme(t *testing.T) {
	dark := themedReader(t, "dark", config.Colors{})
	bold := dark.styles.commentaryBold.Render("x")
	if !strings.Contains(bold, "38;2;255;255;255") || !strings.Contains(bold, "48;2;42;42;42") {
		t.Errorf("the dark theme's commentary bold is not white on its band: %q", bold)
	}
	if got := dark.styles.link.Render("x"); !strings.Contains(got, "38;2;150;220;90") || !dark.styles.link.GetUnderline() {
		t.Errorf("the dark theme's link is not the green underline: %q", got)
	}

	overridden := themedReader(t, "light", config.Colors{Link: "9"})
	if got := overridden.styles.link.Render("x"); !strings.Contains(got, "91") {
		t.Errorf("the link color override was ignored: %q", got)
	}
}

// The hint is the modal's own footer: a quiet grey, not the body text's color.
func TestTheModalFooterTakesTheHintColor(t *testing.T) {
	m := themedReader(t, "dark", config.Colors{Hint: "9"})
	m, _ = send(t, m, press('t', "t", 0))

	got := m.modalView()
	if !strings.Contains(got, "enter to select") {
		t.Fatalf("the modal footer is missing\n%q", got)
	}
	if !strings.Contains(got, "\x1b[91m") {
		t.Errorf("the modal footer did not take the hint color\n%q", got)
	}
}

func TestTheReaderKeepsThemeAndLanguageWhenItSavesState(t *testing.T) {
	dir := t.TempDir()
	modtest.BibleAll(t, dir, "RST.SQLite3", map[string]string{"description": "Reina-Valera 1909"},
		genesis(), []modtest.Verse{{Book: 10, Chapter: 1, Verse: 1, Text: "En el principio"}})

	configDir := t.TempDir()
	cfg := config.Config{ModuleDir: dir, ActiveTranslation: "RST", Theme: "light", Language: "es"}
	if err := config.Save(configDir, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	m := newReader(t, configDir, dir)
	send(t, m, press('q', "q", 0)) // the reader persists its state on quit

	saved, err := config.Load(configDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if saved.Theme != "light" || saved.Language != "es" {
		t.Errorf("saving state lost the theme or the language: %+v", saved)
	}
}
