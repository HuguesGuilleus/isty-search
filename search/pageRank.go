package search

import (
	"bytes"
	"fmt"
	"github.com/HuguesGuilleus/isty-search/crawler"
	"github.com/HuguesGuilleus/isty-search/crawler/database"
	"golang.org/x/exp/slog"
	"sort"
)

type PageRank struct {
	links map[crawldatabase.Key][]crawldatabase.Key
	urls  map[crawldatabase.Key]string // only for dev
}

func NewPageRank() PageRank {
	return PageRank{
		links: make(map[crawldatabase.Key][]crawldatabase.Key),
		urls:  make(map[crawldatabase.Key]string),
	}
}

func (pr *PageRank) Process(page *crawler.Page) {
	urls := page.GetURLs()
	links := make([]crawldatabase.Key, len(urls))
	i := 0
	for key := range urls {
		links[i] = key
		i++
	}
	pr.links[crawldatabase.NewKeyURL(&page.URL)] = links
	pr.urls[crawldatabase.NewKeyURL(&page.URL)] = page.URL.String()
}

func (pr *PageRank) DevStats(logger *slog.Logger) {
	max := 0
	for _, links := range pr.links {
		if len(links) > max {
			max = len(links)
		}
	}

	distribution := make([]int, max+1, max+1)
	for _, link := range pr.links {
		distribution[len(link)]++
	}

	logger.Info("pr.stats", "page", len(pr.links), "max", max)
	for i, count := range distribution {
		logger.Info("pr.distribution", "i", i, "count", count)
	}
}

// Run pr.Score(), sort the result and print beter line.
func (pr *PageRank) DevScore() {
	scores := pr.Score(200)

	if len(scores) > 30 {
		scores = scores[:30]
	}

	for _, score := range scores {
		fmt.Printf("%2.3f %s\n", score.Rank, pr.urls[score.Key])
	}
}

func (pr *PageRank) Score(repeat int) []Score {
	pr.filterKey()

	key2index := make(map[crawldatabase.Key]int, len(pr.links))
	index2key := make([]crawldatabase.Key, len(pr.links))
	i := 0
	for key := range pr.links {
		index2key[i] = key
		key2index[key] = i
		i++
	}
	pages := make([][]int, len(pr.links))
	for i := range pages {
		linksKey := pr.links[index2key[i]]
		linksIndex := make([]int, len(linksKey))
		for i, link := range linksKey {
			linksIndex[i] = key2index[link]
		}
		pages[i] = linksIndex
	}

	rank := pageRankMultiplication(pages, repeat)
	scores := make([]Score, len(rank))
	for i, rank := range rank {
		scores[i] = Score{
			Key:  index2key[i],
			Rank: rank,
		}
	}

	SortScores(scores)
	return scores
}

// Filter unknown key, double key, and key pointed to the page key.
func (pr *PageRank) filterKey() {
	for key, links := range pr.links {
		i := 0
		sort.Slice(links, func(i, j int) bool { return bytes.Compare(links[i][:], links[j][:]) < 0 })
		previous := crawldatabase.Key{}
		for _, linkKey := range links {
			if _, exist := pr.links[linkKey]; !exist || key == linkKey || bytes.Equal(previous[:], linkKey[:]) {
				continue
			}
			previous = linkKey
			links[i] = linkKey
			i++
		}
		pr.links[key] = links[:i]
	}
}

func pageRankMultiplication(pages [][]int, repeat int) []float32 {
	oldRank := make([]float32, len(pages))
	newRank := make([]float32, len(pages))
	for i := range newRank {
		newRank[i] = 1.0
	}

	for i := 0; i < repeat; i++ {
		oldRank, newRank = newRank, oldRank
		for i := range newRank {
			newRank[i] = 0
		}

		for pageIndex, pageLinks := range pages {
			v := oldRank[pageIndex] / float32(len(pageLinks))
			for _, linkIndex := range pageLinks {
				newRank[linkIndex] += v
			}
		}
	}

	return newRank
}
