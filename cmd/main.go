package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/HuguesGuilleus/isty-search/crawler"
	"github.com/HuguesGuilleus/isty-search/crawler/db"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"github.com/HuguesGuilleus/isty-search/search"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"
)

var dbPath string

func getDB() (*crawler.DB, []*url.URL) {
	db, urls, err := crawler.OpenDB(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	return db, urls
}

func main() {
	actions := map[string]func() error{
		"crawl": mainCrawl,
		"vocab": mainVocab,
		"stats": mainDBStatistics,
	}
	if len(os.Args) < 2 || actions[os.Args[1]] == nil {
		os.Stderr.WriteString("Unknown action. Possible actions are:\n")
		for name := range actions {
			fmt.Fprintf(os.Stderr, "\t%s\n", name)
		}
		os.Exit(2)
	}

	flag.StringVar(&dbPath, "db", "db", "dataBase directory path (can not exist)")

	defer func(begin time.Time) { fmt.Println("\nduration: ", time.Since(begin)) }(time.Now())

	os.Args = os.Args[1:]
	if err := actions[os.Args[0]](); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func mainCrawl() error {
	flag.Parse()

	db, urls := getDB()
	defer db.Close()

	config := crawler.Config{
		DB:    db,
		Input: urls,

		FilterURL: []func(*url.URL) string{
			func(u *url.URL) string {
				if u.Scheme != "https" {
					return "url-not_https"
				}
				switch u.Host {
				case "uvsq.fr":
				case "cas.uvsq.fr", "cas2.uvsq.fr":
					return "url-cas_uvsq"
				default:
					if !strings.HasSuffix(u.Host, ".uvsq.fr") {
						return "url-not_uvsq"
					}
				}
				return ""
			},
		},
		FilterPage: []func(*htmlnode.Root) string{
			func(root *htmlnode.Root) string {
				switch root.Meta.Langage {
				case "":
					fallthrough
				case "fr", "fr_FR", "fr-FR", "fr-incl", "fr_incl":
					return ""
				case "en":
					// No language log
					return "unknwo_language"
				default:
					log.Println("unknwo_language:", root.Meta.Langage)
					return "unknwo_language"
				}
			},
		},

		MaxLength: 15_000_000,
		MaxGo:     30,

		MinCrawlDelay: time.Millisecond * 500,
		MaxCrawlDelay: time.Second * 10,

		LogOutput: os.Stdout,
	}

	ctx, ctxCancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer ctxCancel()

	return crawler.Crawl(ctx, config)
}

func mainVocab() error {
	db, _ := getDB()
	defer db.Close()

	pageCounter := search.PageCounter(0)
	vocabCounter := make(search.VocabCounter)
	if err := crawler.Process(db, []crawler.Processor{vocabCounter, &pageCounter}); err != nil {
		return err
	}

	log.Println("number of page", pageCounter)
	log.Println("words number:", len(vocabCounter))
	log.Println("total words:", vocabCounter.TotalWords())

	return nil
}

func mainDBStatistics() error {
	dataBase, _ := getDB()
	defer dataBase.Close()

	nb, size, err := db.Statistics(dataBase.URLsDB, dataBase.KeyValueDB)

	log.Printf("  nb: %10d\n", nb)
	log.Printf("size: %10d\n", size)

	return err
}
