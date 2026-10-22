# Embedded Subheadings

Status: ready-for-agent

## Problem Statement

MyBible Bibles often carry their own section headings in the optional `stories`
table of the Bible Module — the subheadings of the printed edition. The reader
ignores that table. As a result, a Translation with no companion
`subheadings` Module shows no Subheadings at all, even though Android shows them
from the Translation itself. In the maintainer's library this affects
`NBLA'05.SQLite3` (2,575 rows), `LBLA.SQLite3` and `LBLA+.SQLite3` (2,413 rows
each) and `RVR60.SQLite3` (2,643 rows, of which 2,077 survive the cross-reference
filter); `RV'60+.SQLite3` has none.

## Solution

Read the Translation's own `stories` table as its Subheadings. A readable
companion Subheadings Module still wins; the Translation's own headings are the
fallback when there is no companion, or when the companion cannot be read.
Failures are reported in the Log and never stop the Translation from loading.

## User Stories

1. As a reader, I want the Subheadings a Translation already carries, so that
   `NBLA'05` reads with its sections without installing another Module.
2. As a reader, I want the headings above the Verses they introduce, in the same
   style as the companion's Subheadings, so that the source makes no difference
   to how I read.
3. As a reader, I want a companion Subheadings Module to win when it can be read,
   so that installing one deliberately overrides the Translation's own headings.
4. As a reader, I want a broken companion to fall back to the Translation's own
   headings and say so in the Log, so that a bad companion does not silently
   change what I see.
5. As a reader, I want a broken `stories` table reported in the Log while the
   Bible stays readable, so that I know why the headings are missing.
6. As a reader, I want a Translation with no `stories` table and no companion to
   keep reading exactly as before, with no warning, so that the feature is
   invisible when unused.
7. As a reader, I want cross-reference markup inside a heading hidden, as it
   already is in a companion's Subheadings, so that the heading stays clean.

## Implementation Decisions

**Source of Subheadings.** A Translation's headings come from one source at a
time. `Load` reads the Translation's own `stories` table first, if present, and
then loads a paired companion when `Translation.subheadings` names one. A
readable companion replaces the embedded headings entirely, even when it has no
row for a given Chapter. A companion that cannot be read leaves the embedded
headings in place and adds a warning.

**Reading `stories`.** `stories (book_number, chapter, verse, order_if_several,
title)` has the same shape as a Subheadings Module's `subheadings` table. Rows are
read in `book_number, chapter, verse, order_if_several` order, mapped into the
Canonical book list, trimmed, rendered through `displayText` (so `<x>` content is
hidden), and skipped when the result holds no letter or digit. The loader is
shared with the companion's table so both sources get exactly the same
treatment.

**Warnings.** `Translation.Load` returns `(*Text, []string, error)`: the loaded
text, the warnings collected while reading companion data, and a fatal error. A
missing `stories` table is not a warning; a table that exists but cannot be read
is one. A broken companion is also a warning now, replacing the old silent
shortcut. Warning text names the file and stays in English, like the warnings
`Discover` already produces:

- `NBLA'05.SQLite3: cannot read the stories table: …`
- `RV60-s.subheadings.SQLite3: cannot read the subheadings module: …`

The TUI appends these warnings to `Model.warnings` in `loadActive`, so they land
in the existing Log modal and make its top-bar entry appear.

**No new configuration.** Embedded headings are shown whenever they exist. The
reader does not grow a toggle for them.

## Testing Decisions

- **Module reader.** A Bible with a `stories` table loads those rows; a readable
  companion replaces them; a broken companion falls back to them and warns; a
  broken `stories` table warns without stopping the Bible from loading; a Bible
  with neither source loads without warnings.
- **TUI model.** A Bible's own Subheading renders above its Verse, and a broken
  `stories` table appears in the Log.

Fixtures are generated at test time with `modtest`. A new `modtest.Stories`
helper adds a `stories` table to an existing synthetic Bible Module.

## Out of Scope

- `<h>…</h>` inline subheadings embedded inside a Verse, which split a Verse
  mid-text. They are documented separately and tracked as their own issue.
- Rendering the target of an `<x>` cross-reference as a label. `<x>` content
  stays hidden in this feature; tracked in
  `.scratch/subheading-verse-links/issues/01-render-x-references.md`.
- Styling a heading from a Subheadings Module's `font_*` info parameters.
- Merging embedded and companion headings, or filling per-Chapter gaps.
- Introductions, and the other companion Module types.

## Further Notes

- MyBible's module specification says the `STORIES` table "allows to provide
  subheadings from a corresponding printed Bible", that it is optional, and that
  it was largely replaced by separate subheadings modules. Android therefore
  shows these headings without a companion; this feature matches that.
- The chosen precedence follows the same history: a companion module is the
  deliberate, shared source; the Translation's own table is its built-in
  fallback.
- Cross-reference-only headings such as RVR60's `(<x>130 1:1-4</x>)` become empty
  after `<x>` is hidden and are skipped, exactly as they are today in companion
  Subheadings. Showing their labels is tracked in
  `.scratch/subheading-verse-links/issues/01-render-x-references.md`.
