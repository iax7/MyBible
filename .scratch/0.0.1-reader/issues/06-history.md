# 06: History

**What to build:** The user can retrace their steps. Navigation jumps push onto a back/forward stack; a plain cursor move does not. `H` opens a "History" modal listing recently visited References, most recent first, and choosing one jumps to it. History is kept for the session and the last-read Reference continues to be restored on open (as delivered in ticket 04).

**Blocked by:** 04 (Navigate the Bible).

**Status:** ready-for-agent

- [ ] A navigation jump pushes the previous Reference onto the back stack.
- [ ] Back and forward move through visited References without duplicating the current one.
- [ ] Moving the Verse cursor alone does not create a history entry.
- [ ] `H` opens the History modal listing recent References, most recent first; `Enter` jumps to the chosen Reference.
- [ ] A TUI-model test performs a sequence of jumps, asserts the history contents and order, and asserts that jumping from the modal reaches the expected Reference.
