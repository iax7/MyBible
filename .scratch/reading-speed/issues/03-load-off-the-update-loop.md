# 03: Load off the Update loop

**What to build:** Opening the reader and switching translation stop blocking.
`tui.New` no longer loads the Active translation; it sets a loading state naming
the translation and returns a model whose `Init` returns a `tea.Cmd` that loads
the text and delivers it as a message. A switch uses the same command, so both
paths show the same state: the reading pane's existing empty state, plus a
loading line naming the translation.

Measured today on `RV'60+`: loading is 460–520 ms. At startup it delays the first
frame; on a switch it freezes the interface inside `Update`.

Three edges the change has to keep right:

- **The last-read Reference.** `New` currently restores it after loading. It must
  instead be remembered and applied when the load lands, or the reader reopens at
  the first Verse.
- **A late load.** Pressing `t` twice, or quitting while loading, must not let a
  result for a translation that is no longer active replace the current text or
  the saved Reference.
- **The saved position.** `persist` writes `LastReference` from the loaded Chapter
  and writes an empty string when there is none, so quitting during a load must
  not erase where the reader was.

**Status:** ready-for-agent

- [ ] `New` does not call `Load`; the first frame renders before the text arrives.
- [ ] `Init` returns a load command; the text arrives as a message and the reading
      pane shows the first Chapter as it does today.
- [ ] While loading, the pane keeps its empty state and adds one line naming the
      translation being loaded (translated in both languages).
- [ ] Switching translation with `t` shows the same state and does not freeze the
      modal.
- [ ] The last-read Reference is applied after the load lands, not dropped.
- [ ] A load result for a translation that is no longer active is ignored.
- [ ] Quitting while loading keeps the last-read Reference instead of erasing it.
- [ ] A load failure still becomes a warning, and the reader stays usable.
- [ ] A TUI-model test asserts loading state, command return, arrival, and the
      stale-result case.

**Comments:**

- `Load` already reports errors as values; nothing about it changes here.
- Cancelling an in-flight load is out of scope: the stale-result check is what
  keeps the two loads from racing.
