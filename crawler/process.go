package crawler

import (
	"fmt"
	"github.com/HuguesGuilleus/isty-search/crawler/db"
)

type Processor interface {
	// Process a HTML page with notnil Html field.
	// Is not executed simultanly
	Process(*Page)
}

func Process(database *DB, processList []Processor) error {
	return database.URLsDB.ForDone(func(key db.Key, i, total int) error {
		page, err := database.KeyValueDB.Get(key)
		if err != nil {
			return err
		} else if page.Html == nil {
			return nil
		}

		fmt.Printf("[%6d/%6d] %s\n", i, total, page.URL.String())
		for _, p := range processList {
			p.Process(page)
		}
		return nil
	})
}
