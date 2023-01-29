package crawler

import (
	"github.com/HuguesGuilleus/isty-search/crawler/database"
	"github.com/HuguesGuilleus/isty-search/keys"
)

// Call each Page with a HTML from the database call is sequenticaly.
func Process(db *crawldatabase.Database[Page], processList ...interface{ Process(*Page) }) error {
	return db.ForHTML(func(key keys.Key, page *Page) {
		if page.Html == nil {
			return
		}
		for _, process := range processList {
			process.Process(page)
		}
	})
}

// A function that can be used by Process function.
type ProcessFunc func(page *Page)

func (process ProcessFunc) Process(page *Page) {
	process(page)
}
