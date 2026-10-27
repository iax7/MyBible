package tui

import (
	"fmt"
	"strings"

	"github.com/iax7/mybible/internal/config"
	"github.com/iax7/mybible/internal/module"
)

// preferredAbbreviation is the abbreviation the reader shows for the Book: the
// Preferred one when the reader's set is enabled and applies to the Active
// translation, otherwise the Module's own short name.
func (m Model) preferredAbbreviation(b module.Book) string {
	if !m.cfg.PreferredAbbreviations {
		return b.ShortName
	}
	if language := strings.TrimSpace(m.abbreviations.Language); language != "" {
		translation, ok := m.activeTranslation()
		if !ok || !strings.EqualFold(language, strings.TrimSpace(translation.Language)) {
			return b.ShortName
		}
	}
	if abbreviation := strings.TrimSpace(m.abbreviations.ByBookNumber[b.Number]); abbreviation != "" {
		return abbreviation
	}
	return b.ShortName
}

// seedPreferredAbbreviations copies the chosen Translation's Book abbreviations
// into the reader's Preferred set and writes it beside config.toml.
func (m Model) seedPreferredAbbreviations(index int) Model {
	if index < 0 || index >= len(m.translations) {
		return m
	}
	source := m.translations[index]
	books, err := source.Books()
	if err != nil {
		m.warnings = append(m.warnings, fmt.Sprintf("%s: %v", source.Abbreviation, err))
		return m
	}

	set := config.Abbreviations{ByBookNumber: make(map[int]string, len(books))}
	for _, b := range books {
		if abbreviation := strings.TrimSpace(b.ShortName); abbreviation != "" {
			set.ByBookNumber[b.Number] = abbreviation
		}
	}
	if len(set.ByBookNumber) == 0 {
		m.warnings = append(m.warnings, fmt.Sprintf("preferred abbreviations: %s has no Book abbreviations", source.Abbreviation))
		return m
	}

	set.Description = "Preferred abbreviations"
	if title := strings.TrimSpace(source.Title); title != "" {
		set.Description += " (from " + title + ")"
	}
	if err := config.SaveAbbreviations(m.configDir, set); err != nil {
		m.warnings = append(m.warnings, "preferred abbreviations: "+err.Error())
	}
	m.abbreviations = set
	return m
}
