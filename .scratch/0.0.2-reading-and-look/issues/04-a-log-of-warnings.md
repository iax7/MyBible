# 04: A Log of warnings

**What to build:** Warnings leave the reading pane. The top bar gains a `Log`
entry — key `L`, a warning emoji in the theme's warning colour — that appears
**only when there are warnings**. `L` opens a Log modal listing every warning as a
bullet; the modal carries no filter line and typing does not narrow it. `Log` is
listed in the Help modal too.

**Blocked by:** 03 (both change the bar and the modal frame).

**Status:** done

- [x] The reading pane renders no warning text.
- [x] The bar shows `Log` only when there are warnings; a clean run shows none.
- [x] `L` opens the Log modal and lists each warning as a bullet.
- [x] The Log modal has no filter line; typing does not narrow it; `esc` closes it.
- [x] `Log` appears in the Help modal.
- [x] A TUI-model test asserts the entry's absence and presence, and the modal's list.
