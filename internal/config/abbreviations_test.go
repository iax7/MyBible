package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestAbbreviationsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	want := Abbreviations{
		Description:  "Preferred abbreviations (from NBLA'05)",
		Language:     "es",
		ByBookNumber: map[int]string{10: "Gn", 20: "Éx", 100: "2S"},
	}
	if err := SaveAbbreviations(dir, want); err != nil {
		t.Fatalf("SaveAbbreviations: %v", err)
	}

	got, err := LoadAbbreviations(dir)
	if err != nil {
		t.Fatalf("LoadAbbreviations: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("LoadAbbreviations = %+v, want %+v", got, want)
	}
}

func TestSaveAbbreviationsWritesMyBiblesFields(t *testing.T) {
	dir := t.TempDir()
	set := Abbreviations{
		Description:  "Preferred abbreviations (from NBLA'05)",
		Language:     "es",
		ByBookNumber: map[int]string{10: "Gn", 20: "Éx", 100: "2S"},
	}
	if err := SaveAbbreviations(dir, set); err != nil {
		t.Fatalf("SaveAbbreviations: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, abbreviationsFile))
	if err != nil {
		t.Fatal(err)
	}
	var record abbreviationsRecord
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("the saved file is not JSON: %v", err)
	}
	if record.FileName != abbreviationsFile {
		t.Errorf("fileName = %q, want %q", record.FileName, abbreviationsFile)
	}
	if record.Description != set.Description || record.Language != set.Language {
		t.Errorf("description/language = %q/%q, want %q/%q", record.Description, record.Language, set.Description, set.Language)
	}
	if got := record.AbbreviationByBookNumber["10"]; got != "Gn" {
		t.Errorf("abbreviationByBookNumber[10] = %q, want %q", got, "Gn")
	}

	// The Books stay in number order, not the lexicographic order of their keys.
	last := -1
	for _, key := range []string{`"10":`, `"20":`, `"100":`} {
		at := strings.Index(string(data), key)
		if at < 0 || at < last {
			t.Errorf("the saved Books are out of number order:\n%s", data)
			break
		}
		last = at
	}
}

func TestLoadAbbreviationsReadsMyBiblesFormat(t *testing.T) {
	dir := t.TempDir()
	raw := `{
  "fileName": "NBLA'05.abbreviations.json",
  "description": "A set from Android",
  "language": "es",
  "abbreviationByBookNumber": {
    "10": "Gn",
    "20": " Éx ",
    "30": "",
    "not a number": "X",
    "-5": "Y"
  }
}`
	if err := os.WriteFile(filepath.Join(dir, abbreviationsFile), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := LoadAbbreviations(dir)
	if err != nil {
		t.Fatalf("LoadAbbreviations: %v", err)
	}
	if got.Description != "A set from Android" || got.Language != "es" {
		t.Errorf("description/language = %q/%q, want the MyBible values", got.Description, got.Language)
	}
	want := map[int]string{10: "Gn", 20: "Éx"}
	if !reflect.DeepEqual(got.ByBookNumber, want) {
		t.Errorf("ByBookNumber = %v, want %v", got.ByBookNumber, want)
	}
}

func TestLoadAbbreviationsReturnsAnEmptySetWhenTheFileIsMissing(t *testing.T) {
	got, err := LoadAbbreviations(t.TempDir())
	if err != nil {
		t.Fatalf("LoadAbbreviations: %v", err)
	}
	if len(got.ByBookNumber) != 0 || got.Description != "" || got.Language != "" {
		t.Errorf("LoadAbbreviations = %+v, want an empty set", got)
	}
}

func TestLoadAbbreviationsReportsAMalformedFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, abbreviationsFile), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadAbbreviations(dir); err == nil {
		t.Error("LoadAbbreviations accepted a malformed file")
	}
}
