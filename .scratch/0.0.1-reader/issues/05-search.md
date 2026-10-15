# 05: Search

**What to build:** The user can find Verses by their words. `/` opens a "Search" modal that searches the Active translation, ignoring letter case and accents, so `jesus` finds `Jesús`. Insensitivity comes from normalizing text in Go at load time rather than from SQL. Results appear as a scrollable list of References each with a text snippet, in canonical order; pressing `Enter` on a result jumps to that Verse and closes the modal.

**Blocked by:** 03 (Read a Chapter).

**Status:** done

- [x] `/` opens the Search modal; `Esc` closes it.
- [x] Search runs against the Active translation only.
- [x] Search ignores letter case and accents.
- [x] Results are a scrollable list of Reference + snippet, in canonical order.
- [x] `Enter` on a result jumps to that Verse and closes the modal.
- [x] A search with no matches shows an explicit empty result rather than a blank list.
- [x] A TUI-model test asserts that a query returns the expected Reference and that `Enter` jumps to it, using a generated fixture with accented text.
