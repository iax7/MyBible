# 02: Module directory, discovery, and Translation selection

**What to build:** The reader learns where the user's MyBible Modules live and which Translations they offer. The Module directory is read from config in the XDG config directory. At startup the reader scans that directory for `*.SQLite3`, determines each Module's type from its filename suffix, and reads its `info` table for the human title and language. Bible Modules are surfaced as Translations in a filterable "Select Translation" modal; picking one makes it the Active translation and that choice is remembered across runs. Non-Bible Module types are ignored. A Module that fails to open, or lacks a required table, is skipped with a visible warning rather than stopping the reader. When no Modules are found, the reader shows a clear message explaining how to set the Module directory.

**Blocked by:** 01 (Scaffold, CI, and the app shell).

**Status:** ready-for-agent

- [ ] The Module directory is read from config; a missing or unset value shows the empty-state message rather than failing.
- [ ] The reader scans the Module directory at startup and lists every Bible Module it finds as a Translation.
- [ ] Each Translation shows the title read from its `info` table.
- [ ] The "Select Translation" modal is filterable by typing and confirms with `Enter`.
- [ ] The Active translation is persisted and restored on the next launch.
- [ ] A corrupt Module, or one missing a required table, is skipped with a warning; the remaining Translations still load.
- [ ] Module files are opened read-only; a test asserts a Module is byte-identical after the reader opens it.
- [ ] A TUI-model test picks a Translation from a generated fixture and asserts it becomes active.
- [ ] A module-reader test asserts that a generated fixture is discovered and its type and title are read correctly.
