package tui

import (
	"charm.land/lipgloss/v2"

	"github.com/iax7/mybible/internal/config"
)

// Palette is the full set of colors a theme is built from: field for field the
// reader's config colors. An empty field means the terminal's default color.
type Palette = config.Colors

// The built-in themes, selected by name in config.
var (
	darkPalette = Palette{
		Bar:       "236",
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
		Bar:       "254",
		BarText:   "236",
		Key:       "25",
		Label:     "240",
		Reference: "236",
		Warning:   "130",
		Jesus:     "160",
		VerseNum:  "245",
		CursorNum: "25",
	}
)

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

// styles are every style the reader draws with, derived from one Palette so no
// file needs its own color literal.
type styles struct {
	bar       lipgloss.Style
	key       lipgloss.Style
	label     lipgloss.Style
	reference lipgloss.Style
	warning   lipgloss.Style
	help      lipgloss.Style
	italics   lipgloss.Style
	jesus     lipgloss.Style
	verseNum  lipgloss.Style
	cursorNum lipgloss.Style
}

// newStyles builds the rendered styles from a Palette.
func newStyles(p Palette) styles {
	return styles{
		bar:       lipgloss.NewStyle().Background(lipgloss.Color(p.Bar)).Foreground(lipgloss.Color(p.BarText)),
		key:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.Key)),
		label:     lipgloss.NewStyle().Foreground(lipgloss.Color(p.Label)),
		reference: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.Reference)),
		warning:   lipgloss.NewStyle().Foreground(lipgloss.Color(p.Warning)),
		help:      lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 3),
		italics:   lipgloss.NewStyle().Italic(true),
		jesus:     lipgloss.NewStyle().Foreground(lipgloss.Color(p.Jesus)),
		verseNum:  lipgloss.NewStyle().Foreground(lipgloss.Color(p.VerseNum)),
		cursorNum: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.CursorNum)),
	}
}
