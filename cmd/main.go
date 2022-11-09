package main

import (
	"context"
	"github.com/HuguesGuilleus/isty-search/crawler"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	log.Println("main()")

	config := crawler.Config{
		DBRoot: "db",
		Input: []string{
			"https://www.gouvernement.fr/",
			"https://www.vie-publique.fr/",
		},
		FilterURL: []func(*url.URL) string{
			func(u *url.URL) string {
				if u.Scheme != "https" {
					return "url-not_https"
				}
				h := u.Host
				if strings.HasSuffix(h, "legifrance.gouv.fr") {
					return "url-legifrance"
				}
				valid := h == "www.vie-publique.fr" ||
					strings.HasSuffix(h, "gouvernement.fr") ||
					strings.HasSuffix(h, "gouv.fr")
				if !valid {
					return "url-not_gouv"
				}
				return ""
			},
		},
		FilterPage: []func(*htmlnode.Root) string{
			func(root *htmlnode.Root) string {
				switch root.Meta.Langage {
				case "fr", "fr_FR", "fr-FR":
					return ""
				default:
					log.Println("unknwo_langage:", root.Meta.Langage)
					return "unknwo_langage"
				}
			},
		},

		MaxLength: 15_000_000,
		MaxGo:     20,

		MinCrawlDelay: time.Millisecond * 500,
		MaxCrawlDelay: time.Second * 10,

		LogOutput: os.Stdout,
	}

	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer ctxCancel()

	err := crawler.Crawl(ctx, config)
	if err != nil {
		log.Fatal(err)
	}
}
