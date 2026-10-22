# 01: Show `<x>` references in Subheadings as Verse links

**What to build:** A Subheading's text may carry `<x>book chapter:verse…</x>`
cross-references. Parse each target, render a readable label from the Canonical
book name and the location, and mark that label as a `Run.Link` so the TUI paints
it with the same green underline as a Commentary Verse link. The text around the
tag stays as it is. A Subheading that is only references is kept, because its
labels now carry text. The Subheading's own words are the quieter cyan.

This is visual only: no click, no key, no `jumpTo`. The target is parsed to build
the label, not to navigate.

**Blocked by:** nothing.

**Status:** done

**Context:** Commentary links already produce `Run.Link` runs, but
`ParseCommentary` only recognises `<a href='B:…'>` to mark the **anchor text** as
a link; it does not parse the `href` target. `<x>` is different: the content
**is** the target and there is no anchor text, so this ticket needs target
parsing plus a generated label. The `Run.Link` mechanism, the tag-scanning
pattern, and the Commentary's green `styles.link` are reusable.

**Target grammar:** `<x>` content follows MyBible's dictionary hyperlink target
format, with the leading `B:` optional (RVR60 writes it without). Parse:

- `book chapter:verse`
- `book chapter:verse-endVerse`
- `book chapter:verse-endChapter:endVerse`
- `book chapter:verse1,verse2,…`
- `book chapter`
- `book chapterStart-chapterEnd`
- `book verse-verse` — RVR60's `720 3-13`; without a `:` this is a chapter range
  per the spec, and its label is `Jude 3–13`

**Label rules:** Use the reader's Canonical `Book.Name` (the same name the top bar
shows) and an en dash for ranges, matching the Commentary range formatting:

- `10 3:1-24` → `Genesis 3:1–24`
- `500 1:3-2:10` → `John 1:3–2:10`
- `10 1:1,3,5` → `Genesis 1:1,3,5`
- `10 1` → `Genesis 1`
- `10 1-3` → `Genesis 1–3`
- `720 3-13` → `Jude 3–13`

Multiple `<x>` tags in one title become multiple link runs; the text between them
(`; `, `—`, parentheses) stays plain. Surrounding literal text is preserved:
`(<x>130 1:1-4</x>)` renders as `(` + link `1 Chronicles 1:1–4` + `)`.

**Malformed targets:** If the book number is unknown, a number is below 1, or the
target does not parse, hide the `<x>` and its content, as today. If the whole
heading becomes empty, drop it. Do not change `ParseMarkup`: `<x>` in Verse text
is not documented and stays hidden.

**Where it lives:** Parse in the module layer, in `loadHeadings`, so both
`stories.title` and a companion `subheadings.subheading` get it. Recommendation:
add `Runs []Run` to `Subheading` and keep `Text` as the flattened display text;
build both in `loadHeadings`, run `hasLetterOrDigit` on `Text` (so a
reference-only heading is no longer dropped), and have the TUI render `Runs` with
`m.styles.subheading` for plain runs and `m.styles.link` for link runs. Headings
still render as one line; long ones can overflow as they do today.

**Color:** The Subheading's own words are cyan, quieter than the green Verse
links inside it:

- dark theme: `#5FD7D7`
- light theme: ANSI `30` (dark cyan, which reads on a light background)

Add `Subheading string \`toml:"subheading"\`` to `config.Colors`, wire it through
`paletteFor`, and make `styles.subheading` bold with that color. The link runs
keep `styles.link` (green + underline), so `subheading` and `link` are
independently overridable; an empty override keeps the theme default.

**Testing:**

- Module tests for every target format, multiple references in one title,
  surrounding punctuation, a reference-only heading kept, and a malformed target
  hidden.
- Update the existing tests that use `The <x>Gen 1:1</x>fall` and expect
  `The fall`: `Gen 1:1` is not a numeric target, so either turn that into the
  malformed-target case or replace it with `10 1:1` and assert the label.
- A TUI test asserts plain Subheading text takes the cyan `styles.subheading`
  and the label takes the green `styles.link`; no navigation is wired.
- Theme tests assert the dark default is cyan, the light default is its dark
  cyan, and a `subheading` override wins without touching `link`.
- Config round-trip tests cover the new `subheading` key.
- Real-data check: RVR60 has 2,643 `stories` rows and currently loads 2,077
  headings; 567 titles carry `<x>` and 566 are reference-only and dropped. After
  this change all 2,643 should load. NBLA'05, LBLA and LBLA+ have no `<x>`, so
  their counts do not change.

**Out of scope:** clicking or pressing a key to jump to the target (the only part
not supported yet), `<x>` in Verse text, `.crossreferences` modules, rendering
the target label of a Commentary `<a href>`, wrapping long Subheadings, and
`<h>` inline Subheadings.

- [x] `<x>` targets are parsed in `stories.title` and `subheadings.subheading`.
- [x] Labels use the Canonical Book name and the documented range/list formats.
- [x] Multiple references in one heading and the surrounding text render correctly.
- [x] Reference-only headings are kept instead of being dropped by `hasLetterOrDigit`.
- [x] Malformed targets stay hidden and do not crash or leak raw book numbers.
- [x] Subheading words use a cyan `subheading` style; links use the Commentary's green `link` style, with no jump behaviour.
- [x] `subheading` is a separate, configurable color; `link` still styles Commentary links.
- [x] Existing subheading and theme tests are updated for the new behaviour.
- [x] The glossary's **Verse link** entry covers Subheadings.
- [x] The embedded-subheadings spec's out-of-scope note points at this ticket.
