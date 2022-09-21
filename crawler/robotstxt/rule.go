package robotstxt

import (
	"net/url"
	"strings"
)

// One allow or disallow parsed rule.
type Rule struct {
	// Allow or Disallow
	Allow bool
	// Part before every *
	First string
	// Splited part between every *, without the first.
	Middle []string
	// End with '$', so must match exact
	EndMatch bool
}

// Parsed the patern to create a new rule.
func parseMatcher(pattern string, allow bool) *Rule {
	endMatch := false
	if l := len(pattern) - 1; l > 0 {
		switch pattern[l] {
		case '$':
			pattern = pattern[:l]
			endMatch = true
		case '*':
			pattern = pattern[:l]
		}
	}

	splitedPattern := strings.Split(pattern, "*")
	for i, a := range splitedPattern {
		a, err := url.QueryUnescape(a)
		if err != nil {
			return nil
		}
		splitedPattern[i] = a
	}

	return &Rule{
		Allow:    allow,
		First:    splitedPattern[0],
		Middle:   splitedPattern[1:],
		EndMatch: endMatch,
	}
}

func (rule *Rule) match(testedURL string) bool {
	if !strings.HasPrefix(testedURL, rule.First) {
		return false
	}
	testedURL = testedURL[len(rule.First):]

	for _, middle := range rule.Middle {
		found := false
		_, testedURL, found = strings.Cut(testedURL, middle)
		if !found {
			return false
		}
	}

	if rule.EndMatch {
		return testedURL == ""
	}
	return true
}
