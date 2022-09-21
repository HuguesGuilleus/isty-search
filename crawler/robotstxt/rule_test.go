package robotstxt

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMatcher(t *testing.T) {
	tester := func(pattern string, rule *Rule) {
		found := parseMatcher(pattern, true)
		assert.Equal(t, rule, found, pattern)
	}
	tester("/", &Rule{true, "/", []string{}, false})
	tester("/%64%69%72", &Rule{true, "/dir", []string{}, false})
	tester("*file.txt", &Rule{true, "", []string{"file.txt"}, false})
	tester("/dir/*", &Rule{true, "/dir/", []string{}, false})
	tester("/dir*file.txt", &Rule{true, "/dir", []string{"file.txt"}, false})
	tester("*file.txt$", &Rule{true, "", []string{"file.txt"}, true})
}

func TestAllower(t *testing.T) {
	tester := func(pattern, testedURL string, mustAllow bool) {
		allowed := parseMatcher(pattern, true).match(testedURL)
		if allowed != mustAllow {
			t.Errorf("rule %q match\t%q != %t", pattern, testedURL, mustAllow)
		}
	}

	tester("/", "/dir/subdir/file.txt", true)
	tester("/%64%69%72", "/dir/subdir/file.txt", true)
	tester("*file.txt", "/dir/subdir/file.txt", true)
	tester("/dir*file.txt", "/dir/subdir/file.txt", true)
	tester("/$", "/", true)

	tester("/$", "/dir/subdir/file.txt.odt", false)
	tester("*file.txt$", "/dir/subdir/file.txt.odt", false)
}
