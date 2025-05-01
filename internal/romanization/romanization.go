package romanization

import (
	"unicode"

	"lrcsnc/internal/pkg/global"
	"lrcsnc/internal/pkg/structs"
)

// TODO: PROBABLY DONE BUT CHECK! there may be songs with mixed languages, so we may need to romanize each symbol separately

// Supported languages are listed here

// Language is a bitset of supported languages
// (each language has its own bit)
type Language uint8

const (
	LanguageDefault  Language = 0b0
	LanguageJapanese Language = 0b1 << iota
	LanguageKorean
	LanguageChinese
)

// Unicode range tables accordingly are listed here

var jpUnicodeRangeTable = []*unicode.RangeTable{
	unicode.Hiragana,
	unicode.Katakana,
}

var krUnicodeRangeTable = []*unicode.RangeTable{
	unicode.Hangul,
}

var zhUnicodeRangeTable = []*unicode.RangeTable{
	unicode.Diacritic,
	unicode.Ideographic,
	unicode.Han,
}

// Returns romanized lyrics (or the same lyrics if the language is not supported)
func Romanize(lyrics []structs.Lyric) {
	global.Config.M.Lock()

	if !global.Config.C.Lyrics.Romanization.IsEnabled() {
		return
	}

	global.Config.M.Unlock()

	for i := range lyrics {
		lang := getLang(lyrics[i])
		if lang == LanguageDefault {
			continue
		}

		if len(lyrics[i].Text) != 0 {
			rstr := sanitizeAfter(romanize(sanitizeBefore(lyrics[i].Text), lang)) // one hell of a mess lmao
			lyrics[i].Text = rstr
		}
	}
}

// getLang returns the detected languages that are supported by romanization module
// in form of the Language bit set.
func getLang(lyric structs.Lyric) (lang Language) {
	global.Config.M.Lock()
	defer global.Config.M.Unlock()

	lang = LanguageDefault

	// The idea of checking Japanese only by Hirogana and Katakana may backfire
	// if there are actually any songs where some lines contain only kanji.
	// But I don't seem to come up with any other solution of how to
	// differentiate between Japanese and Chinese when it comes to complex characters.
	// So, right now, kanji are detected as Chinese, but only if there were no
	// Japanese characters in the line already.

	if global.Config.C.Lyrics.Romanization.Japanese {
		if hasCharsOf(lyric.Text, jpUnicodeRangeTable) {
			lang |= LanguageJapanese
		}
	}
	if global.Config.C.Lyrics.Romanization.Korean {
		if hasCharsOf(lyric.Text, krUnicodeRangeTable) {
			lang |= LanguageKorean
		}
	}
	// Chinese romanization only enables if no Japanese has been detected
	if global.Config.C.Lyrics.Romanization.Chinese && lang&LanguageJapanese == 0 {
		if hasCharsOf(lyric.Text, zhUnicodeRangeTable) {
			lang |= LanguageChinese
		}
	}

	return
}

// Returns a romanized string based on the provided language
func romanize(str string, lang Language) (out string) {
	switch {
	case lang&LanguageJapanese != 0:
		out = Romanizers[LanguageJapanese].Romanize(str)
	case lang&LanguageKorean != 0:
		out = Romanizers[LanguageKorean].Romanize(str)
	case lang&LanguageChinese != 0:
		out = Romanizers[LanguageChinese].Romanize(str)
	}
	return out
}

func hasCharsOf(s string, rangeTable []*unicode.RangeTable) bool {
	for _, r := range s {
		if unicode.IsOneOf(rangeTable, r) {
			return true
		}
	}
	return false
}
