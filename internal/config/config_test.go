package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReturnsAnEmptyConfigWhenTheFileIsMissing(t *testing.T) {
	got, err := Load(t.TempDir())
	if err != nil {
		t.Fatalf("Load on a missing file: %v", err)
	}
	if got != (Config{}) {
		t.Errorf("Load = %+v, want a zero Config", got)
	}
}

func TestSaveThenLoadRoundTrips(t *testing.T) {
	dir := t.TempDir()
	want := Config{
		ModuleDir:         "/home/me/modules",
		ActiveTranslation: "RST",
		LastReference:     "John 3:16",
		Theme:             "light",
		Language:          "es",
		Colors:            Colors{Jesus: "9"},
	}

	if err := Save(dir, want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got != want {
		t.Errorf("Load = %+v, want %+v", got, want)
	}
}

func TestLoadReportsAMalformedConfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, fileName), []byte("this is not toml ="), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(dir); err == nil {
		t.Error("Load accepted a malformed config file")
	}
}

func TestDirFollowsTheXDGConfigHome(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)

	got, err := Dir()
	if err != nil {
		t.Fatalf("Dir: %v", err)
	}
	if want := filepath.Join(base, "mybible"); got != want {
		t.Errorf("Dir = %q, want %q", got, want)
	}
}
