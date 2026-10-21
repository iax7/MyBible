# 01: Search answers within a frame

**What to build:** Typing in the Search modal stops locking the reader. `Search`
walks the loaded text and returns at most **500** hits in canonical order instead
of every match, and it copies the Verse text out of the way it already does; the
modal's hint line says when the list is capped.

Measured today on `RV'60+`: the query `j` matches 18,773 Verses and blocks the
interface for **487 ms** because a snippet is built for every hit and `snippet`
normalizes the Verse rune by rune. `jesucristo` (184 hits) takes 6–10 ms. The cap
bounds the work to what the modal can show.

**Status:** ready-for-agent

- [ ] `Search` returns at most 500 `Result`s, the first 500 in canonical order.
- [ ] A query that hits the cap costs milliseconds, not hundreds of milliseconds.
- [ ] The Search modal's hint line reads `showing the first 500 matches — type more to narrow`
      (translated) exactly when it received 500 results, and keeps today's
      `enter to select` otherwise.
- [ ] An empty query still matches nothing; a query under the cap returns every
      hit as it does now.
- [ ] The `Reader` seam keeps its signature.
- [ ] A module test asserts the cap and that the kept hits are the first ones in
      canonical order.
- [ ] A TUI-model test asserts the capped hint and the unchanged hint below the cap.

**Comments:**

- The cap is a bound on work, not on what a reader can find: narrowing the query is
  how the rest is reached, which is what the hint says.
- `matchIndex` inside `snippet` is the remaining hot spot; at 500 hits it is a few
  milliseconds and does not need rewriting here.
