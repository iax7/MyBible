# MyBible TUI — Reading speed and startup

Status: ready-for-agent

## Problem Statement

The reader feels slower than it is. Measured against a real library
(`RV'60+`: 31,103 Verses, 8.7 MB of verse text; `LBLA+` behaves the same), three
places stall while everything else is already instant:

| Path | Measured |
| --- | --- |
| Discover the Module directory | 0.6–0.8 ms |
| Read the Active translation into memory | **460–520 ms** |
| — of which: scan `verses` rows | 71 ms |
| — of which: `ParseMarkup` | 113 ms |
| — of which: `searchable` | **329 ms** |
| Type a broad letter in Search (`j`, 18,773 hits) | **487 ms** |
| Type a narrow query (`jesucristo`, 184 hits) | 6–10 ms |
| Render one frame at 120×40 | 1–3 ms |
| Heap after loading | 14–22 MB |

So:

- **Startup waits half a second before the first frame.** `tui.New` loads the
  Active translation synchronously, and `main` only starts the program once `New`
  returns.
- **Switching translation freezes the interface for the same half second**, inside
  `Update`, with no sign that anything is happening.
- **Typing a broad letter in the Search modal blocks the interface for ~0.5 s.**
  `Search` builds a snippet for every hit, and `snippet` normalizes the Verse text
  rune by rune to find where the query sits, so cost grows with the number of
  hits rather than with what the modal can show.
- **Loading is three sequential transformations over every Verse**
  (`ParseMarkup`, `collapse`, `Normalize`, each allocating), so the same text is
  parsed twice and normalized through NFD per Verse.

None of this is a rendering problem: a frame is 1–3 ms, and discovery is under a
millisecond. The work is concentrated in normalization at load and in eager
snippets at search time.

## Solution

Keep the reader's behaviour and make it stop stalling:

- **Search answers within a frame** by collecting the first 500 hits in canonical
  order and telling the reader, in the modal's hint line, that the list is capped
  and the query should be narrowed.
- **Loading does one pass per Verse** instead of three, with no intermediate
  strings, and carries a benchmark so the cost stays visible.
- **Loading happens off the `Update` loop**, behind a loading state, both at
  startup and when the Active translation changes.

The first is a bug fix a reader feels immediately; the other two are what makes
opening and switching feel like a TUI rather than a batch job.

## User Stories

1. As a reader, I want typing in the Search modal to keep up with my keystrokes,
   so that a broad first letter does not lock the reader for half a second.
2. As a reader, I want the Search modal to tell me when it is showing only the
   first matches, so that a capped list never reads as "these are all of them".
3. As a reader, I want the reader to open with the first frame as soon as I press
   Enter, so that I am not staring at a blank terminal while a Bible loads.
4. As a reader, I want switching translation to show that the new text is loading
   without freezing the interface, so that I know the keypress landed.
5. As a maintainer, I want a benchmark that loads a Bible-sized Module, so that
   load and search cost can be measured instead of guessed.
6. As a maintainer, I want the load cost to stay in `internal/module`, so that the
   TUI keeps depending on values rather than on SQLite.

## Implementation Decisions

**Search is capped, in canonical order.** `Search` stops walking the loaded text at
500 hits instead of returning every match. Canonical order is already the order
`Search` returns, so the cap keeps the earliest matches and narrowing the query is
what reaches the rest. The `Reader` seam keeps its signature: the modal knows the
cap was reached because it got exactly 500 results. The Search modal's existing
hint line gains the count in that case ("showing the first 500 matches — type more
to narrow") and stays as it is otherwise. Snippets stay eager, now bounded to 500,
which is a few milliseconds.

**One pass per Verse at load.** `searchable` stops composing `collapse` and
`Normalize` over an intermediate display string. It walks the runs `ParseMarkup`
already produces once, lowercasing, folding accents, and collapsing whitespace
runs into the output builder directly. A Verse whose text is plain ASCII skips the
accent folding. `Normalize` keeps its current behaviour for queries, because a
query is short and readability matters more than its cost.

**Load runs as a command.** `tui.New` no longer loads the Active translation. It
sets a loading state naming the translation and returns a model whose `Init`
returns a `tea.Cmd` that loads the text and delivers it as a message. The same
command is used when the Active translation changes, so both paths show the same
loading state. A load result arriving for a translation that is no longer active
(or after a newer load was started) is dropped. A load error still becomes a
warning, and a Translation that is loading renders the reading pane's existing
empty state with a loading line naming it.

**The benchmark generates its own Module.** A benchmark in `internal/module` builds
a Bible-sized fixture with `modtest` (66 Books, ~31,000 Verses, accented text) in
the test's temporary directory, then reports load and search time. No committed
binary, no environment variable, and it runs on `go test -bench`.

## Testing Decisions

- **Search cap.** A TUI-model test types a query that matches more than the cap and
  asserts the modal's hint names the cap and the number shown; a query under the
  cap keeps today's hint. A module test asserts `Search` returns at most 500 hits
  and that they are the first 500 in canonical order.
- **Load off the loop.** A TUI-model test asks for a translation switch, asserts the
  model reports loading and that `Update` returned a command, delivers the load
  message, and asserts the text arrives and the loading state clears. Another
  delivers a stale load for a translation that is no longer active and asserts it
  is dropped.
- **Normalization.** The existing `Normalize` and search tests are the behaviour
  contract: case- and accent-insensitivity must not change. The benchmark is the
  measurement, not an assertion — no test asserts a wall-clock time.

## Out of Scope

- A search index, SQLite FTS, or an inverted index.
- Caching the normalized text on disk.
- Loading in parallel with rendering the first Chapter, or streaming Verses in.
- Cancelling an in-flight load (the stale-result check is enough).
- Any change to what a search matches, or to the order results appear in.
- Comments and Commentaries loading, which follow the same shape but are a
  separate change.
