package module

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/iax7/mybible/internal/modtest"
)

func TestDiscoverCommentariesReadsTheTitleAndLanguage(t *testing.T) {
	dir := t.TempDir()
	modtest.Commentaries(t, dir, "MAC.commentaries.SQLite3", map[string]string{
		"description": "MacArthur Study Bible",
		"language":    "en",
	}, nil)

	commentaries, warnings := DiscoverCommentaries(dir)

	if len(warnings) != 0 {
		t.Fatalf("DiscoverCommentaries warned: %v", warnings)
	}
	if len(commentaries) != 1 {
		t.Fatalf("DiscoverCommentaries found %d Commentaries, want 1", len(commentaries))
	}
	got := commentaries[0]
	if got.Abbreviation != "MAC" {
		t.Errorf("Abbreviation = %q, want MAC", got.Abbreviation)
	}
	if got.Title != "MacArthur Study Bible" {
		t.Errorf("Title = %q, want MacArthur Study Bible", got.Title)
	}
	if got.Language != "en" {
		t.Errorf("Language = %q, want en", got.Language)
	}
}

func TestDiscoverCommentariesIgnoresOtherModuleTypes(t *testing.T) {
	dir := t.TempDir()
	modtest.Bible(t, dir, "RST.SQLite3", map[string]string{"description": "Reina-Valera"})
	modtest.Create(t, dir, "Strong.dictionary.SQLite3",
		"CREATE TABLE info (name TEXT, value TEXT)",
		"CREATE TABLE dictionary (topic TEXT, definition TEXT NOT NULL)")
	modtest.Commentaries(t, dir, "MAC.commentaries.SQLite3", map[string]string{"description": "MacArthur"}, nil)

	commentaries, warnings := DiscoverCommentaries(dir)

	if len(warnings) != 0 {
		t.Fatalf("DiscoverCommentaries warned about ignored Modules: %v", warnings)
	}
	if len(commentaries) != 1 || commentaries[0].Abbreviation != "MAC" {
		t.Fatalf("DiscoverCommentaries = %+v, want only MAC", commentaries)
	}
}

func TestDiscoverCommentariesWarnsAboutACorruptModuleAndKeepsTheRest(t *testing.T) {
	dir := t.TempDir()
	modtest.Commentaries(t, dir, "MAC.commentaries.SQLite3", map[string]string{"description": "MacArthur"}, nil)
	if err := os.WriteFile(filepath.Join(dir, "Broken.commentaries.SQLite3"), []byte("not a database"), 0o644); err != nil {
		t.Fatal(err)
	}

	commentaries, warnings := DiscoverCommentaries(dir)

	if len(commentaries) != 1 {
		t.Errorf("DiscoverCommentaries found %d Commentaries, want the one valid Module", len(commentaries))
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "Broken.commentaries.SQLite3") {
		t.Fatalf("warnings = %v, want one naming the corrupt Module", warnings)
	}
}

func TestCommentaryRowsReadsTheChapterOnDemand(t *testing.T) {
	dir := t.TempDir()
	modtest.Commentaries(t, dir, "MAC.commentaries.SQLite3", map[string]string{"description": "MacArthur"},
		[]modtest.Commentary{
			{Book: 10, ChapterFrom: 1, VerseFrom: 1, Text: "one"},
			{Book: 10, ChapterFrom: 1, VerseFrom: 2, ChapterTo: 1, VerseTo: 4, Text: "two-four"},
			{Book: 10, ChapterFrom: 1, VerseFrom: 0, Text: "overview"},
			{Book: 10, ChapterFrom: 2, VerseFrom: 1, Text: "other chapter"},
			{Book: 790, ChapterFrom: 1, VerseFrom: 1, Text: "legacy"},
			{Book: 10, ChapterFrom: 1, VerseFrom: 5, NullEnd: true, Text: "null end"},
		})

	commentaries, _ := DiscoverCommentaries(dir)
	if len(commentaries) != 1 {
		t.Fatalf("DiscoverCommentaries found %d, want 1", len(commentaries))
	}

	rows, err := commentaries[0].Rows(10, 1)
	if err != nil {
		t.Fatalf("Rows: %v", err)
	}
	got := make([]string, 0, len(rows))
	for _, row := range rows {
		got = append(got, row.Text)
	}
	if want := []string{"overview", "one", "two-four", "null end"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Rows(10, 1) = %v, want %v", got, want)
	}

	rows, err = commentaries[0].Rows(10, 2)
	if err != nil {
		t.Fatalf("Rows: %v", err)
	}
	if len(rows) != 1 || rows[0].Text != "other chapter" {
		t.Errorf("Rows(10, 2) = %+v, want only the Chapter 2 row", rows)
	}

	rows, err = commentaries[0].Rows(145, 1) // 790 is the legacy number for Manasseh
	if err != nil {
		t.Fatalf("Rows: %v", err)
	}
	if len(rows) != 1 || rows[0].Text != "legacy" {
		t.Errorf("Rows(145, 1) = %+v, want the legacy-numbered row", rows)
	}
}

func TestCommentaryRowCovers(t *testing.T) {
	cases := []struct {
		name    string
		row     CommentaryRow
		chapter int
		verse   int
		want    bool
	}{
		{"single verse", CommentaryRow{ChapterFrom: 1, VerseFrom: 3}, 1, 3, true},
		{"single verse before", CommentaryRow{ChapterFrom: 1, VerseFrom: 3}, 1, 2, false},
		{"range inside", CommentaryRow{ChapterFrom: 1, VerseFrom: 2, ChapterTo: 1, VerseTo: 5}, 1, 4, true},
		{"range before", CommentaryRow{ChapterFrom: 1, VerseFrom: 2, ChapterTo: 1, VerseTo: 5}, 1, 1, false},
		{"range after", CommentaryRow{ChapterFrom: 1, VerseFrom: 2, ChapterTo: 1, VerseTo: 5}, 1, 6, false},
		{"chapter level", CommentaryRow{ChapterFrom: 1}, 1, 20, true},
		{"chapter level other chapter", CommentaryRow{ChapterFrom: 1}, 2, 1, false},
		{"to end of chapter", CommentaryRow{ChapterFrom: 1, VerseFrom: 2, ChapterTo: 1}, 1, 99, true},
		{"across chapters", CommentaryRow{ChapterFrom: 1, VerseFrom: 2, ChapterTo: 2, VerseTo: 3}, 2, 3, true},
		{"across chapters after", CommentaryRow{ChapterFrom: 1, VerseFrom: 2, ChapterTo: 2, VerseTo: 3}, 2, 4, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.row.Covers(tc.chapter, tc.verse); got != tc.want {
				t.Errorf("Covers(%d, %d) = %v, want %v", tc.chapter, tc.verse, got, tc.want)
			}
		})
	}
}

func TestParseCommentary(t *testing.T) {
	cases := []struct {
		name string
		text string
		want []Run
	}{
		{
			name: "paragraphs and italics",
			text: "<p>one <i>two</i></p><p>three &amp; four</p>",
			want: []Run{
				{Text: "one "},
				{Text: "two", Italics: true},
				{Text: "\n\n"},
				{Text: "three & four"},
			},
		},
		{
			name: "line break and list",
			text: "a<br/>b<ul><li>item</li></ul>",
			want: []Run{
				{Text: "a"},
				{Text: "\n"},
				{Text: "b"},
				{Text: "\n\n"},
				{Text: "• item"},
			},
		},
		{
			name: "span italics",
			text: `<span class="it">faz</span> normal`,
			want: []Run{{Text: "faz", Italics: true}, {Text: " normal"}},
		},
		{
			name: "bold tags",
			text: `<b>one</b> <strong>two</strong>`,
			want: []Run{{Text: "one", Bold: true}, {Text: " "}, {Text: "two", Bold: true}},
		},
		{
			name: "bold span",
			text: `<span style="font-size:18pt;font-weight: bold;">title</span>`,
			want: []Run{{Text: "title", Bold: true}},
		},
		{
			name: "bold and italics",
			text: `<b><i>both</i></b>`,
			want: []Run{{Text: "both", Italics: true, Bold: true}},
		},
		{
			name: "links keep their text",
			text: `<a href='B:540 4:6'>2 Co. 4.6</a>`,
			want: []Run{{Text: "2 Co. 4.6", Link: true}},
		},
		{
			name: "only verse links are links",
			text: `<a class='S' href='S:Thompson 2596'>2596</a> <a href='C:@230 120:1'>nota</a>`,
			want: []Run{{Text: "2596"}, {Text: " "}, {Text: "nota"}},
		},
		{
			name: "a link inside bold keeps both",
			text: `<b><a href='B:10 1:1'>Gén 1:1</a></b>`,
			want: []Run{{Text: "Gén 1:1", Bold: true, Link: true}},
		},
		{
			name: "entities",
			text: "&quot;quoted&quot; &nbsp; &lt;literal&gt;",
			want: []Run{{Text: `"quoted"   <literal>`}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ParseCommentary(tc.text); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ParseCommentary(%q) = %+v, want %+v", tc.text, got, tc.want)
			}
		})
	}
}
