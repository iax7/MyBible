# 02: Load a translation in one pass and measure it

**What to build:** Reading a translation into memory stops transforming every
Verse three times. `searchable` walks the runs `ParseMarkup` already produces
once, lowercasing, folding accents, and collapsing whitespace runs directly into
one builder, with no intermediate display string and no `strings.Fields` slice.
A Verse whose text is plain ASCII skips accent folding entirely. `Normalize` keeps
its current behaviour, because queries are short and the function is easiest to
read as it is.

Measured today on `RV'60+` (31,103 Verses, 8.7 MB): load 460–520 ms, of which
71 ms is scanning the rows, 113 ms `ParseMarkup`, and **329 ms `searchable`**
(`ParseMarkup` again, then `collapse`, then `Normalize` through NFD). The same
Verse is therefore parsed twice and allocated for at three times.

**What it also adds:** a benchmark in `internal/module` that builds a Bible-sized
Module with `modtest` (66 Books, ~31,000 Verses, accented text) in the test's
temporary directory and reports load and search time. It runs with
`go test -bench` and needs no environment variable or committed file.

**Status:** ready-for-agent

- [ ] `searchable` produces the normalized form in one pass; `displayText` and
      `collapse` are no longer both called from it.
- [ ] Case- and accent-insensitivity does not change: `jesus` still finds `Jesús`,
      `genesis` still answers to `Génesis`, and a `<pb/>`/`<br/>` inside a Verse
      still collapses to a single space.
- [ ] `Normalize` itself keeps its current behaviour and tests.
- [ ] A benchmark under `internal/module` builds its own Bible-sized fixture with
      `modtest` and reports load and search time, with `b.ReportAllocs`.
- [ ] Load time on the reference library drops well below 300 ms, and the
      benchmark's own fixture shows the same order of improvement.
- [ ] No test asserts a wall-clock time; the benchmark is the measurement.

**Comments:**

- The reference library is the maintainer's `RV'60+.SQLite3` (31,103 Verses); the
  benchmark's generated fixture is what a contributor without a library can run.
- The target above is deliberately loose: the shape of the change is the point,
  not shaving the last millisecond.
