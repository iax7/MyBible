# 01: Read a Chapter as prose

**What to build:** A Verse longer than the pane wraps at the pane's width instead
of running off the right edge. Wrapping breaks between words and keeps each Run's
style (italics, the words of Jesus); the Verse number keeps the first line and
every continuation line is indented by the number's width plus one space, so text
lines up under text. A word wider than the line is left whole rather than cut. The
Chapter opens with a centred `Chapter N` heading (translated) highlighted in the
Book's classification colour, rendered as the first line of the Chapter so it
scrolls with the text and stays out of the Verse cursor's span.

**Blocked by:** nothing.

**Status:** done

- [x] A long Verse wraps and the whole of it is visible.
- [x] Continuation lines are indented under the Verse's text, not under its number.
- [x] Wrapping keeps `<i>` italics and `<J>` red on the words they cover.
- [x] The Chapter opens with a centred `Chapter N` line in the Book's colour.
- [x] The heading scrolls with the Chapter and is not selectable by the Verse cursor.
- [x] A leading `<pb/>` still keeps the Verse number on the first line, wrapped.
- [x] A TUI-model test asserts the wrapped lines and the heading for a fixture.
