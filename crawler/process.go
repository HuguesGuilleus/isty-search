package crawler

import (
	"github.com/HuguesGuilleus/isty-search/crawler/db"
	"golang.org/x/exp/slog"
)

// Call each Page with a HTML from the database call is not concurently.
func Process(database *DB, slogHandler slog.Handler, processList ...interface{ Process(*Page) }) error {
	logger := slog.New(slogHandler)
	defer logger.Info("%end")

	return database.URLsDB.ForDone(func(key db.Key, i, total int) error {
		page, err := database.KeyValueDB.Get(key)
		if err != nil {
			return err
		} else if page.Html == nil {
			return nil
		}

		for _, p := range processList {
			p.Process(page)
		}

		logger.Info("%", "%i", i, "%len", total)

		return nil
	})
}
