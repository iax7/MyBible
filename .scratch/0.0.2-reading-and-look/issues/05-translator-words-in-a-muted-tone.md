# 05: Translator words in a muted tone

**What to build:** The words a Translation adds for clarity — marked `<i>` in the
Module — keep their italics and take a quieter color, the way a printed Bible sets
them below the base text's tone. The color is a theme color, `italics`, so it
ships with a value in both themes and can be overridden in config like the rest.
Where `<i>` falls inside `<J>`, the quieter color wins: the added word is not the
speaker's own, so it is not red. Italics in Commentary text are unaffected.

**Blocked by:** nothing.

**Status:** done

- [x] `<i>` Verse text renders italic in the theme's `italics` color.
- [x] `dark` uses `#8C8C8C`, `light` uses `243`.
- [x] A `colors.italics` override in config wins over the theme's value.
- [x] A word added inside `<J>` takes the italics color, not the red; the red around it stays.
- [x] Commentary italics keep their current look.
- [x] A TUI-model test asserts both themes, the override, and the `<J>` overlap.
