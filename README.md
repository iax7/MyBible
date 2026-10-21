# mybible

A terminal (TUI) Bible reader for Linux that reads [MyBible](https://mybible.zone) modules in place.

MyBible stores each translation as a `.SQLite3` file. This reader opens those files
read-only — no import, no second copy, no new library to maintain.

**Status: 0.0.2, pre-release.** The module format handling and the interface are still
settling, so there are no published binaries yet.

## Build and run

Requires Go 1.27.2 or newer: the floor `go.mod` sets, and the 1.27 line `mise.toml` installs.

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
| `m` | Commentary |
| `Enter` | Open commentary |
| `/` | Search |
| `H` | History |
| `L` | Log (only when there are warnings) |
| `?` | Help |
| `q` | Quit |
| `Ctrl-B` | Toggle the top bar |
| `Esc` | Close a modal |

## Modules

Modules live in `~/.local/share/mybible` (or `$XDG_DATA_HOME/mybible` when set).
Drop your `.SQLite3` files there and the reader finds them at startup. To read them
from somewhere else, set `module_directory` in the config file. Module files are
opened read-only and are never modified or copied. Module content is mixed-license
and mostly not redistributable, so no module file belongs in this repository.

Subheadings are read too: a `subheadings` module named after its Bible
(`RV60-s.subheadings.SQLite3` belongs to `RV60.SQLite3`) is paired with it
automatically, and its headings render above the verses they introduce.

Commentary modules are found by their declared type, not by file name: every
module the file calls a `commentaries` module is offered under `m`, whichever
Bible it was written for. The one you pick labels the verses it covers and opens
in its own modal on `Enter`, with `Tab` to switch between modules.

## License

MIT. See `LICENSE`.
