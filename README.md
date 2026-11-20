# mybible

A terminal (TUI) Bible reader for Linux that reads [MyBible](https://mybible.zone) modules in place.

MyBible stores each translation as a `.SQLite3` file. This reader opens those files
read-only — no import, no second copy, no new library to maintain.

**Status: 0.3.0, pre-release.** The module format handling and the interface are still
settling, so the installers below start working with the first tagged release.

## Install

```sh
mise use -g github:iax7/MyBible        # mise, downloads the release binary
```

Releases are cut by pushing a `v<version>` tag: CI builds linux on amd64 and
arm64 and attaches the archives.

## Build and run

Requires Go 1.27.2 or newer: the floor `go.mod` sets, and the 1.27 line `mise.toml` installs.

The simplest way to get that Go is [mise](https://mise.jdx.dev), which reads
`mise.toml` and installs the pinned toolchain:

```sh
mise install
mise exec -- go build ./cmd/mybible
```

Or, inside the directory, `mise activate` / `mise shell` once and use plain `go`.

With a system Go already at 1.27.2+:

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
| `X` | Cross-reference modules |
| `Enter` | Open commentary |
| `n` | Footnotes |
| `x` | References |
| `/` | Search |
| `H` | History |
| `L` | Log (only when there are warnings) |
| `?` | Help |
| `q` | Quit |
| `Ctrl-B` | Toggle the top bar |
| `Esc` | Close a modal |
| `j` / `↓` | Verse down |
| `k` / `↑` | Verse up |
| `Ctrl-D` / `Ctrl-U` | Page down / up |
| `[` / `]` | Previous / next Chapter |
| `g` / `G` | Top / bottom of Chapter |

## Modules

Modules live in `~/.local/share/mybible` (or `$XDG_DATA_HOME/mybible` when set).
Drop your `.SQLite3` files there and the reader finds them at startup. To read them
from somewhere else, set `module_directory` in the config file. Module files are
opened read-only and are never modified or copied. Module content is mixed-license
and mostly not redistributable, so no module file belongs in this repository.

### Getting modules

The reader downloads nothing: modules are downloaded by the MyBible app on Android.
With the phone connected in file transfer (MTP) mode, this copies everything from
the app's data folder into the Module directory:

```sh
mybible copy-modules
```

Modules are copied as they are, and the app's `user` folder (bookmarks, notes,
themes, abbreviations) comes along as a backup. The app's caches, search indexes
and journals are left out. Files already in the destination are overwritten and
nothing is ever deleted from it. The command respects `module_directory` from the
config.

It reaches the phone over MTP with KDE's `kioclient`, the same stack Dolphin uses.
The app keeps its data in
`/storage/emulated/0/Android/data/ua.mybible/files/MyBible/` or in
`/storage/emulated/0/MyBible/`; the command looks in both, and also under any SD
card the phone mounts.

To copy by hand, the files are in that folder: copy the `.SQLite3` modules, plus
any companion that came with one (`RV60-s.subheadings.SQLite3` next to
`RV60.SQLite3`). Android 11+ hides `Android/data` from USB/MTP on some phones; if
yours does, move the app's data directory or pull the folder with `adb`.

Copying an abbreviation set is only a backup: the reader reads its Preferred set
from `~/.config/mybible/preferred.abbreviations.json`, described below.

By default, Book names come from the Module, not from the reader: a Spanish
Translation lists `Gn · Génesis` and completes `Génesis`. The built-in English
name still works as input — `Genesis 3:16` reaches the same Verse — and a Book's
number is what identifies it, so changing translations keeps your place.

To keep one set of abbreviations across translations, set
`preferred_abbreviations = true` in the config. When no set exists yet, the next
start asks which translation to copy the abbreviations from, then stores them in
`preferred.abbreviations.json` beside `config.toml` (in `~/.config/mybible` by
default); edit that file to taste. The file uses MyBible's desired-abbreviation
format, so it can be shared with the Android app: to start from one of its sets,
copy it to that path. A Book with no entry keeps the module's abbreviation, and
turning the option off leaves the reader exactly as it was.

Subheadings are read too. A Translation carries its own in the module's
`stories` table — the section titles of the printed edition, as `NBLA'05` does
— or a `subheadings` module named after its Bible supplies them
(`RV60-s.subheadings.SQLite3` belongs to `RV60.SQLite3`). The companion wins when
it can be read; the Translation's own headings are the fallback. Either way they
render above the verses they introduce, and a table that cannot be read is
reported in the Log.

Commentary modules are found by their declared type, not by file name: every
module the file calls a `commentaries` module is offered under `m`, whichever
Bible it was written for. The one you pick labels the verses it covers and opens
in its own modal on `Enter`, with `Tab` to switch between modules.

Footnotes are shown where the module writes them: the marker the text carries
(`[1]`, `[1†]`, `ⓐ`) is drawn in the link style, and `n` opens the note — the
Commentary modal's Footnotes tab, chosen for the translation you are reading.
`Enter` opens the same modal on the commentary, and `Tab` switches between the
modules, where a module that declares itself Footnotes puts each note behind its
marker.

Cross-references come from `.crossreferences.SQLite3` modules. `X` lists them and
`Space` turns each on or off; every Verse the chosen sets speak to ends with
`[⇅N]`, N being the number of distinct targets after the sets are merged, the
most voted link winning when two sets share a target. `x` opens the References
popup: each reference with the passage it points to, read without leaving the
Chapter. The sets you choose are kept in `active_crossreferences` in the config.

## License

MIT. See `LICENSE`.
