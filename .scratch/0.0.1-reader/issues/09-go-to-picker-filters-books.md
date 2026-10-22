# 09: The Go to Reference modal filters Books

**What to build:** `v` becomes one input that filters Books as you type and still takes a Chapter and a Verse after them. The modal lists only the Books its input names — nothing before the first keystroke, since listing every Book is what `b` is for — so `Jo` shows Job, Joel, Jonah, John, 1 John, 2 John, and 3 John and the reader picks one with the arrows. `Tab` writes the highlighted Book into the input, spaced so a Chapter typed next lands beside the Book's name and not on it (keeping any `Chapter:Verse` already typed), which is the only way the list and the input meet: `Enter` never guesses from the highlight. `Enter` applies what is written — a Book alone means its first Verse (1:1), a Chapter without a Verse means Verse 1, a `Chapter:Verse` means that Verse, and an input that names no Book is read as the translation's first Book, so pressing `Enter` on an empty input reads the first Book of the Bible. A Book named the way the Module itself names it (`Gn`, `Gen`, `Juan` in a Spanish Module) is resolved when exactly one Book of the Active translation answers to it; a name no Book has, or one that more than one Book claims, is rejected with the existing message rather than guessed at. Filtering matches the Module's own short and long names in addition to the reader's canonical name, ignoring letter case and accents, which makes `Gn` and `Jhn` find their Book on the first keystroke. All three names reach the Book as input; what the interface shows comes from the Module, as ADR-0004 later settled. The parser keeps rejecting nonsense: the space before the Chapter stays required, and `ParseReference` stays a pure canonical parser that now also accepts a Book on its own.

**Blocked by:** none.

**Status:** done

- [x] The Reference modal lists the Active translation's Books in Canonical order, colored by Book classification like the Select Book modal, and lists nothing until the input names a Book — listing them all is what Select Book is for.
- [x] `Tab` completes the highlighted row into the input, spacing the Book from the Chapter typed next (`Jo` → `Job `) and preserving a `Chapter:Verse` already typed (`Jo 3:16` → `Job 3:16`); it does nothing when no row matches.
- [x] `Enter` takes what is written: `John` → John 1:1, `John 3` → John 3:1, `John 3:16` → John 3:16, and an input naming no Book is read as the translation's first Book.
- [x] A Book named only by the Module's own short or long name is resolved when exactly one Book claims that name; an unknown or ambiguous name keeps the existing "not a Reference" message and the modal stays open.
- [x] `module.ParseReference` accepts a canonical Book name on its own as Chapter 1, Verse 1, and still rejects a Chapter with no space before it.
- [x] A loaded Book carries the Module's own short and long names, filled at load time, while the Canonical list carries none.
- [x] The modal's footer names `Tab`, and the example line shows the three accepted shapes.
- [x] Tests drive the modal from the TUI model: filtering, `Tab` completion, each `Enter` shape, the Module's own abbreviation, and the rejected name.

## Notes

Nothing is added to `GLOSSARY.md`: the change widens which names a Book answers to, which is a matching rule rather than a new term, and the glossary already fixes Book names as the reader's own. No ADR either: the decision is small, reversible, and has no alternatives worth recording.

ADR-0004 later reversed the display half of that decision: the Module's names are what the reader shows, while the Canonical name stays the identity and the fallback. The matching rules above are unchanged, so this issue's behaviour still holds; only the words `Tab` writes and the rows show changed.
