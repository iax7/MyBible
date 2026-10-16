# 11: A paragraph mark that opens a Verse

**What to build:** A Verse whose text begins with a paragraph or line break renders on its number, not under an empty line of it. MyBible modules mark the Verse that starts a paragraph by starting its text with a break, so in one Spanish Bible 7,480 of 31,103 Verses — 24%, spread over all 66 Books and 1,187 of 1,189 Chapters — read as a lone verse number, a blank line, then the text. The reading pane drops that opening break: it is the layout code that knows the verse number belongs on the first line of the Verse, and a break with nothing before it says nothing. Interior breaks keep their meaning, since a paragraph or line break in the middle of a Verse is exactly what the reader is meant to see.

**Blocked by:** none.

**Status:** done

- [x] `renderLines` trims the newlines a rendered Verse opens with before it splits the Verse into lines, so its number never sits on a line of its own.
- [x] Interior `<pb/>` and `<br/>` still break the Verse, and the existing paragraph-break test still passes.
- [x] A test drives the reading pane with a Verse opening on each break tag, one with an interior break, and one control Verse, and fails on the unfixed code.

## Notes

Evidence, read-only, from `NBLA'05.SQLite3`: Job 4:1 is `<pb/>Entonces respondió Elifaz, el temanita:` and 4:2 is `<pb/><t>«Si alguien…</t>`, while the control 4:3 opens on `<t>`. Across the module, 7,480 Verses start with `<pb/>`, none with `<br/>`, and none with whitespace. RVR60 has 6,368, LBLA 5,713, LBLA+ 2,134 and RV'60+ none, so this is a convention of the format and of the modules that mark paragraphs, not a quirk of one file.

The fix is in the reading pane rather than in `ParseMarkup`, which stays a faithful parser of the markers, or in the loaded text, which keeps the Module's markup as it came. Verse text reaches the screen in only two places, and Search already collapsed the break away through `strings.Fields`, so nothing else needed changing; the empty line was also inside the cursor Verse's line span, which the same fix removes.
