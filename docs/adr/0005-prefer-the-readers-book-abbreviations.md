# Let the reader prefer its own Book abbreviations

The reader may keep one set of Preferred abbreviations and use it instead of the Module's own short name for a Book in the Book lists and in input matching. It is off by default; when on with no set, the reader asks which Translation to copy the abbreviations from, writes them beside `config.toml` in MyBible's desired-abbreviation JSON format, and applies them to every Translation (or only to Translations in one language, when the set names one). This amends ADR-0004: the Module still supplies the Book's long name and References, but the reader may choose the short form, so NBLA and RVR60 no longer disagree.

## Considered Options

- **Keep the Module's abbreviations everywhere (ADR-0004)** — the Book is always named the way the Active translation prints it, but switching between Spanish editions changes `Ex` to `Éx` and `1S` to `1Sam`.
- **One Preferred set (chosen)** — consistent abbreviations across Translations, seeded once from a Translation and edited by hand.
- **MyBible's several active sets, ordered by name and filtered by language** — more flexible, but the reader would need a set manager and an active list that Android keeps outside the set files.
- **Store the set in `config.toml`** — one format, but gives up interop with the Android app and rewrites the whole map on every state save.

## Consequences

- The preferred set lives in `preferred.abbreviations.json` beside `config.toml`, in MyBible's `DesiredAbbreviations` schema, so a set can be copied to or from the Android app.
- The set is seeded with the chosen Translation's Books and an empty language, so it applies to every Translation. A non-empty `language` in the file limits it to Modules in that language, the way MyBible treats a set.
- A Book with no entry keeps the Module's short name, and the Canonical name, long name, and Module short name still work as input, so enabling the set never makes an existing Reference untypeable.
- References, History, Search, and the top bar keep the Module's long names; only the Book list abbreviation changes.
- With the option off, no abbreviation file is read and the reader behaves exactly as before.
- A malformed file is reported in the Log and left alone; the reader falls back to the Module's abbreviations rather than overwriting a hand-edited file.
