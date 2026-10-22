# Name Books the way the Active translation names them

The reader shows the name and abbreviation the Module itself gives a Book — `Gn · Génesis` for NBLA — in the Book lists, the Go to Reference completion, the top bar, Search results, History, the Commentary title, and the cross-reference labels under Subheadings. The Canonical name stays the Book's identity, the fallback when a Module carries no name, and the form in which state files are written and from which they are parsed.

## Considered Options

- **Show the Canonical name everywhere** — naming is identical in every Translation, but a Spanish Module under a Spanish interface puts an English name beside a Spanish abbreviation, and completion writes the English name back. This was the first design.
- **Show the Module's names** — the reader reads the Book as the translation prints it. Naming then varies by Translation, so the Book's number is the only stable identity.

## Consequences

- NBLA reads `Génesis` and KJV reads `Genesis` for Book 10. The number, not the name, is what state and navigation keep.
- Input still matches the Canonical name, the Module's short name, and its long name, so `Genesis 3:16` and `Génesis 3:16` reach the same Verse, whichever is typed.
- The persisted last-read Reference keeps the Canonical form, and History keeps the book number, so both survive a change of Translation.
- A Module that carries no long name falls back to the Canonical name in the rows and in every Reference; one that carries no short name shows the name alone.
