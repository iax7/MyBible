# 03: Read a Chapter

**What to build:** The user can read. The reader ships a Canonical book list (the 66 Books, plus deuterocanonicals) with stable ids matching MyBible's `book_number` scheme, and maps each Module's own book table into it by `book_number`, honouring an `is_present` flag where present and tolerating book numbers that changed across format revisions. Making a Translation active loads its full text into memory. The reading pane renders one Chapter at a time: one Verse per line with its number, preserving paragraph and line breaks. Inline markup is rendered in the supported subset — `<i>` as italics, `<J>` in the red "words of Jesus" style — while `<S>`, `<m>`, `<f>`, and `<h>` markers are stripped from the display.

**Blocked by:** 02 (Module directory, discovery, and Translation selection).

**Status:** done

- [x] The Canonical book list exists with ids matching MyBible's `book_number` scheme, including deuterocanonicals.
- [x] A Module's books are mapped into the Canonical book list by `book_number`; absent books respect `is_present`.
- [x] Legacy book numbers that changed across format revisions do not break navigation.
- [x] Reading renders one Chapter at a time, one Verse per line, with verse numbers.
- [x] `<i>` renders as italics and `<J>` renders in the red style.
- [x] `<pb/>` and `<br/>` produce paragraph and line breaks; `<S>`, `<m>`, `<f>`, and `<h>` are not shown.
- [x] An empty Verse row (a Verse with no text) renders without error.
- [x] A module-reader test asserts parsing, book mapping, and markup handling for a generated fixture, including a legacy-numbered one.
- [x] A TUI-model test asserts the rendered Chapter content for a generated fixture.
