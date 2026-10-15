# 07: Themes and languages

**What to build:** The reader fits the user's terminal and their language. It ships a dark theme and a light theme, selectable in config, with individual colors overridable in the same config. The words-of-Jesus red and the other semantic styles follow the active theme. Interface strings are embedded English and Spanish maps inside the binary with no i18n framework; the language resolves from config, then the system locale, then English. The top bar, all modals, the Help modal, and the empty-state message are all translated.

**Blocked by:** 03 (Read a Chapter).

**Status:** ready-for-agent

- [ ] A dark theme and a light theme are selectable in config.
- [ ] Individual theme colors can be overridden in config; the words-of-Jesus red respects the override.
- [ ] Embedded English and Spanish string maps exist; no i18n framework is used.
- [ ] Language resolves from config, then the system locale, then English.
- [ ] The top bar, all modals, the Help modal, and the empty-state message appear fully translated in both languages.
- [ ] A TUI-model test asserts the rendered view in English and in Spanish for the same state.
