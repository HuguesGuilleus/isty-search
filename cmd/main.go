package main

import (
	"context"
	"github.com/HuguesGuilleus/isty-search/crawler"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"
)

func main() {
	db, urls, err := crawler.OpenDB("db")
	if err != nil {
		log.Fatal(err)
	}
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
					return ""
				case "fr", "fr_FR", "fr-FR", "fr-incl", "fr_incl":
					return ""
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

	if err = crawler.Crawl(ctx, config); err != nil {
		log.Fatal(err)
	}
}
