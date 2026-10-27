package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// abbreviationsFile is the reader's Preferred abbreviation set. The name and
// JSON shape are MyBible's, so the same file can be copied to the Android app.
const abbreviationsFile = "preferred.abbreviations.json"

// Abbreviations is the reader's Preferred abbreviation set: one abbreviation
// per Book number, and the language the set applies to (empty for every
// language).
type Abbreviations struct {
	Description  string
	Language     string
	ByBookNumber map[int]string
}

// abbreviationsRecord is MyBible's desired-abbreviation JSON shape.
type abbreviationsRecord struct {
	FileName                 string            `json:"fileName"`
	Description              string            `json:"description"`
	Language                 string            `json:"language"`
	AbbreviationByBookNumber map[string]string `json:"abbreviationByBookNumber"`
}

// LoadAbbreviations reads the Preferred abbreviation set from dir. A missing
// file yields an empty set and no error, like Load.
func LoadAbbreviations(dir string) (Abbreviations, error) {
	if dir == "" {
		return Abbreviations{}, errors.New("no config directory")
	}
	path := filepath.Join(dir, abbreviationsFile)
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Abbreviations{}, nil
	}
	if err != nil {
		return Abbreviations{}, fmt.Errorf("read %s: %w", path, err)
	}

	var record abbreviationsRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return Abbreviations{}, fmt.Errorf("read %s: %w", path, err)
	}
	set := Abbreviations{
		Description:  record.Description,
		Language:     strings.TrimSpace(record.Language),
		ByBookNumber: make(map[int]string, len(record.AbbreviationByBookNumber)),
	}
	for key, value := range record.AbbreviationByBookNumber {
		number, err := strconv.Atoi(strings.TrimSpace(key))
		abbreviation := strings.TrimSpace(value)
		if err != nil || number <= 0 || abbreviation == "" {
			continue
		}
		set.ByBookNumber[number] = abbreviation
	}
	return set, nil
}

// SaveAbbreviations writes the Preferred abbreviation set to dir in MyBible's
// desired-abbreviation JSON format, with the Books in number order so the file
// is readable when edited by hand.
func SaveAbbreviations(dir string, set Abbreviations) error {
	if dir == "" {
		return errors.New("no config directory")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	numbers := make([]int, 0, len(set.ByBookNumber))
	for number := range set.ByBookNumber {
		numbers = append(numbers, number)
	}
	sort.Ints(numbers)

	var out bytes.Buffer
	out.WriteString("{\n")
	fmt.Fprintf(&out, "  \"fileName\": %s,\n", jsonString(abbreviationsFile))
	fmt.Fprintf(&out, "  \"description\": %s,\n", jsonString(set.Description))
	fmt.Fprintf(&out, "  \"language\": %s,\n", jsonString(set.Language))
	out.WriteString("  \"abbreviationByBookNumber\": {")
	for i, number := range numbers {
		if i > 0 {
			out.WriteByte(',')
		}
		fmt.Fprintf(&out, "\n    %s: %s", jsonString(strconv.Itoa(number)), jsonString(set.ByBookNumber[number]))
	}
	if len(numbers) > 0 {
		out.WriteString("\n  ")
	}
	out.WriteString("}\n}\n")

	path := filepath.Join(dir, abbreviationsFile)
	if err := os.WriteFile(path, out.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// jsonString quotes text as a JSON string. Marshalling a string cannot fail.
func jsonString(text string) string {
	quoted, _ := json.Marshal(text)
	return string(quoted)
}
