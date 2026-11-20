package mtp

import "testing"

func TestParseListingDropsBlankLinesAndTheDotEntry(t *testing.T) {
	got := parseListing("\nRVR60.SQLite3\n\n-downloads.cache.SQLite3\n\n.\n\nuser\n")
	want := []string{"RVR60.SQLite3", "-downloads.cache.SQLite3", "user"}
	if len(got) != len(want) {
		t.Fatalf("parseListing returned %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("parseListing[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSkippedLeavesOutTheAppsOwnFiles(t *testing.T) {
	cases := map[string]bool{
		"RVR60.SQLite3":                           false,
		"RV60-s.subheadings.SQLite3":              false,
		"NBLA'05.commentaries.SQLite3":            false,
		"TSK-x.crossreferences.SQLite3":           false,
		"RMMC-p.plan.SQLite3":                     false,
		"ThemeArt.bundle.json":                    false,
		"user":                                    false,
		"-modules_cache.json":                     true,
		"-downloads.cache.SQLite3":                true,
		"NBLA'05.SQLite3.search":                  true,
		"-book-chapter-verse-count.cache.SQLite3": true,
		"-downloads.cache.SQLite3-journal":        true,
		"dictionaries_lookup.SQLite3.new":         true,
		"ReferenceData.SQLite3":                   true,
		"persisted_registry.json":                 true,
	}
	for name, want := range cases {
		if got := skipped(name); got != want {
			t.Errorf("skipped(%q) = %v, want %v", name, got, want)
		}
	}
}
