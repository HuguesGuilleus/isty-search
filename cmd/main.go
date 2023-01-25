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

func main() {
	db := flag.String("db", "db1", "dataBase directory path (can not exist)")
	flag.Parse()

	os.MkdirAll(filepath.Join(*db, "log"), 0o775)
	jsonLogFile, err := os.Create(filepath.Join(*db, "log", time.Now().UTC().Format("2006-01-02_15-04-05Z.log")))
	fatal(err)

	logger := slog.New(sloghandlers.NewMultiHandlers(
		sloghandlers.NewConsole(slog.InfoLevel),
		slog.NewJSONHandler(jsonLogFile),
	))

	actions := map[string]func(logger *slog.Logger, dbbase string) error{
		"crawl":    mainCrawl,
		"vocab":    mainVocab,
		"stats":    mainDBStatistics,
		"pagerank": mainPageRank,
		"index":    mainIndex,
		"search":   mainSearch,
		"demoserv": mainDemoServ,
	}
	action := actions[flag.Arg(0)]
	if action == nil {
		fmt.Println("Unknown action. Possible actions are:")
		for name := range actions {
			fmt.Printf("\t%s\n", name)
		}
		os.Exit(1)
	}

	defer func(begin time.Time) { logger.Info("duration", "d", time.Since(begin)) }(time.Now())
	err = action(logger, *db)
	fatal(err)
}

func fatal(err error) {
	if err != nil {
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

func mainVocab(logger *slog.Logger, dbbase string) error {
	_, db, err := crawldatabase.Open[crawler.Page](logger, dbbase, false)
	if err != nil {
		return err
	}
	defer db.Close()

	pageCounter := index.PageCounter(0)
	// vocabCounter := make(index.VocabCounter)
	allvocab := make(index.VocabAdvanced)
	if err := crawler.Process(db, logger, &pageCounter, allvocab); err != nil {
		return err
	}

	logger.Info("vocab.stats", "page", pageCounter, "words", len(allvocab))
	wordsLen := 0
	for word := range allvocab {
		wordsLen += len(word)
	}
	logger.Info("vocab.allword", "len", "wordsLen")
	// logger.Info("vocab.stats", "page", pageCounter, "words", len(vocabCounter), "occuenrence", vocabCounter.TotalWords())

	return nil
}

func mainPageRank(logger *slog.Logger, dbbase string) error {
	_, db, err := crawldatabase.Open[crawler.Page](logger, dbbase, false)
	if err != nil {
		return err
	}
	defer db.Close()

	pageRank := index.NewPageRank()
	if err := crawler.Process(db, logger, &pageRank); err != nil {
		return err
	}
	pageRank.DevScore()
	pageRank.DevStats(logger)

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
	if err := crawler.Process(db, logger, &pageRank, &wordsIndex); err != nil {
		return err
	}

	logger.Info("order.pagerank")
	scores := pageRank.Score(200)
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

	result := wordsIndex[keys.NewString("isty")]
	logger.Info("result.len", "len", len(result))
	if len(result) > 20 {
		result = result[:20]
	}
	for i, pageResult := range result {
		page, _, err := db.GetValue(pageResult.Page)
		if err != nil {
			return fmt.Errorf("Get value([i:%d] %s): %w", i, pageResult.Page, err)
		}
		logger.Info("result",
			"i", i,
			"url", page.URL.String(),
			"title", page.Html.Meta.Title,
		)
	}

	return nil
}

func mainDemoServ(logger *slog.Logger, _ string) error {
	return http.ListenAndServe(":8000", display.DemoServ(logger))
}
