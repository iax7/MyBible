# MyBible

A terminal (TUI) Bible reader for Linux that reads MyBible (Android) `.SQLite3` modules directly, without importing them.

## Language

**Module**:
A single `.SQLite3` file distributed by the MyBible (Android) app, holding one kind of content.
_Avoid_: Database, asset, file

**Module type**:
Which kind of content a Module holds, encoded as a suffix in the file name — `RST.SQLite3` is a Bible, `Strong.dictionary.SQLite3` is a dictionary.
_Avoid_: Format, kind

**Translation**:
A Module of type Bible: one complete Bible text in one language or edition.
_Avoid_: Version, edition, Bible

**Canonical book list**:
The fixed, built-in ordered list of Bible books (the 66, plus optional deuterocanonicals) whose ids match MyBible's book numbers. Modules map into it by book number.
_Avoid_: Book table, book index, catalogue

**Canonical name**:
A Book's built-in name in the Canonical book list. It identifies the Book across Translations, names it when a Module carries no name, and is the form persisted state is written in.
_Avoid_: Internal name, default name

**Book**:
One entry in the Canonical book list. Its number identifies it across Translations; the name it is shown by comes from the Active translation.

**Book name**:
What the reader shows for a Book: the Module's own long name, abbreviated by the reader's Preferred abbreviation when it has one, or by the Module's own short name otherwise. The Canonical name is the fallback when a Module carries no name.
_Avoid_: Display name, label, title

**Preferred abbreviation**:
An abbreviation the reader keeps for a Book, replacing the Module's own short name when the reader's abbreviation set is enabled. The set is seeded from one Translation and edited by the reader; a Book without a Preferred abbreviation keeps the Module's short name, and a set that names a language applies only to Translations in that language.
_Avoid_: Canonical abbreviation, custom abbreviation, alias

**Book classification**:
The kind of writing a Book belongs to — the Pentateuch, the Historical books, the Poetic books, the Major prophets, the Minor prophets, the Gospels, Acts, the Epistles by Paul, the General Epistles, Revelation, or Non-canonical. MyBible colors its own book selector by it, and the reader colors a Book's name by it wherever the name is shown.
_Avoid_: Group, category, section, genre

**Chapter**:
One numbered division of a Book.

**Verse**:
One numbered unit of text within a Chapter. A Module stores exactly one row per Verse.

**Translator-inserted words**:
Words a Translation supplies to the base text for clarity or grammar, marked in the Module by `<i>`. The reader shows them italic and in a quieter color, so they read as supplied rather than original.
_Avoid_: Added text, italics, emphasis

**Subheading**:
A short heading introducing the Verses that follow it, drawn in the theme's quiet cyan. A Translation carries its own, or a companion Subheadings Module supplies them instead.
_Avoid_: Subtitle, section title, heading

**Commentary**:
What a Commentary module says about a Verse, a range of Verses, or a whole Chapter.
_Avoid_: Note, annotation, comment

**Commentary module**:
A Module of type commentaries: one edition's Commentary, read beside the Active translation and chosen by the reader rather than paired by file name.
_Avoid_: Notes module, commentary translation

**Active commentary**:
The Commentary module whose label the reading pane shows and whose text the Commentary modal opens first. One at a time.
_Avoid_: Selected commentary, current commentary

**Commentary label**:
The Active commentary's abbreviation, shown at the end of every Verse it covers.
_Avoid_: Badge, tag

**Footnote**:
A note a Translation attaches to a word or phrase of a Verse: a translation variant, a clarification, a literal rendering.
_Avoid_: Commentary, annotation, comment

**Footnote marker**:
The text a Translation writes inside `<f>…</f>` — `[1]`, `[1†]`, `ⓐ` — shown in the reading pane in the link style where the word it belongs to stands; `n` opens the note behind it.
_Avoid_: Superscript, symbol, callout

**Footnotes module**:
A Commentary module that declares itself Footnotes, whose rows answer the Footnote markers a Translation carries. One module holds the Footnotes of the Translation it is named after.
_Avoid_: Notes module, footnote translation

**Cross-reference**:
A link from a Verse or a range of Verses to another passage, carrying the votes of the witnesses that attest it.
_Avoid_: Citation, verse link

**Cross-reference set**:
A Module of type crossreferences: one edition's Cross-references for every Verse. The reader may read several sets at once.
_Avoid_: Reference module, citation set

**Active Cross-reference sets**:
The sets the reader turned on. Their Cross-references are merged, a target that several sets share counts once (the most voted link wins), and the reading pane marks every Verse they speak to.
_Avoid_: Selected sets, current sets

**Cross-reference marker**:
`[⇅N]` at the end of a Verse the Active Cross-reference sets speak to. N is the number of distinct targets, and the marker opens the References popup.
_Avoid_: Badge, reference count

**References popup**:
The modal opened with `x`: the cursor Verse's Cross-references, each with the passage it points to, read without leaving the Chapter.
_Avoid_: Cross-reference window, balloon

**Verse link**:
A Reference written inside Commentary text or under a Subheading, underlined and green so the reader can see it names a Verse.
_Avoid_: Citation, hyperlink

**Reference**:
The address of a Verse or a range of Verses — book, chapter, verse — used to name a location in navigation history and search results.
_Avoid_: Citation, location, pointer

**Reading position**:
The Reference of the cursor while the reader is open. A translation switch keeps it whenever the new Translation contains that Reference.
_Avoid_: Cursor position, current verse

**Active translation**:
The Translation currently displayed. The reader shows exactly one at a time.
_Avoid_: Current version, selected module

**Module directory**:
The filesystem folder scanned at startup for Modules. Overridable in config; when config names no folder the reader falls back to its own default.
_Avoid_: Library, module path, store

**Visit**:
One Reference reached by a deliberate jump — a Book, a Chapter, a Reference, a Search result, or a History entry. Moving the cursor or stepping a Chapter is not a Visit.
_Avoid_: Search, hit, step

**History**:
The Visits kept across restarts, grouped by day, that the reader can return to.
_Avoid_: Log, recents

**Trail**:
The Visits of the current session in the order they happened, retraced with back and forward. It does not survive a restart.
_Avoid_: Stack, history

**Last-read position**:
The Reference of the cursor when the reader quits, restored on the next open.
_Avoid_: Last reference, bookmark

**Log**:
The warnings the reader collected — Modules it could not read, state it could not save — shown in a modal of their own.
_Avoid_: Errors, messages
