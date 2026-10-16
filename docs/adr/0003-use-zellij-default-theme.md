# Use zellij's default theme for the interface

The reader is meant to sit quietly inside a terminal its maintainer already reads
in: zellij. We replace the `dark` theme with zellij's built-in `default` palette
(background `#0B0B0B`, text `#D7D7D7`, green `#96DC5A`, orange `#FF9146`, red
`#FF5F6E`) and draw modals the way zellij frames a dialog — a coloured border with
the title in the top edge — instead of inventing a palette. Book classification
colours, taken from MyBible's own book selector, are kept everywhere a Book is
named; in the top bar the Reference is filled with the Book's colour like a
selected tab.

## Considered Options

- **Keep the hand-picked `dark` theme** — it is ours to tune, but it is not the
  look the maintainer asked for, and it left the Reference colour fighting the bar.
- **Add a third `zellij` theme beside `dark` and `light`** — preserves the old
  dark look, but keeps a theme nobody wants and a name to maintain.

## Consequences

- The `dark` palette changes for every existing user of the reader; the `light`
  theme is unchanged, and color overrides in config still win.
- Modals overlay the reading pane instead of replacing it, so a Chapter stays
  visible around an open modal, the way a zellij floating pane sits over a pane.
- The palette is copied from zellij, so it must be updated by hand if zellij's
  `default` theme changes.
