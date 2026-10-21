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
		ActiveCommentary:  "MAC",
		LastReference:     "John 3:16",
		Theme:             "light",
		Language:          "es",
		Colors:            Colors{Jesus: "9", Link: "28"},
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

func TestDefaultModuleDirFollowsTheXDGDataHome(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_DATA_HOME", base)

	got, err := DefaultModuleDir()
	if err != nil {
		t.Fatalf("DefaultModuleDir: %v", err)
	}
	if want := filepath.Join(base, "mybible"); got != want {
		t.Errorf("DefaultModuleDir = %q, want %q", got, want)
	}
}

func TestDefaultModuleDirIgnoresARelativeDataHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_DATA_HOME", "relative/to/nowhere") // invalid: the XDG spec ignores relative paths
	t.Setenv("HOME", home)

	got, err := DefaultModuleDir()
	if err != nil {
		t.Fatalf("DefaultModuleDir: %v", err)
	}
	if want := filepath.Join(home, ".local", "share", "mybible"); got != want {
		t.Errorf("DefaultModuleDir = %q, want %q", got, want)
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
