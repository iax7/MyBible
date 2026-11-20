# 02: Show Footnote markers

**What to build:** `ParseVerse` is `ParseMarkup` for the reading pane: a
`<f>…</f>` tag keeps its content as a Marker Run instead of hiding it. The marker
is drawn in the link style, and a Marker Run carries no other style. When the
Translation's info sets `add_space_before_footnote_marker`, the marker gets one
space before it unless the text already ends in whitespace. An empty or unclosed
marker is dropped, as `ParseMarkup` hides it.

`ParseMarkup` keeps hiding the marker, so Search and snippets stay free of
`[1]`. The Commentary modal prefixes each note of a Footnotes Module
(`info.is_footnotes`) with the marker that answers it; older Commentary Modules
without the `marker` column are read as before, with no marker.

**Blocked by:** nothing.

**Status:** done

- [x] `Run` has a `Marker` flag, kept through wrapping and joining.
- [x] `ParseVerse` emits the marker text; `ParseMarkup` still hides it.
- [x] `add_space_before_footnote_marker` inserts the space only when it is needed.
- [x] Empty and unclosed markers do not leak text into the Verse.
- [x] Search and snippets read the Verse without its markers.
- [x] A Footnotes module shows each note behind its marker, in the link style.
- [x] `n` opens the Footnote module's tab, preferring the Translation's own; `Enter` opens it when nothing else speaks to the Verse.
- [x] Commentary Modules without a `marker` column still load.
- [x] Module and TUI tests cover the parser, the spacing flag, search and the note prefix.
