package crawler

import (
	"bytes"
	"fmt"
	"github.com/HuguesGuilleus/isty-search/bytesrecycler"
	"github.com/HuguesGuilleus/isty-search/crawler/db"
	"os"
)

type Processor interface {
	// Process a HTML page with no nil Html field.
	// Is not executed simultanly
	Process(*Page)
}

func Process(database *DB, processList ...Processor) error {
	bar := progessBar{
		buff: recycler.Get(),
		long: 80,
	}
	defer recycler.Recycle(bar.buff)
	defer bar.clean()

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

		bar.progress(i, total)

		return nil
	})
}

type progessBar struct {
	buff *bytes.Buffer
	long int

	previousPercentage int
}

func (bar *progessBar) progress(position, total int) {
	percentage := position * 100 / total
	if percentage == bar.previousPercentage {
		return
	}
	bar.previousPercentage = percentage

	fmt.Fprintf(bar.buff, " %3d %% ", percentage)

	rate := position * bar.long / total
	for i := 0; i < rate; i++ {
		bar.buff.WriteRune('█')
	}
	for i := rate; i < bar.long; i++ {
		bar.buff.WriteRune('░')
	}

	bar.buff.WriteString("\033[1G\033[1F")
	bar.buff.WriteTo(os.Stdout)
	bar.buff.Reset()
}

func (bar *progessBar) clean() {
	os.Stdout.WriteString("\033[1G\033[K")
}
