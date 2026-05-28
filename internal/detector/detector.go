package detector

import (
	"fmt"
	"regexp"
	"strings"
)

// ScriptureRef represents a parsed scripture reference.
type ScriptureRef struct {
	Book         string
	StartChapter int
	StartVerse   int // 0 means chapter-only
	EndChapter   int
	EndVerse     int // 0 means no end verse
}

func (r ScriptureRef) String() string {
	if r.StartVerse == 0 {
		return fmt.Sprintf("%s %d", r.Book, r.StartChapter)
	}
	if r.EndChapter != r.StartChapter {
		return fmt.Sprintf("%s %d:%d–%d:%d", r.Book, r.StartChapter, r.StartVerse, r.EndChapter, r.EndVerse)
	}
	if r.EndVerse != 0 && r.EndVerse != r.StartVerse {
		return fmt.Sprintf("%s %d:%d–%d", r.Book, r.StartChapter, r.StartVerse, r.EndVerse)
	}
	return fmt.Sprintf("%s %d:%d", r.Book, r.StartChapter, r.StartVerse)
}

// bookAliases maps lowercase alias → canonical name.
var bookAliases map[string]string
var refPattern *regexp.Regexp

func init() {
	type entry struct {
		canonical string
		aliases   []string
	}

	books := []entry{
		{"Genesis", []string{"gen", "gen.", "ge", "gn"}},
		{"Exodus", []string{"exod", "exod.", "ex", "exo"}},
		{"Leviticus", []string{"lev", "lev.", "le", "lv"}},
		{"Numbers", []string{"num", "num.", "nu", "nm", "nb"}},
		{"Deuteronomy", []string{"deut", "deut.", "dt"}},
		{"Joshua", []string{"josh", "josh.", "jos"}},
		{"Judges", []string{"judg", "judg.", "jdg", "jg"}},
		{"Ruth", []string{"ruth", "ru", "rth"}},
		{"1 Samuel", []string{"1 sam", "1 sam.", "1sam", "1sa", "1 sa"}},
		{"2 Samuel", []string{"2 sam", "2 sam.", "2sam", "2sa", "2 sa"}},
		{"1 Kings", []string{"1 kgs", "1 kgs.", "1kgs", "1 ki", "1ki"}},
		{"2 Kings", []string{"2 kgs", "2 kgs.", "2kgs", "2 ki", "2ki"}},
		{"1 Chronicles", []string{"1 chron", "1 chron.", "1chron", "1 chr", "1chr", "1 ch"}},
		{"2 Chronicles", []string{"2 chron", "2 chron.", "2chron", "2 chr", "2chr", "2 ch"}},
		{"Ezra", []string{"ezra", "ezr"}},
		{"Nehemiah", []string{"neh", "neh.", "ne"}},
		{"Esther", []string{"esth", "esth.", "est", "es"}},
		{"Job", []string{"job", "jb"}},
		{"Psalms", []string{"ps", "ps.", "psa", "psa.", "psalm", "psalms"}},
		{"Proverbs", []string{"prov", "prov.", "pro", "pr", "prv"}},
		{"Ecclesiastes", []string{"eccles", "eccles.", "eccl", "eccl.", "ecc", "ec", "qoh"}},
		{"Song of Solomon", []string{"song", "song of sol", "song of sol.", "sos", "ss", "cant"}},
		{"Isaiah", []string{"isa", "isa.", "is"}},
		{"Jeremiah", []string{"jer", "jer.", "je", "jr"}},
		{"Lamentations", []string{"lam", "lam.", "la"}},
		{"Ezekiel", []string{"ezek", "ezek.", "eze", "ezk"}},
		{"Daniel", []string{"dan", "dan.", "da", "dn"}},
		{"Hosea", []string{"hos", "hos.", "ho"}},
		{"Joel", []string{"joel", "jl"}},
		{"Amos", []string{"amos", "am"}},
		{"Obadiah", []string{"obad", "obad.", "ob"}},
		{"Jonah", []string{"jonah", "jon"}},
		{"Micah", []string{"mic", "mic.", "mc"}},
		{"Nahum", []string{"nah", "nah.", "na"}},
		{"Habakkuk", []string{"hab", "hab.", "hb"}},
		{"Zephaniah", []string{"zeph", "zeph.", "zep"}},
		{"Haggai", []string{"hag", "hag.", "hg"}},
		{"Zechariah", []string{"zech", "zech.", "zec", "zc"}},
		{"Malachi", []string{"mal", "mal.", "ml"}},
		{"Matthew", []string{"matt", "matt.", "mat", "mt"}},
		{"Mark", []string{"mark", "mrk", "mk"}},
		{"Luke", []string{"luke", "luk", "lk"}},
		{"John", []string{"john", "joh", "jn"}},
		{"Acts", []string{"acts", "act", "ac"}},
		{"Romans", []string{"rom", "rom.", "ro", "rm"}},
		{"1 Corinthians", []string{"1 cor", "1 cor.", "1cor", "1 co", "1co"}},
		{"2 Corinthians", []string{"2 cor", "2 cor.", "2cor", "2 co", "2co"}},
		{"Galatians", []string{"gal", "gal.", "ga"}},
		{"Ephesians", []string{"eph", "eph.", "ep"}},
		{"Philippians", []string{"phil", "phil.", "php"}},
		{"Colossians", []string{"col", "col.", "co"}},
		{"1 Thessalonians", []string{"1 thess", "1 thess.", "1thess", "1 thes", "1thes", "1 th", "1th"}},
		{"2 Thessalonians", []string{"2 thess", "2 thess.", "2thess", "2 thes", "2thes", "2 th", "2th"}},
		{"1 Timothy", []string{"1 tim", "1 tim.", "1tim", "1 ti", "1ti"}},
		{"2 Timothy", []string{"2 tim", "2 tim.", "2tim", "2 ti", "2ti"}},
		{"Titus", []string{"titus", "tit", "ti"}},
		{"Philemon", []string{"philem", "philem.", "phlm", "phm"}},
		{"Hebrews", []string{"heb", "heb."}},
		{"James", []string{"james", "jas", "jm"}},
		{"1 Peter", []string{"1 pet", "1 pet.", "1pet", "1 pe", "1pe", "1 pt", "1pt"}},
		{"2 Peter", []string{"2 pet", "2 pet.", "2pet", "2 pe", "2pe", "2 pt", "2pt"}},
		{"1 John", []string{"1 john", "1 joh", "1john", "1 jn", "1jn"}},
		{"2 John", []string{"2 john", "2 joh", "2john", "2 jn", "2jn"}},
		{"3 John", []string{"3 john", "3 joh", "3john", "3 jn", "3jn"}},
		{"Jude", []string{"jude", "jud", "jd"}},
		{"Revelation", []string{"rev", "rev.", "re", "revelation"}},
	}

	bookAliases = make(map[string]string, len(books)*4)
	for _, b := range books {
		bookAliases[strings.ToLower(b.canonical)] = b.canonical
		for _, a := range b.aliases {
			bookAliases[strings.ToLower(a)] = b.canonical
		}
	}

	refPattern = buildPattern()
}

func buildPattern() *regexp.Regexp {
	aliases := make([]string, 0, len(bookAliases))
	seen := make(map[string]bool)
	for a := range bookAliases {
		if !seen[a] {
			aliases = append(aliases, a)
			seen[a] = true
		}
	}
	for i := 0; i < len(aliases); i++ {
		for j := i + 1; j < len(aliases); j++ {
			if len(aliases[j]) > len(aliases[i]) {
				aliases[i], aliases[j] = aliases[j], aliases[i]
			}
		}
	}
	escaped := make([]string, len(aliases))
	for i, a := range aliases {
		escaped[i] = regexp.QuoteMeta(a)
	}

	bookGroup := strings.Join(escaped, "|")
	pattern := `(?i)\b(` + bookGroup + `)\s+(\d+)(?::(\d+)(?:\s*[-–—]\s*(\d+)(?::(\d+))?)?)?(?:\b|$)`
	return regexp.MustCompile(pattern)
}

// ParseReferences finds all scripture references in the given text.
func ParseReferences(text string) []ScriptureRef {
	matches := refPattern.FindAllStringSubmatch(text, -1)
	var refs []ScriptureRef

	for _, m := range matches {
		bookRaw := strings.TrimSpace(m[1])
		canonical, ok := bookAliases[strings.ToLower(bookRaw)]
		if !ok {
			continue
		}

		startChapter := atoi(m[2])
		startVerse := atoi(m[3])
		num4 := atoi(m[4])
		num5 := atoi(m[5])

		ref := ScriptureRef{
			Book:         canonical,
			StartChapter: startChapter,
			StartVerse:   startVerse,
		}

		if num5 != 0 {
			ref.EndChapter = num4
			ref.EndVerse = num5
		} else if num4 != 0 {
			ref.EndChapter = startChapter
			ref.EndVerse = num4
		} else {
			ref.EndChapter = startChapter
			ref.EndVerse = startVerse
		}

		refs = append(refs, ref)
	}
	return refs
}

func atoi(s string) int {
	if s == "" {
		return 0
	}
	n := 0
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n
}
