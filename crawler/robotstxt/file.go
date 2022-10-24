// Parse robots.txt and match url the rules. We silence ignore error to
// not create too noise.
//
// Specification: https://www.rfc-editor.org/rfc/rfc9309.html
package robotstxt

import (
	"net/url"
	"sort"
	"strconv"
	"strings"
)

const (
	keyAllow      = "allow"
	keyDisallow   = "disallow"
	keyCrawlDelay = "crawl-delay"
	keySitemap    = "sitemap"
	keyUserAgent  = "user-agent"
)

// A parsed robots.txt file.
type File struct {
	// Delay in second. The value is positive and no bounded.
	CrawlDelay int
	// Link to all sitemap.
	SiteMap []url.URL
	// Allow and disallow rules. The rule are sorted by encoded pattern
	// length decrement.
	Rules []Rule
}

var DefaultRobots = File{}

// Parse the robots.txt file content to create a new File.
func Parse(content []byte) (file File) {
	rules := parseLines(content)

	// Get global option sitemap
	for _, rule := range rules {
		switch rule[0] {
		case keyCrawlDelay:
			delay, err := strconv.Atoi(rule[1])
			if err == nil && delay > 0 {
				file.CrawlDelay = delay
			}
		case keySitemap:
			u, _ := url.Parse(rule[1])
			if u != nil && u.Scheme == "https" {
				file.SiteMap = append(file.SiteMap, *u)
			}
		}
	}

	// Get user-agent, allow and disallow rules.
	for _, rule := range filterLines(rules) {
		m := parseMatcher(rule[1], rule[0] == keyAllow)
		if m != nil {
			file.Rules = append(file.Rules, *m)
		}
	}

	return
}

// Filter and sort only allow and disallow rule.
func filterLines(rules [][2]string) (filtered [][2]string) {
	filtered = make([][2]string, 0, len(rules))

	middleOfGroup := true
	isCurrentUserAgent := true
	for _, rule := range rules {
		switch rule[0] {
		case keyUserAgent:
			if middleOfGroup {
				middleOfGroup = false
				isCurrentUserAgent = false
			}
			if rule[1] == "*" {
				isCurrentUserAgent = true
			}
		case keyDisallow, keyAllow:
			middleOfGroup = true
			if isCurrentUserAgent {
				filtered = append(filtered, rule)
			}
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i][1] > filtered[j][1]
	})

	return
}

// Split lines, remove comment and return a slice of [key, value]. The key is
// in lower case.
func parseLines(content []byte) (rules [][2]string) {
	for _, lineGroup := range strings.Split(string(content), "\n") {
		for _, line := range strings.Split(lineGroup, "\r") {
			line, _, _ = strings.Cut(line, "#")

			ruleKey, ruleValue, haveSeparator := strings.Cut(line, ":")
			if !haveSeparator {
				continue
			}

			rules = append(rules, [2]string{
				lowerRuleName(strings.TrimSpace(ruleKey)),
				strings.TrimSpace(ruleValue),
			})
		}
	}

	return
}

// Return the rule key in lower case.
func lowerRuleName(ruleName string) string {
	switch ruleName {
	case keyUserAgent, "User-agent", "USER-AGENT":
		return keyUserAgent
	case keyDisallow, "Disallow", "DISALLOW":
		return keyDisallow
	case keyAllow, "Allow", "ALLOW":
		return keyAllow
	case keySitemap, "Sitemap", "SITEMAP":
		return keySitemap
	case keyCrawlDelay, "Crawl-delay", "CRAWL-DELAY":
		return keyCrawlDelay
	default:
		return strings.ToLower(ruleName)
	}
}

// Check if a url are allow or not by this robots.txt. Use only the path and
// the query from u.
func (file *File) Allow(u *url.URL) bool {
	testedURL := u.Path
	if u.RawQuery != "" {
		testedURL += u.RawQuery
	}

	for _, rule := range file.Rules {
		if rule.match(testedURL) {
			return rule.Allow
		}
	}

	return true
}
