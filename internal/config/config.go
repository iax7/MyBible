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
