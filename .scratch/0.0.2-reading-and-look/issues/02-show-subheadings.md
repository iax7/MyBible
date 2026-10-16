# 02: Show Subheadings

**What to build:** The reader discovers `subheadings` Modules without surfacing
them as Translations, pairs each with the Bible whose abbreviation it carries
(`RV60-s.subheadings.SQLite3` with `RV60`), loads the companion when its
Translation is loaded, and renders its Subheadings above the Verse each one
introduces, in their own style and outside the Verse cursor. A Translation with no
companion loads unchanged and shows nothing. `<x>` (cross-reference) content is
hidden. The pairing lives in the module layer: the TUI asks
`Reader.Subheadings(book, chapter)` and never sees a file name.

**Blocked by:** 01 (both render the Chapter's lines).

**Status:** done

- [x] `subheadings` Modules are discovered and paired by abbreviation; they never appear as Translations.
- [x] Loading a Translation loads its companion's Subheadings, mapped into the Canonical book list.
- [x] A Subheading renders on its own line above the Verse it introduces, in its own style.
- [x] A Subheading is not part of any Verse's cursor span.
- [x] `<x>` content is hidden in Subheadings and in Verse text.
- [x] A Translation with no companion, or a broken one, reads normally with no Subheadings and no crash.
- [x] Module-reader tests assert pairing, loading, and `<x>` hiding for generated fixtures.
- [x] A TUI-model test asserts a Subheading renders above its Verse.
