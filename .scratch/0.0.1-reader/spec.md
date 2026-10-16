# MyBible TUI 0.0.1 — Terminal reader for MyBible modules

Status: ready-for-agent

## Problem Statement

I read the Bible on Android with MyBible and I already have a library of MyBible modules (`.SQLite3` files). On my Linux desktop I have no reader that uses those modules: I would have to re-download a Bible, learn a second app, and keep two libraries in sync. I want to read my existing modules in the terminal, where I already live, with a modern keyboard-driven interface.

## Solution

A terminal (TUI) Bible reader for Linux that reads MyBible modules **in place, read-only** — no import, no second copy, no new library to maintain. Modules are read from a default folder, so it works the moment it is dropped in, with config to point it elsewhere; it lists the Translations it finds, remembers my Active translation, and lets me read (Book → Chapter → Verse), jump anywhere through filterable modals, search the active text, and retrace my steps with back/forward history. It ships as a single static binary under the MIT license, starting at version 0.0.1.

## User Stories

1. As a reader, I want a Module directory that works without setup, so that I can start reading the MyBible modules I already have.
2. As a reader, I want the reader to scan the Module directory at startup, so that new modules I drop in are picked up without extra setup.
3. As a reader, I want to see the list of discovered Translations, so that I can pick the one I want to read.
4. As a reader, I want to switch the Active translation from a modal, so that I can compare editions without leaving the reader.
5. As a reader, I want my Active translation remembered between runs, so that I do not re-pick it every time.
6. As a reader, I want the reading pane to be the whole screen, so that the text is the focus and nothing competes with it.
7. As a reader, I want to read one Chapter at a time, so that I always know where I am.
8. As a reader, I want a Verse cursor that moves with `j`/`k`, so that I can track my place and the reader knows which Verse I am on.
9. As a reader, I want `Ctrl-D`/`Ctrl-U` to page up and down within a Chapter, so that I can move fast without counting verses.
10. As a reader, I want `[` and `]` to move to the previous and next Chapter, so that I can read continuously through a Book.
11. As a reader, I want `g`/`G` to jump to the top and bottom of a Chapter, so that I can reach either end immediately.
12. As a reader, I want a modal to pick any Book, so that I can jump across the Bible quickly.
13. As a reader, I want to filter a modal by typing, so that I can find a Book or Translation by name without scrolling.
14. As a reader, I want a modal to pick a Chapter, so that I can jump to a specific chapter of the current Book.
15. As a reader, I want a "go to Reference" modal that lists and filters the Books as I type and accepts input like `John`, `John 3`, or `John 3:16`, so that I can jump to a Book, a Chapter, or an exact Verse with the same keystrokes.
16. As a reader, I want to search the Active translation for a text string, so that I can find verses by their words.
17. As a reader, I want search to ignore letter case and accents, so that `jesus` finds `Jesús`.
18. As a reader, I want search results as a scrollable list of References with a text snippet, so that I can choose where to jump.
19. As a reader, I want pressing Enter on a search result to jump to that Verse and close the modal, so that search and reading flow together.
20. As a reader, I want back/forward navigation history, so that after a jump I can retrace where I came from.
21. As a reader, I want a history modal listing recently visited References, so that I can jump straight to one of them.
22. As a reader, I want my last-read Reference restored when I reopen the reader, so that I resume where I left off.
23. As a reader, I want a top bar that lists the available commands together with their keys, so that I can discover what I can do without memorizing everything.
24. As a reader, I want the current Reference shown in the top bar, so that I always know what I am reading.
25. As a reader, I want to toggle the top bar with `Ctrl-B`, so that I can maximize reading space when I want it.
26. As a reader, I want the words of Jesus rendered in red, so that the text reads the way it is meant to be read.
27. As a reader, I want translator-inserted words rendered in italics, so that I can tell them apart from the base text.
28. As a reader, I want paragraph breaks and line breaks preserved, so that poetry and prose keep their shape.
29. As a reader, I want Strong's numbers, morphology, footnotes, and subheadings hidden for now, so that the reading text stays clean rather than littered with markup.
30. As a reader, I want each Verse on its own line with its number, so that I can point to a Verse at a glance.
31. As a reader, I want deuterocanonical Books to appear when the module contains them, so that my canon is not silently trimmed.
32. As a reader, I want a dark and a light theme, so that the reader fits my terminal.
33. As a reader, I want to override theme colors in config, so that I can match my terminal palette.
34. As a reader, I want the interface in English by default, so that it is usable for anyone.
35. As a reader, I want Spanish as a second interface language, so that I can use it in my own language.
36. As a reader, I want the interface language chosen by config, falling back to my system locale, so that I do not have to set it twice.
37. As a reader, I want a help modal listing every keybinding, so that I can learn the interface from inside it.
38. As a reader, I want `q` to quit, and `Esc` to close any modal, so that leaving is predictable.
39. As a reader, I want my modules opened read-only, so that the reader can never damage or modify my library.
40. As a reader, I want a corrupt or incomplete module skipped with a visible warning, so that one bad file does not stop me reading with the others.
41. As a reader, I want the reader to work fully offline, so that reading never depends on a network.
42. As a reader, I want a single static binary with no runtime dependencies, so that I can drop it anywhere and run it.
43. As a reader, when no modules are found, I want a clear message naming where the reader looked and how to point it elsewhere, so that I am not stuck at an empty screen.
44. As a maintainer, I want the version to start at 0.0.1 with no public releases yet, so that the format and UX can settle before anyone depends on them.
45. As a maintainer, I want continuous integration running build, tests, and lint, so that the code stays green as it grows.
46. As a maintainer, I want the code, comments, and generated documents written in English, so that the project is consistent and open to contributors.

## Implementation Decisions

**Language and stack.** Go with Bubble Tea v2, Lip Gloss v2, and Bubbles v2, using a pure-Go SQLite driver so the binary stays CGO-free and cross-compiles to a single static binary. MIT license. Go module path `github.com/iax7/mybible`. (ADR-0001.)

**Modules are read in place.** The reader opens MyBible `.SQLite3` files read-only and never imports or copies their content. Our own state (config, history) lives in XDG files. No database of our own. (ADR-0002.)

**Module discovery.** At startup the reader scans the Module directory — the folder named in config, or the reader's default, `$XDG_DATA_HOME/mybible` (`~/.local/share/mybible`) — for `*.SQLite3`, determines each Module's type from the filename suffix, and reads its `info` table for the human title and language. In 0.0.1 only Bible Modules (no suffix) are surfaced as Translations; other types are ignored. A module that fails to open, or lacks a required table, is skipped and reported as a warning.

**In-memory data.** When a Translation becomes active, its full text is loaded into memory (a Bible is roughly 31,000 verses). Navigation and search run against that in-memory copy, so there are no per-scroll queries. Moving to another Translation replaces the in-memory copy.

**Canonical book list.** The reader ships its own ordered list of Bible Books (the 66, plus deuterocanonicals), with stable ids that match MyBible's `book_number` scheme. Each module's own book table is mapped into it by `book_number`, honouring an `is_present` flag where present and tolerating book numbers that changed across MyBible format revisions. UI book names come from our list, not from the module, so naming is consistent across Translations. The module's own short and long names are read too and kept on the loaded Books, so an input can name a Book the way the module does (`Gn`, `Jhn`) while everything the reader shows stays canonical.

**Search.** Text search over the Active translation, case- and accent-insensitive. Insensitivity is achieved by normalizing text in Go at load time (lowercase, strip diacritics) rather than by SQL. Results are a Reference plus a snippet, returned in canonical order.

**Markup rendering.** Verse text keeps inline markup in a small supported subset: `<i>` becomes italics, `<J>` becomes the red "words of Jesus" style, `<pb/>` and `<br/>` become paragraph and line breaks. `<S>` (Strong's), `<m>` (morphology), `<f>` (footnote), and `<h>` (subheading) markers are stripped from the display in 0.0.1.

**Interface.** The reading pane fills the screen. A top bar shows the Active translation and the Book and Chapter being read, with their keys, plus the three commands that leave the reading flow — History, Search, Help. Every other key is listed in the Help modal. The bar is toggled with `Ctrl-B` and visible by default, and a Book's name carries the color of its Book classification. Every non-reading action opens a modal (Select Book, Select Chapter, Go to Reference, Select Translation, Search, History, Help). Modal lists are filterable by typing; `Enter` confirms; `Esc` closes. The Go to Reference modal is a list and an input at once: it lists the Books its input names — nothing until the first keystroke, since listing them all is what Select Book is for — `Tab` writes the highlighted Book into the input, spaced so a Chapter typed next lands beside the Book's name, and `Enter` applies what is written: a Book alone means its first Verse, a Chapter without a Verse means Verse 1, and an input naming no Book is read as the translation's first Book. A Chapter and a Verse may be divided by `:`, a comma, a period, or a space, and the modal says what it accepts: the shapes to type while no Book is named, the range of Chapters of the Book it names once one is, and that Chapter's range of Verses once a separator closes it. A Chapter being typed is never taken as final — `3` may still become `33` — so the Verses appear only behind a separator, the space included, and a rejected Reference keeps the range it should have fallen in beside the reason. A Book named by the module's own short or long name (`Gn`, `Jhn`) is resolved when a single Book answers to it; anything else is rejected without leaving the modal. Keymap: `b` Book, `c` Chapter, `v` Go to Reference, `t` Translation, `/` Search, `H` History, `?` Help, `q` Quit; `j`/`k` Verse down/up, `Ctrl-D`/`Ctrl-U` page, `[`/`]` previous/next Chapter, `g`/`G` top/bottom of Chapter, `Ctrl-B` toggle top bar.

**History.** Two related behaviours: a back/forward stack of visited References driven by jumps, and a persisted last-read Reference restored on open. The back/forward stack is kept for the session only; the last-read Reference is persisted in the XDG config directory.

**Configuration.** A TOML config in the XDG config directory holds the Active translation, the interface language, the theme selection, and optional theme color overrides, plus an optional Module directory overriding the default. Interface strings are embedded English and Spanish maps inside the binary, with no i18n framework; language resolves from config, then the system locale, then English.

**Interfaces (seams).** The TUI depends on a small reader interface that exposes the available Translations and, for the Active translation, Books, Chapters, and Verses as values plus lookup by Reference and by search term. The SQLite-backed implementation lives behind that interface, so the TUI can be tested without SQLite and the reader can be tested without the TUI. The interface is defined at the point of use in the TUI and implemented by the module layer.

## Testing Decisions

Good tests here exercise **external behaviour**: what the reader returns for a given module on disk, and what the TUI renders and does for a given sequence of keystrokes. They do not reach into internal structures, call unexported functions, or assert on private state.

Two seams:

- **TUI model (primary, highest).** Tests construct the model against a temporary Module directory and temporary config/data directories, send keystrokes, and assert on the rendered view and observable outcomes (Active translation, current Reference, search results, history). This one seam covers navigation, translation switching, search, history, the modals, the top bar, and i18n.
- **Module reader (low, narrow).** Tests open a generated `.SQLite3` fixture directly and assert the parsed Translation: Books, Chapters, Verses, markup handling, canonical book mapping, and tolerance of quirks (legacy book numbers, missing optional tables, a corrupt file). This exists because those input quirks are impractical to provoke through the TUI.

**Fixtures are generated at test time**, using the same Go SQLite driver, not committed as binary files. This keeps binaries out of the repository and lets each quirk case be built to measure.

Only a few flows use `teatest` golden files (full-screen snapshots); the rest assert on returned values or on view content directly, so the suite does not become brittle. Prior art: there is no existing test code in this repository; the teatest package's own examples are the reference for the golden tests.

## Out of Scope

Notes, bookmarks, and highlights. Cross-references. Commentaries. Dictionaries and lexicons. Strong's and morphology display. Daily devotions. Reading plans. Subheadings display. Audio and text-to-speech. Parallel side-by-side Translations. Right-to-left / bidi scripts. Mouse input. macOS and Windows support. Published release binaries and Homebrew packaging. Any writing to Module files. Non-Bible Module types beyond ignoring them.

## Further Notes

- The project name matches the name of the upstream Android app. This was accepted deliberately; it carries a small confusion and trademark risk worth remembering if the project is ever distributed widely.
- MyBible module content is mixed-license and mostly not redistributable. No module file may ever be committed to this repository; tests rely on synthetic fixtures only.
- Relevant decisions are recorded in `docs/adr/0001-use-go-bubbletea.md` and `docs/adr/0002-read-modules-in-place.md`; vocabulary is fixed in `GLOSSARY.md`.
