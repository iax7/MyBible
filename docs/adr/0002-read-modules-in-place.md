# Read MyBible modules in place; keep our own state in files

The reader opens the user's existing MyBible `.SQLite3` modules read-only and never imports or copies their content into a database of our own. Module content belongs to the user; our state (config and history) lives in XDG config and data files. This reuses the whole MyBible module ecosystem, avoids duplicating data, and sidesteps the fact that modules are mixed-license and mostly not redistributable.

## Consequences

- Every read must tolerate the module format's quirks: book numbers changed across format revisions, and optional tables may be absent.
- Anything that needs a stable shape is built in memory at startup, not written to disk.
- No module content is ever committed to this repository; tests use a small synthetic fixture.
