package index

import (
	"strings"
	"unicode"
)

func getVocab(s string) (words []string) {
	words = make([]string, 0)
	for _, word := range strings.FieldsFunc(s, splitWords) {
		word = strings.ToLower(word)
		words = append(words, word)
	}
	return
}

func splitWords(c rune) bool { return c != '-' && !unicode.IsLetter(c) && !unicode.IsNumber(c) }
