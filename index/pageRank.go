package index

import (
	"github.com/HuguesGuilleus/isty-search/crawler"
	"github.com/HuguesGuilleus/isty-search/keys"
	"golang.org/x/exp/slog"
	"sort"
)

type PageRank struct {
	links map[keys.Key][]keys.Key
}

func NewPageRank() PageRank {
	return PageRank{
		links: make(map[keys.Key][]keys.Key),
	}
}

func (pr *PageRank) Process(page *crawler.Page) {
	urls := page.GetURLs()
	links := make([]keys.Key, len(urls))
	i := 0
	for key := range urls {
		links[i] = key
		i++
	}
	pr.links[keys.NewURL(&page.URL)] = links
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

func (pr *PageRank) Score(repeat int, epsilon float32) (int, []Score) {
	return score(pr.links, repeat, epsilon)
}

func score(allLinks map[keys.Key][]keys.Key, repeat int, epsilon float32) (int, []Score) {
	pageRankFilter(allLinks)

	key2index := make(map[keys.Key]int, len(allLinks))
	index2key := make([]keys.Key, len(allLinks))
	i := 0
	totalLinks := 0
	for key, links := range allLinks {
		index2key[i] = key
		key2index[key] = i
		i++
		totalLinks += len(links)
	}

	pagesLinks := make([]int, totalLinks)
	i = 0
	pages := make([][]int, len(allLinks))
	for key, links := range allLinks {
		begin := i
		for _, link := range links {
			pagesLinks[i] = key2index[link]
			i++
		}
		pages[key2index[key]] = pagesLinks[begin:i]
	}

	repeatition, rank := pageRankMultiplication(pages, repeat, epsilon)
	scores := make([]Score, len(rank))
	for i, rank := range rank {
		scores[i] = Score{
			Key:  index2key[i],
			Rank: rank,
		}
	}

	SortScores(scores)
	return repeatition, scores
}

// Filter unknown key, double key, and key pointed to the page key.
func pageRankFilter(allLinks map[keys.Key][]keys.Key) {
	for key, links := range allLinks {
		sort.Slice(links, func(i, j int) bool {
			return links[i].Less(&links[j])
		})
		writeIndex := 0
		previous := keys.Key{}
		for _, linkKey := range links {
			if key == linkKey || previous == linkKey {
				continue
			} else if _, exist := allLinks[linkKey]; !exist {
				continue
			}
			previous = linkKey
			links[writeIndex] = linkKey
			writeIndex++
		}
		allLinks[key] = links[:writeIndex]
	}
}

func pageRankMultiplication(pages [][]int, maxRepeat int, epsilon float32) (int, []float32) {
	oldRank := make([]float32, len(pages))
	newRank := make([]float32, len(pages))
	for i := range newRank {
		newRank[i] = 1.0
	}

	r := 0
	for ; r < maxRepeat && epsilon < norm2(oldRank, newRank); r++ {
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

	return r, newRank
}

// Return ||v1-v2||^2
func norm2(v1, v2 []float32) (sum float32) {
	for i, f1 := range v1 {
		v := f1 - v2[i]
		sum += v * v
	}
	return
}
