# 13: Show the Module's own Book names

**What to build:** Every Book the interface shows is named by the Active translation rather than by the Canonical list: the Book rows read `Gn · Génesis` for NBLA, the top bar and every Reference read `Génesis 1:1`, and completing a Reference writes the displayed name. The Canonical name remains the Book's identity, the fallback when a Module carries no name, and the form persisted state is written in. Input still matches any of the three names, so an English name reaches a Book in a Spanish translation and the other way around.

**Blocked by:** none.

**Status:** done

- [x] Book rows show the Module's own short and long names, falling back to the Canonical name when the Module carries no long name.
- [x] The top bar, Search results, History, the Commentary title, and a Subheading's cross-reference label name the Book the way the Active translation does, including a History entry loaded from an earlier run.
- [x] `Tab` and the space shortcut complete with the displayed name; the Canonical name is still what `ParseReference` receives.
- [x] The Book list filters by the Canonical name, the Module's short name, and its long name.
- [x] The Go to Reference example and rejection message name the reader's Book in the Module's own words.
- [x] Tests cover a Spanish fixture for display, filtering, completion, jumping, Search, History, and the modal text.
- [x] The spec, `GLOSSARY.md`, and ADR-0004 record the decision.

## Notes

The research against the MyBible Android app set the direction: it names Books from the Module's `book_name`/`short_name` columns, while a separate multilingual alias table is used only to recognize typed references. That alias table and the app's content are not copied here; this implementation keeps its own Canonical list and uses the Module's names for display.

The Canonical names in the loaded Books come from the module layer's Canonical lookup, so a Book held only by number — a History entry from a previous run, the Trail after a translation change — is resolved through the Active translation before it is shown. `Reference.String` therefore keeps its Canonical form for config and `ParseReference`, while the new `DisplayString` is used only where the reader looks.
