package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/HuguesGuilleus/isty-search/common"
	"github.com/HuguesGuilleus/isty-search/crawler"
	"github.com/HuguesGuilleus/isty-search/crawler/database"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"github.com/HuguesGuilleus/isty-search/display"
	"github.com/HuguesGuilleus/isty-search/index"
	"github.com/HuguesGuilleus/isty-search/index/database"
	"github.com/HuguesGuilleus/isty-search/keys"
	"github.com/HuguesGuilleus/isty-search/sloghandlers"
	"golang.org/x/exp/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var actions = map[string]func(logger *slog.Logger, dbbase string) error{
	"crawl":         mainCrawl,
	"dbstats":       mainDBStatistics,
	"index":         mainIndex,
	"search":        mainSearch,
	"demo-vocab":    mainDemoVocab,
	"demo-pagerank": mainDemoPageRank,
	"demo-search":   mainDemoSearch,
}

func main() {
	db := flag.String("db", "db1", "dataBase directory path (can not exist)")
	flag.Parse()

	jsonLogCloser, jsonHandler := sloghandlers.NewFileHandler(*db, slog.InfoLevel)
	defer jsonLogCloser()
	logger := slog.New(sloghandlers.NewMultiHandlers(
		sloghandlers.NewConsole(slog.InfoLevel),
		jsonHandler,
	))

	defer func(begin time.Time) { logger.Info("duration", "d", time.Since(begin)) }(time.Now())
	if action := actions[flag.Arg(0)]; action == nil {
		fmt.Println("Unknown action. Possible actions are:")
		for name := range actions {
			fmt.Printf("\t%s\n", name)
		}
		os.Exit(1)
	} else if err := action(logger, *db); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func mainCrawl(logger *slog.Logger, dbbase string) error {
	config := crawler.Config{
		DBopener: crawldatabase.OpenWithKnow[crawler.Page],
		DBbase:   dbbase,
		Input:    common.ParseURLs("https://www.uvsq.fr/"),

		FilterURL: []func(*url.URL) bool{
			func(u *url.URL) bool {
				return u.Scheme != "https" || (u.Host != "uvsq.fr" &&
					u.Host != "cas.uvsq.fr" &&
					u.Host != "cas2.uvsq.fr" &&
					!strings.HasSuffix(u.Host, ".uvsq.fr"))
			},
		},
		FilterPage: []func(*htmlnode.Root) bool{
			func(root *htmlnode.Root) bool {
				switch root.Meta.Langage {
				case "":
					fallthrough
				case "fr", "fr_FR", "fr-FR", "fr-incl", "fr_incl":
					return false
				case "en":
					// No language log
					return true
				default:
					logger.Info("unknwo_lang", "lang", root.Meta.Langage)
					return true
				}
			},
		},

		MaxLength: 15_000_000,
		MaxGo:     10,

		MinCrawlDelay: time.Millisecond * 500,
		MaxCrawlDelay: time.Second * 10,

		Logger: logger,
	}

	ctx, ctxCancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer ctxCancel()

	return crawler.Crawl(ctx, config)
}

func mainDBStatistics(logger *slog.Logger, dbbase string) error {
	_, db, err := crawldatabase.Open[crawler.Page](logger, dbbase, false)
	if err != nil {
		return err
	}
	defer db.Close()

	db.Statistics().LogAll(logger)

	return nil
}

func mainDemoVocab(logger *slog.Logger, dbbase string) error {
	_, db, err := crawldatabase.Open[crawler.Page](logger, dbbase, false)
	if err != nil {
		return err
	}
	defer db.Close()

	countHTML := db.Statistics().Count[crawldatabase.TypeFileHTML]
	logger.Info("demo.db.page", "count", countHTML)

	counterWords := make(index.CounterVocab)
	if err := crawler.Process(db, counterWords); err != nil {
		return err
	}

	logger.Info("demo.vocab.stats",
		"wordsCount", len(counterWords),
		"wordsSum", counterWords.Sum())

	frequency := counterWords.Frequency()
	if len(frequency) > 100 {
		frequency = frequency[:100]
	}
	for _, w := range frequency {
		logger.Info("demo.vocab.frequency", "count", w.Count, "word", w.Word)
	}

	return nil
}

func mainDemoPageRank(logger *slog.Logger, dbbase string) error {
	_, db, err := crawldatabase.Open[crawler.Page](logger, dbbase, false)
	if err != nil {
		return err
	}
	defer db.Close()

	countHTML := db.Statistics().Count[crawldatabase.TypeFileHTML]
	logger.Info("demo.db.page", "count", countHTML)

	pageRank := index.NewPageRank()
	if err := crawler.Process(db, &pageRank); err != nil {
		return err
	}

	repeatition, scores := pageRank.Score(1000, 0.00_01)
	logger.Info("demo.pagerank.repeatition", "nb", repeatition)

	if len(scores) > 30 {
		scores = scores[:30]
	}
	for _, score := range scores {
		page, _, err := db.GetValue(score.Key)
		if err != nil {
			return err
		}
		logger.Info("demo.pagerank.x", "rank", score.Rank, "url", page.URL.String())
	}

	// pageRank.DevStats(logger)

	return nil
}

func mainIndex(logger *slog.Logger, dbbase string) error {
	_, db, err := crawldatabase.Open[crawler.Page](logger, dbbase, false)
	if err != nil {
		return err
	}
	defer db.Close()

	wordsIndex := make(index.VocabAdvanced)
	pageRank := index.NewPageRank()
	if err := crawler.Process(db, &pageRank, &wordsIndex); err != nil {
		return err
	}

	logger.Info("order.pagerank")
	_, scores := pageRank.Score(200, 0.00_001)
	globalOrder := make(map[keys.Key]float32, len(scores))
	for _, score := range scores {
		globalOrder[score.Key] = score.Rank
	}

	logger.Info("order.sort")
	for _, pages := range wordsIndex {
		sort.Slice(pages, func(i, j int) bool {
			return globalOrder[pages[i].Page] < globalOrder[pages[j].Page]
		})
	}

	logger.Info("order.save")
	indexdatabase.Store("words.db", wordsIndex)

	return nil
}

func mainSearch(logger *slog.Logger, dbbase string) error {
	_, db, err := crawldatabase.Open[crawler.Page](logger, dbbase, false)
	if err != nil {
		return err
	}
	defer db.Close()

	wordsIndex, err := index.LoadVocabAdvanced(filepath.Join(dbbase, "words.db"))
	if err != nil {
		return fmt.Errorf("Load words index (in 'words.db'): %w", err)
	}

	logger.Info("listen", "address", ":8000")
	return http.ListenAndServe(":8000", display.Handler(logger, &index.RealQuerier{
		DB:    db,
		Words: wordsIndex,
	}))
}

func mainDemoSearch(logger *slog.Logger, _ string) error {
	logger.Info("listen", "address", ":8000")
	return http.ListenAndServe(":8000", display.Handler(logger, index.FakeQuerier()))
}
