# 03: Mark the Verses and open the References popup

**What to build:** Read the active Cross-reference sets' rows once per Chapter,
grouped by the Verse each link speaks to. The reading pane ends every Verse with
links with `[⇅N]`, in the link style, after the Verse text or on an indented line
of its own when it does not fit. `x` opens the References popup: the title names
the Reference, and the body is each link's label followed by the passage it
points to, read from the Active translation; a target the translation lacks keeps
its Canonical-named label. `Esc` closes it, `↑/↓` or `j/k` and the page keys
scroll. `X` opens the selector: one row per installed set, `Space` or `Enter`
turns it on or off, and the change persists and rereads the Chapter at once. The
popup explains the two empty states — no set on, no link for the Verse.

**Blocked by:** 01, 02.

**Status:** done

- [x] The Chapter cache rereads on a Chapter change, a set change and a translation change.
- [x] A Verse the active sets speak to ends with `[⇅N]`, N counting distinct targets.
- [x] A Verse without links carries no marker.
- [x] `x` opens the popup with each label and its target passage; a missing Book keeps the label.
- [x] `X` turns sets on and off, several at once, and the markers follow immediately.
- [x] The empty states are explained instead of opening an empty box.
- [x] `active_crossreferences` persists, and an explicit empty list survives a restart.
- [x] A config that never chose starts with the first installed set.
- [x] New keys, titles and hints read in both languages.
- [x] TUI tests cover the markers, the popup, the selector, the empty states and persistence.
