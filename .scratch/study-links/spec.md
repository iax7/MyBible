# Study links in the reading pane

Status: done

## Problem Statement

MyBible Verses carry two kinds of study link, and the reader shows neither. A
Translation writes Footnote markers in the Verse text as `<f>[1†]</f>`, but
`ParseMarkup` hides the whole tag. Cross-references live in `.crossreferences`
Modules — `cross_references(book, chapter, verse, verse_end, book_to, chapter_to,
verse_to_start, verse_to_end, votes)` — which the reader does not even discover.

The maintainer wants the behaviour MySword and MyBible show: the markers are
visible in the text, and activating one brings up the related verses or the
explanatory note without losing the reading place. MyBible 5.8.7 was disassembled
to settle the details, so the reader follows the format rather than inventing a
shape: `BibleChaptersRetriever` builds `[⇅N]` after the Verse, its popup lists
each reference with the target passage, and the link order is votes descending
with Canonical order on a tie.

## Solution

Read `<f>…</f>` as a Footnote marker Run in the link style, and discover the
Cross-reference sets. The reader turns sets on under `X`; the reading pane ends
every Verse they speak to with `[⇅N]`; `x` opens the References popup with each
reference and the passage it points to. The notes a Footnotes module carries show
behind their marker in the Commentary modal, and `n` opens that modal at the
Footnotes module of the translation being read.

## User Stories

1. As a reader, I want to see where a Translation carries Footnotes, so that I
   know which words have an explanation.
2. As a reader, I want the note behind the marker, so that I can read it without
   leaving the Verse.
3. As a reader, I want to see which Verses have Cross-references and how many, so
   that I can decide whether to open them.
4. As a reader, I want the References popup to show the passages, not only their
   addresses, so that the study link does not cost me my place.
5. As a reader, I want to choose which Cross-reference sets are read, several at
   once if I like, so that I can combine editions.
6. As a reader, I want the same target from two sets counted once and the most
   voted link shown, so that the count and the order mean something.
7. As a reader, I want the reader to keep working when a set or a note is
   missing, so that a bad Module never costs me the Chapter.

## Implementation Decisions

**Module layer.** `DiscoverCrossReferences` finds `.crossreferences` Modules;
`CrossReferenceSet.Rows(book, chapter)` reads one Chapter, including the reversed
links a set with `info.requires_reverse_processing` asks for, and maps every book
number into the Canonical list. `CrossReferencesFor(rows, verse)` keeps the links
whose source range covers the Verse, a target once with its highest-voted link,
and sorts by votes then Canonical order. `CrossReference.Label` renders
`Book 8:22–24`, with a `(from–to)` prefix when the link speaks from a range.

**Footnote markers.** `ParseVerse` is `ParseMarkup` plus a Marker Run for each
`<f>`; `ParseMarkup` still hides them, so Search and snippets stay free of
markers. The info flag `add_space_before_footnote_marker` inserts the space a
Translation asks for. The Commentary modal prefixes each note of a Footnotes
Module with its marker; older Modules without the `marker` column still load.

**References popup.** It is read-only in this version: it scrolls, and its links
are drawn in the link style but are not selectable. This mirrors the reader's
Commentary links, which have the same limit.

**Selection.** `active_crossreferences` is a list in the config. A config that
never named one starts with the first set installed; an explicit empty list is
the reader's choice to show none and survives a restart.

## Out of Scope

MyBible's alternative mode that draws each reference inline instead of `[⇅N]`,
the full separate "References" window, user-defined Cross-references, votes and
Module names inside the popup, filtering the popup by typing, jumping from the
popup, and Russian Psalm numbering (`russian_numbering`).
