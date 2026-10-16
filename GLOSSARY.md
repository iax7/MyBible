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

**Reference**:
The address of a Verse or a range of Verses — book, chapter, verse — used to name a location in navigation history and search results.
_Avoid_: Citation, location, pointer

**Active translation**:
The Translation currently displayed. The reader shows exactly one at a time.
_Avoid_: Current version, selected module

**Module directory**:
The filesystem folder scanned at startup for Modules. Overridable in config; when config names no folder the reader falls back to its own default.
_Avoid_: Library, module path, store

**History**:
The back/forward stack of visited References, plus the last-read position restored on open.
_Avoid_: Log, recents

**Log**:
The warnings the reader collected — Modules it could not read, state it could not save — shown in a modal of their own.
_Avoid_: Errors, messages
