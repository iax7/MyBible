# 03: Wear the zellij look

**What to build:** The `dark` theme becomes zellij's default palette — background
`#0B0B0B`, text `#D7D7D7`, green `#96DC5A` for keys and accents, orange `#FF9146`
for warnings, red `#FF5F6E` for the words of Jesus, `#898989` for Verse numbers —
and `light` stays as it is. Book classification colours are kept everywhere a Book
is named. The top bar looks like zellij's: the Reference is filled with its Book's
colour, foreground the bar's colour, like a selected tab; keys are green and
labels light grey. Every modal is drawn by one `frame` helper: a coloured border
with the title in the top edge and `esc` at that edge's right end. Behaviour and
keys are unchanged.

**Blocked by:** nothing.

**Status:** done

- [x] The `dark` theme uses the zellij palette; `light` is unchanged.
- [x] Color overrides in config still win over the new palette.
- [x] Book classification colours still color Books in the lists and the bar.
- [x] The bar's Reference is a filled tab in the Book's colour; keys are green, labels grey.
- [x] Every modal is framed with its title in the top edge and `esc` at the right.
- [x] A modal's footer hint renders in a quieter grey than the body text, overridable in config.
- [x] A modal overlays the Chapter instead of replacing the screen; the reading pane stays visible around it.
- [x] Existing keybindings and modal behaviour are unchanged.
- [x] Theme and modal tests are updated to the new colours and frame.
