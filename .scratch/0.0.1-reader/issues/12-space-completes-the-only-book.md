# 12: A space completes the only Book

**What to build:** In the Go to Reference modal, while the completion menu shows exactly one Book, a space writes that Book into the input the way `Tab` does — the name the reader shows for it, a space after it, the error cleared, the menu closed — so a Chapter can follow without reaching for `Tab`. With no match, or with more than one, the space is ordinary text and the menu stays, which is what lets `2 ` remain search text and narrow to `2 Samuel`, and what keeps the space before a Chapter working. Once the Book is completed the menu is gone and every later space is the separator the input already accepted, so `Genesis 3 16` still reads as Genesis, Chapter 3, Verse 16. `Tab` keeps completing with one or several matches. The footer keeps naming only `Tab`: the space is a shortcut, not a second documented way of completing.

**Blocked by:** none.

**Status:** done

- [x] With exactly one Book on the completion menu, a space completes it like `Tab`: the name the reader shows for it, trailing space, error cleared, menu hidden.
- [x] With no match, or with several, the space is typed into the input and the menu stays; `2 ` stays text and `2 Samuel` completes once `s` narrows the list to one.
- [x] After the Book is completed, a space is a Chapter/Verse separator again, so `Genesis 3 16` lands on Genesis 3:16.
- [x] `Tab` still completes with one or several matches; the footer and the other modals are unchanged.
- [x] The spec records the shortcut.

## Notes

The shortcut reads the same matches the menu renders, so it cannot complete a Book the reader cannot see. It stops once `completed` is set, the point where the menu closes; that is what keeps the space before a Chapter from completing the same Book again. No `GLOSSARY.md` entry: this widens a keystroke, not the vocabulary. No ADR: the decision is small, reversible, and has no alternatives worth recording.

When the reader later began showing the Module's own Book names (ADR-0004), completion kept writing what the menu shows, which is now that displayed name rather than the Canonical one. The tests use fixtures where the two agree, so the shortcut's coverage is unchanged.
