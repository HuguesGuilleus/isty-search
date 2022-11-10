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
	log.Println("main()")

	config := crawler.Config{
		DBRoot: "db",
		Input: []string{
			// "https://www.gouvernement.fr/",
			// "https://www.vie-publique.fr/",
			"https://www.uvsq.fr",
		},
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
				case "fr", "fr_FR", "fr-FR":
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

	err := crawler.Crawl(ctx, config)
	if err != nil {
		log.Fatal(err)
	}
}
