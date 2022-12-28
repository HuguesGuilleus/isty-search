package crawler

import (
	"github.com/HuguesGuilleus/isty-search/crawler/database"
	"golang.org/x/exp/slog"
)

// Call each Page with a HTML from the database call is sequenticaly.
func Process(db *crawldatabase.Database[Page], logger *slog.Logger, processList ...interface{ Process(*Page) }) error {
	defer logger.Info("%end")
	return db.ForHTML(func(key crawldatabase.Key, page *Page, progress, total int) {
		if page.Html == nil {
			return
		}
		for _, process := range processList {
			process.Process(page)
		}
		logger.Info("%", "%i", progress, "%len", total)
	})
}

// A function that can be used by Process function.
type ProcessFunc func(page *Page)

func (process ProcessFunc) Process(page *Page) {
	process(page)
}
