package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/iax7/mybible/internal/config"
	"github.com/iax7/mybible/internal/modtest"
)

// spanishBooks are Books named the way a Spanish Translation names them: the
// Canonical name stays the Book's identity, while "Génesis" is what the reader
// shows and completes.
func spanishBooks() []modtest.Book {
	return []modtest.Book{
		{Number: 10, ShortName: "Gn", LongName: "Génesis", Present: true},
		{Number: 20, ShortName: "Éx", LongName: "Éxodo", Present: true},
		{Number: 500, ShortName: "Jn", LongName: "Juan", Present: true},
	}
}

func spanishVerses() []modtest.Verse {
	return []modtest.Verse{
		{Book: 10, Chapter: 1, Verse: 1, Text: "En el principio creó Dios los cielos"},
		{Book: 10, Chapter: 1, Verse: 2, Text: "La tierra estaba desordenada"},
		{Book: 10, Chapter: 2, Verse: 1, Text: "Así quedaron acabados"},
		{Book: 10, Chapter: 3, Verse: 1, Text: "La serpiente era astuta"},
		{Book: 20, Chapter: 1, Verse: 1, Text: "Estos son los nombres"},
		{Book: 500, Chapter: 3, Verse: 16, Text: "Porque de tal manera amó Dios al mundo"},
	}
}

// spanishReader builds a reader over a Spanish Module and interface, so both the
// Book names and the chrome around them are read in the Translation's language.
func spanishReader(t *testing.T) Model {
	t.Helper()

	dir := t.TempDir()
	path := modtest.BibleAll(t, dir, "NBLA.SQLite3", map[string]string{
		"description": "Nueva Biblia de las Américas",
		"language":    "es",
	}, spanishBooks(), spanishVerses())
	modtest.Stories(t, path, []modtest.Subheading{
		{Book: 10, Chapter: 1, Verse: 1, Text: "La creación <x>b:20 1:1</x>"},
	})

	configDir := t.TempDir()
	if err := config.Save(configDir, config.Config{
		ModuleDir:         dir,
		ActiveTranslation: "NBLA",
		Language:          "es",
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	return newReader(t, configDir, dir)
}

func TestASpanishModuleIsNamedInItsOwnLanguage(t *testing.T) {
	m := spanishReader(t)
	if got := render(t, m); !strings.Contains(got, "Génesis 1") {
		t.Fatalf("the top bar does not name the Book the way the Translation does\n%s", got)
	}

	m, _ = send(t, m, press('b', "b", 0))
	got := modalText(t, m)
	for _, want := range []string{"Gn", "Génesis", "Éx", "Éxodo"} {
		if !strings.Contains(got, want) {
			t.Errorf("the Book list is missing %q\n%s", want, got)
		}
	}
	if strings.Contains(got, "Genesis") || strings.Contains(got, "Exodus") {
		t.Errorf("the Book list shows a Canonical name the Module does not use\n%s", got)
	}
}

func TestTheBookListFiltersByNameInEitherLanguage(t *testing.T) {
	cases := []struct{ typed, want, absent string }{
		{"genesis", "Génesis", "Éxodo"}, // the Canonical name still finds the Book
		{"éxodo", "Éxodo", "Génesis"},   // and so does the Module's own name
		{"Gn", "Génesis", "Éxodo"},      // and its abbreviation
	}
	for _, c := range cases {
		m := spanishReader(t)
		m, _ = send(t, m, press('b', "b", 0))
		m = typeKey(t, m, c.typed)

		got := modalText(t, m)
		if !strings.Contains(got, c.want) {
			t.Errorf("typing %q did not list %q\n%s", c.typed, c.want, got)
		}
		if strings.Contains(got, c.absent) {
			t.Errorf("typing %q also listed %q\n%s", c.typed, c.absent, got)
		}
	}
}

func TestCompletionWritesTheModulesName(t *testing.T) {
	m := spanishReader(t)
	m, _ = send(t, m, press('v', "v", 0))
	m = typeKey(t, m, "gén")
	m, _ = send(t, m, press(tea.KeyTab, "", 0))

	if got := render(t, m); !strings.Contains(got, "❯ Génesis ") {
		t.Fatalf("Tab wrote a name the Translation does not show\n%s", got)
	}

	m = typeKey(t, m, "2")
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))
	if got := on(m); got != "Genesis 2:1" {
		t.Errorf("the reader is on %q, want Genesis 2:1", got)
	}
}

func TestTheCanonicalNameStillNamesTheBook(t *testing.T) {
	m := spanishReader(t)
	m = goTo(t, m, "Genesis 2:1")

	if got := on(m); got != "Genesis 2:1" {
		t.Errorf("the reader is on %q after typing the Canonical name, want Genesis 2:1", got)
	}
}

func TestSearchNamesTheBookTheWayTheTranslationDoes(t *testing.T) {
	m := spanishReader(t)
	m, _ = send(t, m, press('/', "/", 0))
	m = typeKey(t, m, "principio")

	if got := modalText(t, m); !strings.Contains(got, "Génesis 1:1") {
		t.Errorf("the Search result does not use the Translation's Book name\n%s", got)
	}
}

func TestHistoryNamesTheBookTheWayTheTranslationDoes(t *testing.T) {
	m := spanishReader(t)
	m = goTo(t, m, "Génesis 2:1")
	m, _ = send(t, m, press('H', "H", 0))

	got := modalText(t, m)
	if !strings.Contains(got, "Génesis 2:1") {
		t.Errorf("the History entry does not use the Translation's Book name\n%s", got)
	}
	if strings.Contains(got, "Genesis 2:1") {
		t.Errorf("the History entry shows the Canonical name\n%s", got)
	}
}

func TestSubheadingLinksNameTheBookTheWayTheTranslationDoes(t *testing.T) {
	m := spanishReader(t)
	got := render(t, m)
	if !strings.Contains(got, "Éxodo 1:1") {
		t.Errorf("the Subheading link does not use the Translation's Book name\n%s", got)
	}
	if strings.Contains(got, "Exodus 1:1") {
		t.Errorf("the Subheading link shows the Canonical name\n%s", got)
	}
}

func TestTheReferenceModalSpeaksTheModulesLanguage(t *testing.T) {
	m := spanishReader(t)
	m, _ = send(t, m, press('v', "v", 0))

	if got := render(t, m); !strings.Contains(got, "Por ejemplo: Génesis, Génesis 1, Génesis 1:1") {
		t.Fatalf("the modal does not show the shapes to type in the Module's own words\n%s", got)
	}

	m = typeKey(t, m, "Zz 2")
	m, _ = send(t, m, press(tea.KeyEnter, "", 0))
	if got := render(t, m); !strings.Contains(got, "Eso no es una referencia. Prueba con Génesis 1:1.") {
		t.Errorf("the rejection does not suggest a Reference in the Module's own words\n%s", got)
	}
}
