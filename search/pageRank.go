package search

import (
	"bytes"
	"fmt"
	"github.com/HuguesGuilleus/isty-search/crawler"
	"github.com/HuguesGuilleus/isty-search/crawler/db"
	"sort"
)

type PageRank struct {
	links map[db.Key][]db.Key
	urls  map[db.Key]string // only for dev
}

func NewPageRank() PageRank {
	return PageRank{
		links: make(map[db.Key][]db.Key),
		urls:  make(map[db.Key]string),
	}
}

func (pr *PageRank) Process(page *crawler.Page) {
	urls := page.Html.GetURL(&page.URL)

	links := make([]db.Key, len(urls))
	for i, u := range urls {
		links[i] = db.NewURLKey(u)
	}
	pr.links[db.NewURLKey(&page.URL)] = links
	pr.urls[db.NewURLKey(&page.URL)] = page.URL.String()
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

	key2index := make(map[db.Key]int, len(pr.links))
	index2key := make([]db.Key, len(pr.links))
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
		previous := db.Key{}
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