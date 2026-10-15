# 01: Scaffold, CI, and the app shell

**What to build:** The repository exists and builds: a Go module `github.com/iax7/mybible` under the MIT license at version 0.0.1, with CI running build, tests, and lint. The binary launches a full-screen Bubble Tea app: the reading pane fills the screen, a top bar lists the available commands with their keys plus a placeholder for the current Reference, `Ctrl-B` toggles the top bar, `?` opens a Help modal listing the keybindings, and `q` quits. Launching and quitting the reader works end to end.

**Blocked by:** None (can start immediately).

**Status:** done

- [x] `go build` produces a single binary named `mybible`; `go test ./...` runs; the linter runs clean.
- [x] CI (GitHub Actions) runs build, tests, and lint on push.
- [x] Launching the binary shows a full-screen reading pane with a top bar.
- [x] The top bar lists each command with the key that invokes it, and shows a Reference placeholder.
- [x] `Ctrl-B` hides and shows the top bar.
- [x] `?` opens a Help modal listing the keybindings; `Esc` closes it.
- [x] `q` quits cleanly.
- [x] A TUI-model test drives the model with keystrokes and asserts the rendered view for the shell, the toggle, the Help modal, and quit.
