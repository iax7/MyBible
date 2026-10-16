package tui

import (
	"charm.land/lipgloss/v2"

	"github.com/iax7/mybible/internal/config"
	"github.com/iax7/mybible/internal/module"
)

// Palette is the full set of colors a theme is built from: field for field the
// reader's config colors. An empty field means the terminal's default color.
type Palette = config.Colors

// The built-in themes, selected by name in config.
var (
	darkPalette = Palette{
		Bar:       "24",
		BarText:   "252",
		Key:       "39",
		Label:     "245",
		Reference: "252",
		Warning:   "214",
		Jesus:     "196",
		VerseNum:  "240",
		CursorNum: "39",
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
	return p
}

// styles are every style the reader draws with, derived from one theme so no
// file needs its own color literal.
type styles struct {
	bar       lipgloss.Style
	key       lipgloss.Style
	label     lipgloss.Style
	reference lipgloss.Style
	warning   lipgloss.Style
	invalid   lipgloss.Style
	help      lipgloss.Style
	italics   lipgloss.Style
	jesus     lipgloss.Style
	verseNum  lipgloss.Style
	cursorNum lipgloss.Style
	bookColor bookPalette
}

// newStyles builds the rendered styles from a Palette and its book colors.
func newStyles(p Palette, bookColors bookPalette) styles {
	return styles{
		bar:       lipgloss.NewStyle().Background(lipgloss.Color(p.Bar)).Foreground(lipgloss.Color(p.BarText)),
		key:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.Key)),
		label:     lipgloss.NewStyle().Foreground(lipgloss.Color(p.Label)),
		reference: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.Reference)),
		warning:   lipgloss.NewStyle().Foreground(lipgloss.Color(p.Warning)),
		invalid:   lipgloss.NewStyle().Foreground(lipgloss.Color(p.Jesus)), // the theme's red, for a number the translation has no Verse or Chapter for
		help:      lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 3),
		italics:   lipgloss.NewStyle().Italic(true),
		jesus:     lipgloss.NewStyle().Foreground(lipgloss.Color(p.Jesus)),
		verseNum:  lipgloss.NewStyle().Foreground(lipgloss.Color(p.VerseNum)),
		cursorNum: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.CursorNum)),
		bookColor: bookColors,
	}
}

// book styles a Book name by the Classification of the Book with this number.
func (s styles) book(number int) lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(s.bookColor[module.ClassificationOf(number)]))
}
