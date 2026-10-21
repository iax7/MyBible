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

**Book**:
One entry in the Canonical book list.

**Book classification**:
The kind of writing a Book belongs to — the Pentateuch, the Historical books, the Poetic books, the Major prophets, the Minor prophets, the Gospels, Acts, the Epistles by Paul, the General Epistles, Revelation, or Non-canonical. MyBible colors its own book selector by it, and the reader colors a Book's name by it wherever the name is shown.
_Avoid_: Group, category, section, genre

**Chapter**:
One numbered division of a Book.

**Verse**:
One numbered unit of text within a Chapter. A Module stores exactly one row per Verse.

**Subheading**:
A short heading introducing the Verses that follow it, held in a Module of its own rather than in the Bible text.
_Avoid_: Subtitle, section title, heading

**Commentary**:
What a Commentary module says about a Verse, a range of Verses, or a whole Chapter.
_Avoid_: Note, footnote, annotation

**Commentary module**:
A Module of type commentaries: one edition's Commentary, read beside the Active translation and chosen by the reader rather than paired by file name.
_Avoid_: Notes module, commentary translation

**Active commentary**:
The Commentary module whose label the reading pane shows and whose text the Commentary modal opens first. One at a time.
_Avoid_: Selected commentary, current commentary

**Commentary label**:
The Active commentary's abbreviation, shown at the end of every Verse it covers.
_Avoid_: Marker, badge, tag

**Verse link**:
A Reference written inside Commentary text, underlined and green so the reader can see it is a way to that Verse.
_Avoid_: Cross-reference, citation, hyperlink

**Reference**:
The address of a Verse or a range of Verses — book, chapter, verse — used to name a location in navigation history and search results.
_Avoid_: Citation, location, pointer

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
