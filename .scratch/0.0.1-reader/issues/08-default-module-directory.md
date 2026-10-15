# 08: A default Module directory

**What to build:** The reader stops expecting the user to configure a Module directory before it is useful. When config names no `module_directory`, the Module directory defaults to the reader's data directory, `<XDG_DATA_HOME>/mybible` (`~/.local/share/mybible`), and that folder is scanned like any other. Config still overrides it, so nothing is taken away — the setting becomes optional instead of mandatory. Modules are user data rather than configuration, which is why the default is the data directory and not the config directory beside `config.toml`. A folder that does not exist is not an error: discovery already reports it as an empty Module directory, so a fresh install reads an explanation naming the path instead of an empty screen with no way forward. The reader does not create the directory.

**Blocked by:** none.

**Status:** done

- [x] `config.DefaultModuleDir` returns `<XDG_DATA_HOME>/mybible`, falling back to `$HOME/.local/share/mybible`, ignoring a relative `XDG_DATA_HOME` the way the XDG spec requires.
- [x] `tui.New` falls back `opts.ModuleDir` → `cfg.ModuleDir` → `config.DefaultModuleDir()`, so a reader with no configured folder still finds Modules in the default one.
- [x] An unresolvable default (no `XDG_DATA_HOME`, no home directory) leaves the reader with no Module directory and the existing empty-state message.
- [x] Config still overrides the default, and the effective directory is still persisted to `config.toml`.
- [x] Tests never read the developer's real config and data directories.
- [x] `GLOSSARY.md` and `README.md` describe the default; the spec's Module discovery and Configuration decisions match.

## Notes

No migration: 0.0.1 is pre-release with no published binaries, so no existing user has a directory to move.
