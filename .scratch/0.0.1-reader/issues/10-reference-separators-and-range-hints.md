# 10: Reference separators and the range hints

**What to build:** The Go to Reference modal takes a Chapter and a Verse divided by any of `:`, a comma, a period, or a space, so `John 3:16`, `John 3,16`, `John 3.16`, and `John 3 16` all read the same. A separator is what tells a trailing number apart from a Book name: "3" on its own is still a Book name — it narrows the list to the numbered Books — while "3:16" and "3 16" are a location in the Book the input names, or in the translation's first Book when it names none. Multi-word Book names survive because the location is only the run of trailing words holding nothing but digits and separators, so `1 Samuel 3 16` stays 1 Samuel, Chapter 3, Verse 16.

The modal also stops showing the same example whatever is typed. While no Book is named it names the accepted shapes; once its input names a Book it shows that Book's range of Chapters — `Exodus 1..2` — and once a separator closes the Chapter, that Chapter's range of Verses: `Genesis 3:1..31`. A Chapter still being typed is never taken as final, because `3` may be on its way to `33`, so the Verses wait for the separator that closes it; a Chapter the Book does not have falls back to the Book's range. Both are facts about the Active translation rather than examples, so they carry no "For example" and no translation.

A rejected Reference keeps that range line beside the reason, which is where an out-of-range Verse belongs: typing `Genesis 3:333` says the Reference does not exist and the line under it already says how far Genesis 3 goes. The rejection needs no message of its own, and the range line is colored as a warning as soon as a number the translation does not have is typed — a Chapter outside the Book, or a Verse outside the Chapter it closed — so the reader is told before `Enter`, which is left for going rather than for finding out.

**Blocked by:** none.

**Status:** done

- [x] `module.SplitReference` separates the location from the Book name by taking the trailing words that hold nothing but digits and separators, and treats text naming no Book as a location when it carries a separator or spans more than one word.
- [x] `module.ParseReference` accepts `:`, `,`, `.`, and a space as the Chapter/Verse separator, alongside `John`, `John 3`, and `1 Samuel 3 16`.
- [x] The modal filters and completes by the Book name it already resolves, and shows the Book's Chapter range while no Chapter is closed.
- [x] A separator after the Chapter turns the line into that Chapter's Verse range — the space included, which the split drops as soon as it separates the Book name — and a rejected Verse leaves that line visible under the message.
- [x] A Chapter or Verse the Active translation does not have turns that range line red while it is typed, and a number it does have leaves it alone.
- [x] Tests cover the four separators through the TUI, the Book names that contain numbers and spaces, the Chapter range, the Verse range, the rejected Verse that keeps it, and the flagged number the Book lacks.

## Notes

The TUI's own splitter is gone: `module.SplitReference` is now the single definition of where a Book name ends, so filtering, `Tab` completion, and parsing cannot drift apart. The Verse range comes from the loaded Chapter, so it describes the Active translation rather than an ideal Bible.
