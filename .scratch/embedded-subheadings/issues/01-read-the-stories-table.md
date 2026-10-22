# 01: Read the Translation's own stories

**What to build:** `Load` reads the optional `stories` table of the Bible Module
as the Translation's Subheadings, with the same book mapping, ordering, trimming
and `<x>` hiding the companion loader already applies. A readable companion
Subheadings Module still wins; a missing or broken companion leaves the embedded
headings in place. `Load` returns the warnings it collected, and the TUI puts
them in the Log. A missing `stories` table is silent; a table that exists but
cannot be read is reported. Nothing changes for a Translation with neither
source.

**Blocked by:** nothing.

**Status:** done

- [x] `Load` reads `stories` from the Translation's own Module when the table is present.
- [x] A readable companion Subheadings Module replaces the embedded headings.
- [x] A broken companion falls back to the embedded headings and adds a Log warning.
- [x] A broken `stories` table adds a Log warning and the Bible still loads.
- [x] A Translation with no `stories` table and no companion loads without warnings.
- [x] `stories` titles hide `<x>` content and are ordered by `order_if_several`.
- [x] `modtest.Stories` builds a synthetic Bible with an embedded `stories` table.
- [x] Module tests cover the embedded source, precedence, both failure paths and the silent case.
- [x] A TUI-model test asserts an embedded Subheading renders above its Verse.
- [x] A TUI-model test asserts a broken `stories` table appears in the Log.
