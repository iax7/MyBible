package module

import (
	"sort"
	"strings"
)

// Book is one entry in the Canonical book list.
type Book struct {
	Number   int    // MyBible's book_number, stable across the reader
	Name     string // the reader's own name, so naming is the same in every Translation
	Chapters int    // Chapters present in a loaded Translation; 0 in the Canonical list

	// ShortName and LongName are the names the Module itself uses for the Book,
	// filled only in a loaded Translation. The reader never shows them: they
	// widen what an input accepts, so a Module's own naming — and its language —
	// can name a Book without changing what the interface calls it.
	ShortName string
	LongName  string
}

// Names is every name the Book answers to: the reader's own, then the Module's
// own when the Translation is loaded. Matching an input goes through all of
// them; the interface shows only the first.
func (b Book) Names() []string {
	names := []string{b.Name}
	for _, name := range []string{b.ShortName, b.LongName} {
		if name != "" && name != b.Name {
			names = append(names, name)
		}
	}
	return names
}

// Named reports whether text is one of the Book's names, ignoring letter case
// and accents.
func (b Book) Named(text string) bool {
	text = Normalize(strings.TrimSpace(text))
	for _, name := range b.Names() {
		if Normalize(name) == text {
			return true
		}
	}
	return false
}

// CanonicalBooks is the reader's built-in ordered list of Bible Books: the 66
// plus deuterocanonicals, ordered by MyBible's book_number. Modules map into it
// by book_number, so the reader never depends on a Module's own book names.
var CanonicalBooks = []Book{
	{Number: 10, Name: "Genesis"},
	{Number: 20, Name: "Exodus"},
	{Number: 30, Name: "Leviticus"},
	{Number: 40, Name: "Numbers"},
	{Number: 50, Name: "Deuteronomy"},
	{Number: 60, Name: "Joshua"},
	{Number: 70, Name: "Judges"},
	{Number: 80, Name: "Ruth"},
	{Number: 90, Name: "1 Samuel"},
	{Number: 100, Name: "2 Samuel"},
	{Number: 110, Name: "1 Kings"},
	{Number: 120, Name: "2 Kings"},
	{Number: 130, Name: "1 Chronicles"},
	{Number: 140, Name: "2 Chronicles"},
	{Number: 145, Name: "Prayer of Manasseh"},
	{Number: 150, Name: "Ezra"},
	{Number: 160, Name: "Nehemiah"},
	{Number: 165, Name: "2 Esdras"},
	{Number: 170, Name: "Tobit"},
	{Number: 180, Name: "Judith"},
	{Number: 190, Name: "Esther"},
	{Number: 192, Name: "Additions to Esther"},
	{Number: 220, Name: "Job"},
	{Number: 230, Name: "Psalms"},
	{Number: 240, Name: "Proverbs"},
	{Number: 250, Name: "Ecclesiastes"},
	{Number: 260, Name: "Song of Songs"},
	{Number: 270, Name: "Wisdom"},
	{Number: 280, Name: "Sirach"},
	{Number: 290, Name: "Isaiah"},
	{Number: 300, Name: "Jeremiah"},
	{Number: 305, Name: "Prayer of Azariah"},
	{Number: 310, Name: "Lamentations"},
	{Number: 315, Name: "Epistle of Jeremiah"},
	{Number: 320, Name: "Baruch"},
	{Number: 323, Name: "Song of the Three"},
	{Number: 325, Name: "Susanna"},
	{Number: 330, Name: "Ezekiel"},
	{Number: 340, Name: "Daniel"},
	{Number: 345, Name: "Bel and the Dragon"},
	{Number: 350, Name: "Hosea"},
	{Number: 360, Name: "Joel"},
	{Number: 370, Name: "Amos"},
	{Number: 380, Name: "Obadiah"},
	{Number: 390, Name: "Jonah"},
	{Number: 400, Name: "Micah"},
	{Number: 410, Name: "Nahum"},
	{Number: 420, Name: "Habakkuk"},
	{Number: 430, Name: "Zephaniah"},
	{Number: 440, Name: "Haggai"},
	{Number: 450, Name: "Zechariah"},
	{Number: 460, Name: "Malachi"},
	{Number: 462, Name: "1 Maccabees"},
	{Number: 464, Name: "2 Maccabees"},
	{Number: 466, Name: "3 Maccabees"},
	{Number: 467, Name: "4 Maccabees"},
	{Number: 470, Name: "Matthew"},
	{Number: 480, Name: "Mark"},
	{Number: 490, Name: "Luke"},
	{Number: 500, Name: "John"},
	{Number: 510, Name: "Acts"},
	{Number: 520, Name: "Romans"},
	{Number: 530, Name: "1 Corinthians"},
	{Number: 540, Name: "2 Corinthians"},
	{Number: 550, Name: "Galatians"},
	{Number: 560, Name: "Ephesians"},
	{Number: 570, Name: "Philippians"},
	{Number: 580, Name: "Colossians"},
	{Number: 590, Name: "1 Thessalonians"},
	{Number: 600, Name: "2 Thessalonians"},
	{Number: 610, Name: "1 Timothy"},
	{Number: 620, Name: "2 Timothy"},
	{Number: 630, Name: "Titus"},
	{Number: 640, Name: "Philemon"},
	{Number: 650, Name: "Hebrews"},
	{Number: 660, Name: "James"},
	{Number: 670, Name: "1 Peter"},
	{Number: 680, Name: "2 Peter"},
	{Number: 690, Name: "1 John"},
	{Number: 700, Name: "2 John"},
	{Number: 710, Name: "3 John"},
	{Number: 720, Name: "Jude"},
	{Number: 730, Name: "Revelation"},
	{Number: 780, Name: "Laodiceans"},
}

// canonicalByNumber indexes the Canonical list by book_number.
var canonicalByNumber = func() map[int]Book {
	byNumber := make(map[int]Book, len(CanonicalBooks))
	for _, b := range CanonicalBooks {
		byNumber[b.Number] = b
	}
	return byNumber
}()

// legacyNumbers maps book_number values from older MyBible format revisions
// onto their current canonical values. Modules written against either revision
// map onto the same Book.
var legacyNumbers = map[int]int{
	790: 145, // Prayer of Manasseh
	740: 165, // 2 Esdras
	200: 462, // 1 Maccabees
}

// BookByNumber resolves a MyBible book_number to a Canonical Book, tolerating
// numbers that changed across format revisions. State files refer to Books by
// number, so they keep working across Translations.
func BookByNumber(number int) (Book, bool) { return canonicalBook(number) }

// canonicalBook resolves a Module's book_number to a Canonical Book, tolerating
// numbers that changed across format revisions.
func canonicalBook(number int) (Book, bool) {
	if b, ok := canonicalByNumber[number]; ok {
		return b, true
	}
	if current, ok := legacyNumbers[number]; ok {
		b, ok := canonicalByNumber[current]
		return b, ok
	}
	return Book{}, false
}

func init() {
	sort.Slice(CanonicalBooks, func(i, j int) bool {
		return CanonicalBooks[i].Number < CanonicalBooks[j].Number
	})
}

// Classification is the kind of writing a Book belongs to. It is the grouping
// MyBible's own book selector colors by.
type Classification int

const (
	Pentateuch Classification = iota
	HistoricalBooks
	PoeticBooks
	MajorProphets
	MinorProphets
	Gospels
	Acts
	PaulineEpistles
	GeneralEpistles
	Revelation
	NonCanonical
)

// classificationNumbers is the MyBible book numbers in each Classification,
// following MyBible's own book table: the deuterocanonical Books take the
// section MyBible gives them — Tobit, Judith and 2 Esdras with the Historical
// Books, the Prayer of Manasseh and the Wisdom books among the Poetic ones,
// Baruch and the Epistle of Jeremiah with the Major prophets, Laodiceans with
// the General Epistles — while the Greek additions to Esther and Daniel, and the
// Maccabees, are the ones MyBible leaves non-canonical.
var classificationNumbers = map[Classification][]int{
	Pentateuch:      {10, 20, 30, 40, 50},
	HistoricalBooks: {60, 70, 80, 90, 100, 110, 120, 130, 140, 150, 160, 165, 170, 180, 190},
	PoeticBooks:     {145, 220, 230, 240, 250, 260, 270, 280},
	MajorProphets:   {290, 300, 310, 315, 320, 330, 340},
	MinorProphets:   {350, 360, 370, 380, 390, 400, 410, 420, 430, 440, 450, 460},
	Gospels:         {470, 480, 490, 500},
	Acts:            {510},
	PaulineEpistles: {520, 530, 540, 550, 560, 570, 580, 590, 600, 610, 620, 630, 640, 650},
	GeneralEpistles: {660, 670, 680, 690, 700, 710, 720, 780},
	Revelation:      {730},
	NonCanonical:    {192, 305, 323, 325, 345, 462, 464, 466, 467},
}

// classificationByNumber indexes classificationNumbers by book_number.
var classificationByNumber = func() map[int]Classification {
	byNumber := make(map[int]Classification)
	for classification, numbers := range classificationNumbers {
		for _, number := range numbers {
			byNumber[number] = classification
		}
	}
	return byNumber
}()

// ClassificationOf returns the Classification of the Book with this MyBible
// book number. An unknown book number is NonCanonical.
func ClassificationOf(bookNumber int) Classification {
	if current, ok := legacyNumbers[bookNumber]; ok {
		bookNumber = current
	}
	if classification, ok := classificationByNumber[bookNumber]; ok {
		return classification
	}
	return NonCanonical
}
