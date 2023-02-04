package index

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func GetVocab(s string) (words []string) {
	words = make([]string, 0)
	for _, word := range strings.FieldsFunc(s, splitWords) {
		if len(word) < 3 || sameRunes(word) {
			continue
		}
		word = strings.ToLower(word)
		words = append(words, word)
	}
	return
}

func splitWords(c rune) bool { return c != '-' && !unicode.IsLetter(c) && !unicode.IsNumber(c) }

// Return true is all rune of s is the same rune.
func sameRunes(s string) bool {
	first, _ := utf8.DecodeRuneInString(s)
	for _, r := range s {
		if r != first {
			return false
		}
	}
	return true
}
