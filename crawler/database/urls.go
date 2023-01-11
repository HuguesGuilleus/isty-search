package crawldatabase

import (
	"github.com/HuguesGuilleus/isty-search/keys"
	"golang.org/x/exp/slog"
	"net/url"
	"strings"
)

// Load URLS from the data (url encoded as string sepatared by \n).
// Use the logger as warn when url parsing error cooure.
// Do not return URL with not accepted type in the mapMeta.
func loadURLs(logger *slog.Logger, data []byte, mapMeta map[keys.Key]metavalue, acceptedTypes []byte) []*url.URL {
	refusedTypes := [256]bool{}
	for i := range refusedTypes {
		refusedTypes[i] = true
	}
	for _, t := range acceptedTypes {
		refusedTypes[t] = false
	}

	lines := strings.Split(string(data), "\n")
	urls := make([]*url.URL, 0, len(lines))

	for line, s := range lines {
		if refusedTypes[mapMeta[keys.NewString(s)].Type] {
			continue
		}

		u, err := url.Parse(s)
		if err != nil {
			logger.Warn("db.openurl", "line", line, "err", err.Error())
			continue
		}

		urls = append(urls, u)
	}

	return urls
}
