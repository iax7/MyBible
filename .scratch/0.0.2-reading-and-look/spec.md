# MyBible TUI 0.0.2 — Prose reading, Subheadings, the zellij look, and a Log

Status: ready-for-agent

## Problem Statement

0.0.1 proves the reader works, but the reading pane is rough and the interface's
colours clash with the terminal it lives in.

- A Verse longer than the pane is **cut off**: `renderLines` only breaks where the
  Module writes `<br/>`/`<pb/>`, never at the pane's width, so the end of a long
  Verse is unreadable.
- A Chapter has **no heading** inside the pane: nothing says which Chapter the
  text belongs to once the top bar scrolls out of attention.
- The **Subheadings** that MyBible ships for a Translation — the little section
  titles ("La creación", "El diluvio") — are not shown, because the reader ignores
  every Module type that is not a Bible.
- The **top bar and the modals** use colours that fight each other: the Reference
  is painted in its Book classification colour on a dark blue bar, and modals sit
  in a plain box. The maintainer reads in zellij and wants the reader to look like
  it.
- **Warnings** ("`Broken.SQLite3`: not a database") are appended to the bottom of
  the reading pane, where they look like part of the text and cannot be dismissed.

## Solution

Make the reading pane read like prose, show the Subheadings the user already has,
adopt zellij's default theme for the bar and the modals, and move warnings into a
Log modal reached from the bar.

## User Stories

1. As a reader, I want a long Verse to wrap at the pane's width, so that I can read all of it.
2. As a reader, I want the wrapped lines to line up under the Verse's text rather than under its number, so that the shape of a paragraph is easy to see.
3. As a reader, I want a centred Chapter heading at the start of the Chapter, highlighted in the Book's colour, so that I can tell at a glance where I am.
4. As a reader, I want that heading to be part of the Chapter text — it scrolls away like the title of a printed page — so that it does not take space away from the text.
5. As a reader, I want to see the Subheadings of the Chapters I read, so that I can follow the sections of a Book.
6. As a reader, I want a Subheading to sit above the Verse it introduces, in its own style and outside the Verse cursor, so that it reads as a title and not as text.
7. As a reader, I want cross-reference markers inside a Subheading hidden, so that the heading stays clean.
8. As a reader, I want the reader to find the Subheadings Module that belongs to my Active translation by itself, so that I do not have to configure it.
9. As a reader, I want nothing to change when my Active translation has no Subheadings Module, so that a missing companion is not an error.
10. As a reader, I want the top bar and the modals to wear zellij's colours, so that the reader sits quietly inside my terminal.
11. As a reader, I want the Reference in the top bar filled with its Book's classification colour, as zellij fills its selected tab, so that the Book still identifies itself.
12. As a reader, I want a modal framed the way zellij frames a dialog — coloured border with the title in the top edge and the closing key at its right — so that modals are recognisable at a glance.
13. As a reader, I want warnings kept out of the reading pane, so that they never look like part of the text.
14. As a reader, I want a `Log` entry in the top bar, marked with a warning emoji, that appears **only when there are warnings**, so that a clean run stays uncluttered.
15. As a reader, I want `L` to open a Log modal listing every warning as a bullet, so that I can read what went wrong without leaving the reader.
16. As a reader, I want the bar and the modals to keep working exactly as they do now, so that the new look costs me no keystrokes.
17. As a reader, I want the interface in English and Spanish as before, so that the new strings read in both languages.

## Implementation Decisions

**Reading pane.** `renderLines` gains two behaviours:

- *Wrapping.* A Verse's rendered lines are broken at the pane's width, between
  words, keeping each Run's style. The Verse number occupies the first line; every
  continuation line is indented by the number's width plus one space, so text
  lines up under text. A word wider than the line is left whole rather than cut.
- *Chapter heading.* The first line of the Chapter is a centred `Chapter N`
  (translated), rendered with the Book's classification colour as its background
  and the bar's own colour as its foreground. It is a line of the Chapter like any
  other: it scrolls with the text and is not part of any Verse's cursor span.

**Subheadings.** A new Module type, `subheadings`, is discovered but not
surfaced as a Translation. Its file name carries the abbreviation of the Bible it
belongs to with a `-s` suffix (`RV60-s.subheadings.SQLite3` belongs to `RV60`),
and `Discover` pairs the two. The pairing lives in the module layer: the TUI asks
`Reader.Subheadings(book, chapter)` and never sees a file name or a suffix.

When a Translation is loaded, its Subheadings Module is loaded with it, mapped
into the Canonical book list by `book_number` like the Verses are. A Module that
cannot be loaded leaves the Translation readable with no Subheadings; it is a
companion, not a requirement. `Subheading` is `{Verse, Text}`: the Verse it
introduces, and its text with markup rendered away. `ParseMarkup` now hides the
content of `<x>` (cross-references) along with `<S>`, `<m>`, `<f>`, and `<h>`.

**The zellij look.** The `dark` theme is replaced by zellij's default palette —
background `#0B0B0B`, text `#D7D7D7`, green `#96DC5A`, orange `#FF9146` (warnings),
red `#FF5F6E` (the words of Jesus), and `#898989` for Verse numbers. `light` is
unchanged. Book classification colours are untouched and are used everywhere a
Book is named: in the top bar the Reference is filled with its Book's colour
(foreground = the bar's colour) the way zellij fills its selected tab, and in the
lists it stays coloured text. A modal is drawn by one shared `frame` helper: a box
with the title in the top edge and `esc` at that edge's right end, and the key hint
at the foot of the box in the theme's quieter `hint` grey, so a hint does not
compete with the text it explains. `hint` is one more color overridable in config.
A modal **overlays** the reading pane rather than replacing it — it covers only the cells
it needs, so the Chapter stays visible around it — and is centred on the screen.

**Warnings.** The reading pane never renders a warning. `Model.warnings` is shown
by a new list modal (`pickerLog`), reached from a `Log` entry that the top bar
shows only when `warnings` is non-empty, with the key `L` and a warning emoji in
the theme's warning colour. Each warning is one bullet. The Log modal has no
filter, so it does not show a filter line and typing does not narrow it. `Log` is
also listed in the Help modal.

## Testing Decisions

Same two seams as 0.0.1, and the same rule: assert on behaviour, never on private
state.

- **TUI model.** A long Verse wraps and its continuation lines are indented; a
  Chapter opens with a centred heading; a Subheading renders above its Verse;
  the top bar shows `Log` only with warnings; `L` opens the Log and lists the
  warnings as bullets; the reading pane carries no warning text.
- **Module reader.** A synthetic Subheadings Module loads and pairs with its
  Bible by abbreviation; a Translation with no companion loads without
  Subheadings; `<x>` content is hidden by `ParseMarkup`.

Existing tests that assert the old dark palette, the old bar Reference style, and
warnings in the pane are updated to the new behaviour.

## Out of Scope

Cross-reference *display*, Strong's and morphology display, footnotes, notes,
bookmarks and highlights, commentaries, dictionaries, parallel Translations,
reading plans, and any writing to Module files. The `-s` suffix is the only
companion-Module convention honoured in 0.0.2; other companion types stay ignored.

## Further Notes

- The zellij palette is zellij's built-in `default` theme (see zellij's
  `themes/default.kdl`), not an invented one. The decision is recorded in
  `docs/adr/0003-use-zellij-default-theme.md`.
- A broken Subheadings Module is ignored rather than reported: the reader must
  never fail to read the Bible because of a companion file.
