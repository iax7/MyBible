// Package config reads and writes the reader's TOML config file.
//
// The file lives in the XDG config directory; nothing else about the reader is
// persisted here. Modules themselves are never touched.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	fileName = "config.toml"
	appDir   = "mybible"
)

// Config is the reader's persisted settings.
type Config struct {
	ModuleDir         string `toml:"module_directory"`
	ActiveTranslation string `toml:"active_translation"`
	LastReference     string `toml:"last_reference"`
	Theme             string `toml:"theme"`
	Language          string `toml:"language"`
	Colors            Colors `toml:"colors"`
}

// Colors overrides individual theme colors. An empty field keeps the theme's
// own color; a value is an ANSI256 number or a hex color.
type Colors struct {
	Bar       string `toml:"bar"`
	BarText   string `toml:"bar_text"`
	Key       string `toml:"key"`
	Label     string `toml:"label"`
	Reference string `toml:"reference"`
	Warning   string `toml:"warning"`
	Jesus     string `toml:"jesus"`
	VerseNum  string `toml:"verse_num"`
	CursorNum string `toml:"cursor_num"`
	Hint      string `toml:"hint"`
	Field     string `toml:"field"`
}

// Dir returns the directory holding the reader's configuration.
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appDir), nil
}

// Path returns the config file's path inside dir.
func Path(dir string) string {
	return filepath.Join(dir, fileName)
}

// DefaultModuleDir returns the folder the reader scans for Modules when config
// names none: the reader's data directory. Modules are user data rather than
// configuration, so they default under XDG_DATA_HOME, not beside config.toml.
//
// os has no UserDataDir, so the XDG rule is spelled out here: a relative
// XDG_DATA_HOME is invalid and ignored, the way os.UserConfigDir treats it.
//
// shortcut: Modules sit in the root of the data directory, sharing it with any
// state the reader stores there later; move them into a subdirectory when the
// reader grows data files of its own.
func DefaultModuleDir() (string, error) {
	base := os.Getenv("XDG_DATA_HOME")
	if !filepath.IsAbs(base) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(base, appDir), nil
}

// Load reads the config from dir. A missing file yields a zero Config and no
// error, so the reader starts on defaults the first time it runs.
func Load(dir string) (Config, error) {
	var c Config
	_, err := toml.DecodeFile(Path(dir), &c)
	if errors.Is(err, fs.ErrNotExist) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read %s: %w", Path(dir), err)
	}
	return c, nil
}

// Save writes the config to dir, creating the directory when needed.
func Save(dir string, c Config) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.Create(Path(dir))
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	encoder := toml.NewEncoder(f)
	encoder.Indent = ""
	return encoder.Encode(c)
}
