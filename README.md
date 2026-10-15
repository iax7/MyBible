# mybible

A terminal (TUI) Bible reader for Linux that reads [MyBible](https://mybible.zone) modules in place.

MyBible stores each translation as a `.SQLite3` file. This reader opens those files
read-only — no import, no second copy, no new library to maintain.

**Status: 0.0.1, pre-release.** The module format handling and the interface are still
settling, so there are no published binaries yet.

## Build and run

Requires Go 1.24 or newer (the exact version is pinned in `mise.toml`).

```sh
go build ./cmd/mybible
./mybible
```

`mybible --version` prints the version.

## Keys

| Key | Action |
| --- | --- |
| `b` | Book |
| `c` | Chapter |
| `v` | Go To |
| `t` | Translation |
| `/` | Search |
| `H` | History |
| `?` | Help |
| `q` | Quit |
| `Ctrl-B` | Toggle the top bar |
| `Esc` | Close a modal |

## Modules

Point the reader at the folder holding your MyBible modules. Module files are opened
read-only and are never modified or copied. Module content is mixed-license and mostly
not redistributable, so no module file belongs in this repository.

## License

MIT. See `LICENSE`.
