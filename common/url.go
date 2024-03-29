package common

import (
	"fmt"
	"net/url"
	"sort"
)

// Parse one URL, panic if error.
// Use it only in test.
func ParseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

// Parse many URL, panic if error.
// Use it only in test.
func ParseURLs(args ...string) []*url.URL {
	urls := make([]*url.URL, len(args))
	for i, s := range args {
		u, err := url.Parse(s)
		if err != nil {
			panic(fmt.Sprintf("Wrong syntax for %q on index %d: %v", s, i, err))
		}
		urls[i] = u
	}
	return urls
}

// Transform all url as string and sort it.
func URL2String(urls []*url.URL) []string {
	s := make([]string, len(urls))
	for i, u := range urls {
		s[i] = u.String()
	}
	sort.Strings(s)
	return s
}
