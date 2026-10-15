# Use Go with Bubble Tea v2 for the TUI

We build the reader in Go on Bubble Tea v2 + Lip Gloss v2 + Bubbles v2, with a pure-Go SQLite driver (`modernc.org/sqlite`), compiled to a single static binary. This is the stack opencode uses — the look we are matching — and it gives the fastest iteration, trivial cross-compilation, and one-file distribution. The maintainer of this project is an agent, which weights iteration speed heavily.

## Considered Options

- **Rust + ratatui** — renders large corpora faster and keeps native terminal copy/paste, but needs more assembly (modals, tabs, fuzzy search, diff views) and has slower builds.
- **Python + Textual** — richest built-in widget set and least code, but cannot ship a single static binary.

## Consequences

- We use a pure-Go SQLite driver so the binary stays CGO-free and cross-compiles without a C toolchain.
- Long chapters are trimmed to the viewport ourselves; Bubble Tea has no immediate-mode virtualization.
