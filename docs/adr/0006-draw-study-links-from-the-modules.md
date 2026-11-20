# Draw study links from the Modules, and let the References popup read in place

The reading pane shows two study links that were invisible until now: a Footnote marker, the text a Translation writes inside `<f>…</f>` in the Verse itself, and a Cross-reference marker `[⇅N]`, which the reader appends to a Verse the Active Cross-reference sets speak to. The marker's N is the number of distinct targets after merging the sets, a target several sets share counted once with its most voted link winning. `x` opens the References popup with each reference and the passage it points to, read without leaving the Chapter; the popup does not navigate. This is what MyBible 5.8.7 does with the same data — its `BibleChaptersRetriever` builds the marker as `[⇅%d%s]`, its balloon shows each reference followed by the target text, and its link ordering is votes descending, Canonical order on a tie — so the reader invents nothing on top of the format.

## Considered Options

- **A picker of references that jumps on Enter (MyBible's separate References window)** — the smallest popup, but it loses the "read the related verses without leaving your place" that the feature is about.
- **One active Cross-reference set rather than several** — a simpler selector and count, but MyBible reads a list of sets and the Modules are written to be combined.
- **No marker, only a key** — quieter on a Chapter marked on nearly every Verse, but the reader would have no way to see that links exist.

## Consequences

- The reading pane is busier: a set such as the Treasury of Scripture Knowledge speaks to every Verse, so most Verses end with `[⇅N]`. Turning the set off in `X` is the escape, and the config keeps an explicit empty list as the reader's choice.
- `[⇅N]` and the Footnote markers take the existing link style; no new theme color was added.
- The popup's links are text, not selectable, until the reader grows a way to focus a line inside a modal. The Commentary's Verse links have the same limit.
- The reader keeps the format's data rules: `requires_reverse_processing` links are read both ways, source and target Verse ranges are honoured, and Russian Psalm numbering is still ignored.
