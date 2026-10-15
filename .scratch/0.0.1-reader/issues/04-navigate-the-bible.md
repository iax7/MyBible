# 04: Navigate the Bible

**What to build:** The user can move anywhere. A Verse cursor moves with `j`/`k` and always keeps the current Verse visible; `Ctrl-D`/`Ctrl-U` page up and down within a Chapter; `[` and `]` move to the previous and next Chapter; `g`/`G` jump to the top and bottom of the Chapter. A "Select Book" modal, a "Select Chapter" modal, and a "Go to Reference" modal (accepting input like `John 3:16`) let the user jump directly; all are filterable where a list is shown. The last-read Reference is persisted and restored when the reader reopens.

**Blocked by:** 03 (Read a Chapter).

**Status:** done

- [x] `j`/`k` move the Verse cursor by one Verse and keep the current Verse visible.
- [x] `Ctrl-D`/`Ctrl-U` page up and down within the Chapter.
- [x] `[`/`]` move to the previous/next Chapter, crossing Book boundaries at the ends.
- [x] `g`/`G` jump to the top and bottom of the Chapter.
- [x] The "Select Book" and "Select Chapter" modals list and filter their options; `Enter` jumps.
- [x] The "Go to Reference" modal accepts `John 3:16`-style input and jumps to that Verse; malformed input is rejected without leaving the reader.
- [x] The last-read Reference is persisted and restored on the next launch.
- [x] A TUI-model test drives navigation keys and modal jumps from a generated fixture and asserts the resulting Reference.
