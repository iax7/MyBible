package tui

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/iax7/mybible/internal/config"
	"github.com/iax7/mybible/internal/module"
)

// Palette is the full set of colors a theme is built from: field for field the
// reader's config colors. An empty field means the terminal's default color.
type Palette = config.Colors

// The built-in themes, selected by name in config. The dark theme is zellij's
// built-in default palette (see ADR-0003), so the reader sits quietly inside the
// terminal its maintainer already reads in.
var (
	darkPalette = Palette{
		Bar:       "#0B0B0B",
		BarText:   "#D7D7D7",
		Key:       "#96DC5A",
		Label:     "#D7D7D7",
		Reference: "#D7D7D7",
		Warning:   "#FF9146",
		Jesus:     "#FF5F6E",
		VerseNum:  "#898989",
		CursorNum: "#96DC5A",
		Hint:      "#6C6C6C",
		Field:     "#2A2A2A",
	}
	lightPalette = Palette{
		Bar:       "189",
		BarText:   "235",
		Key:       "25",
		Label:     "240",
		Reference: "236",
		Warning:   "130",
		Jesus:     "160",
		VerseNum:  "245",
		CursorNum: "25",
		Hint:      "248",
		Field:     "253",
	}
)

// bookPalette is the color of every Book classification.
type bookPalette map[module.Classification]string

// The classification colors of the two built-in themes, taken from MyBible's
// own book selector: the bright set reads on a dark terminal, the muted one on
// a light one.
var (
	darkBookColors = bookPalette{
		module.Pentateuch:      "#ccccff",
		module.HistoricalBooks: "#ffcc99",
		module.PoeticBooks:     "#66ff99",
		module.MajorProphets:   "#ff9fb4",
		module.MinorProphets:   "#ffff99",
		module.Gospels:         "#ff6600",
		module.Acts:            "#00ffff",
		module.PaulineEpistles: "#ffff00",
		module.GeneralEpistles: "#00ff00",
		module.Revelation:      "#ff7c80",
		module.NonCanonical:    "#c0c0c0",
	}
	lightBookColors = bookPalette{
		module.Pentateuch:      "#7b7ba6",
		module.HistoricalBooks: "#a67b50",
		module.PoeticBooks:     "#24a650",
		module.MajorProphets:   "#a65566",
		module.MinorProphets:   "#a6a650",
		module.Gospels:         "#a64200",
		module.Acts:            "#00a6a6",
		module.PaulineEpistles: "#a6a600",
		module.GeneralEpistles: "#00a600",
		module.Revelation:      "#a6373a",
		module.NonCanonical:    "#7d7d7d",
	}
)

// bookColorsFor resolves the classification colors of the named theme.
func bookColorsFor(theme string) bookPalette {
	if theme == "light" {
		return lightBookColors
	}
	return darkBookColors
}

// paletteFor resolves the theme named in config, applies the color overrides,
// and falls back to the dark theme for an unknown name.
func paletteFor(theme string, overrides config.Colors) Palette {
	p := darkPalette
	if theme == "light" {
		p = lightPalette
	}
	set := func(dst *string, value string) {
		if value != "" {
			*dst = value
		}
	}
	set(&p.Bar, overrides.Bar)
	set(&p.BarText, overrides.BarText)
	set(&p.Key, overrides.Key)
	set(&p.Label, overrides.Label)
	set(&p.Reference, overrides.Reference)
	set(&p.Warning, overrides.Warning)
	set(&p.Jesus, overrides.Jesus)
	set(&p.VerseNum, overrides.VerseNum)
	set(&p.CursorNum, overrides.CursorNum)
	set(&p.Hint, overrides.Hint)
	set(&p.Field, overrides.Field)
	return p
}

// styles are every style the reader draws with, derived from one theme so no
// file needs its own color literal.
type styles struct {
	bar        lipgloss.Style
	barColor   string // the bar's own color: a filled Book tab and the Chapter heading draw their text in it
	key        lipgloss.Style
	label      lipgloss.Style
	reference  lipgloss.Style
	warning    lipgloss.Style
	invalid    lipgloss.Style
	frame      lipgloss.Style
	italics    lipgloss.Style
	jesus      lipgloss.Style
	verseNum   lipgloss.Style
	cursorNum  lipgloss.Style
	hint       lipgloss.Style
	field      lipgloss.Style // the Reference modal's input line, a filled band that marks where typing lands
	popup      lipgloss.Style // the Reference modal's completion menu
	popupKey   lipgloss.Style // the menu's highlighted row
	subheading lipgloss.Style
	bookColor  bookPalette
}

// newStyles builds the rendered styles from a Palette and its book colors.
func newStyles(p Palette, bookColors bookPalette) styles {
	return styles{
		bar:        lipgloss.NewStyle().Background(lipgloss.Color(p.Bar)).Foreground(lipgloss.Color(p.BarText)),
		barColor:   p.Bar,
		key:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.Key)),
		label:      lipgloss.NewStyle().Foreground(lipgloss.Color(p.Label)),
		reference:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.Reference)),
		warning:    lipgloss.NewStyle().Foreground(lipgloss.Color(p.Warning)),
		invalid:    lipgloss.NewStyle().Foreground(lipgloss.Color(p.Jesus)).Background(lipgloss.Color(p.Field)), // the theme's red on the input's band, for a number the translation has no Verse or Chapter for
		frame:      lipgloss.NewStyle().Foreground(lipgloss.Color(p.Key)),
		italics:    lipgloss.NewStyle().Italic(true),
		jesus:      lipgloss.NewStyle().Foreground(lipgloss.Color(p.Jesus)),
		verseNum:   lipgloss.NewStyle().Foreground(lipgloss.Color(p.VerseNum)),
		cursorNum:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.CursorNum)),
		hint:       lipgloss.NewStyle().Foreground(lipgloss.Color(p.Hint)),
		field:      lipgloss.NewStyle().Background(lipgloss.Color(p.Field)).Foreground(lipgloss.Color(p.Label)),
		popup:      lipgloss.NewStyle().Background(lipgloss.Color(p.Bar)).Foreground(lipgloss.Color(p.BarText)),
		popupKey:   lipgloss.NewStyle().Background(lipgloss.Color(p.Bar)).Foreground(lipgloss.Color(p.Key)).Bold(true),
		subheading: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.Key)),
		bookColor:  bookColors,
	}
}

// book styles a Book name by the Classification of the Book with this number.
func (s styles) book(number int) lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(s.bookColor[module.ClassificationOf(number)]))
}

// bookTab styles a Book name the way the top bar and the Chapter heading fill a
// selected tab: the Book's classification color behind the interface's own.
func (s styles) bookTab(number int) lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).
		Background(lipgloss.Color(s.bookColor[module.ClassificationOf(number)])).
		Foreground(lipgloss.Color(s.barColor))
}

// frame draws a modal the way zellij frames a dialog: a box whose top edge
// carries the title on the left and a key hint at its right end.
func (s styles) frameBox(title, hint, body string) string {
	lines := strings.Split(strings.TrimRight(body, "\n"), "\n")

	width := lipgloss.Width(title) + lipgloss.Width(hint) + 6
	for _, line := range lines {
		if w := lipgloss.Width(line) + 4; w > width { // two cells of padding on each side
			width = w
		}
	}

	var b strings.Builder
	edge := func(text string) { b.WriteString(s.frame.Render(text)) }
	fill := width - lipgloss.Width(title) - lipgloss.Width(hint) - 6
	edge("╭─ " + title + " " + strings.Repeat("─", fill) + " " + hint + " ─╮")
	b.WriteString("\n")
	for _, line := range lines {
		edge("│")
		b.WriteString("  " + line + strings.Repeat(" ", width-lipgloss.Width(line)-4) + "  ")
		edge("│")
		b.WriteString("\n")
	}
	edge("╰" + strings.Repeat("─", width) + "╯")
	return b.String()
}
