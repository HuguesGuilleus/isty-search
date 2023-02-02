package index

import (
	"github.com/HuguesGuilleus/isty-search/crawler"
	"github.com/HuguesGuilleus/isty-search/keys"
	"golang.org/x/exp/slog"
	"net/url"
	"sort"

	"fmt"
	"io"
)

type Links struct {
	// All global link (page1 from domain-a.net --> page2 from domain-b.net)
	globalLinks map[keys.Key][]keys.Key
	// The redirection map
	redirection map[keys.Key]keys.Key
}

func NewLinks(redirection map[keys.Key]keys.Key) Links {
	return Links{
		redirection: redirection,
		globalLinks: make(map[keys.Key][]keys.Key),
	}
}

func (links *Links) Process(page *crawler.Page) {
	links.addURLs(&page.URL, page.GetURLs())
}
func (links *Links) addURLs(base *url.URL, m map[keys.Key]*url.URL) {
	s := make([]keys.Key, len(m))
	i := 0
	for key := range m {
		if target, ok := links.redirection[key]; ok {
			key = target
		}
		s[i] = key
		i++
	}
	sort.Slice(s, func(i, j int) bool { return s[i].Less(&s[j]) })
	links.globalLinks[keys.NewURL(base)] = s
}

func (pr *Links) DevStats(logger *slog.Logger) {
	max := 0
	for _, links := range pr.globalLinks {
		if len(links) > max {
			max = len(links)
		}
	}

	distribution := make([]int, max+1, max+1)
	for _, link := range pr.globalLinks {
		distribution[len(link)]++
	}

	logger.Info("pr.stats", "page", len(pr.globalLinks), "max", max)
	for i, count := range distribution {
		logger.Info("pr.distribution", "i", i, "count", count)
	}
}

func (links *Links) PageRank(repeat int, epsilon float32) (int, []Score) {
	return pageRank(links.globalLinks, repeat, epsilon)
}

func (links *Links) GetLinksList(w io.Writer) {
	allLinks := links.globalLinks

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

	fmt.Fprintf(w, "%d %d\n", len(allLinks), totalLinks)
	for _, targets := range pages {
		fmt.Fprintf(w, "%d ", len(targets))
		for _, target := range targets {
			fmt.Fprintf(w, "%d ", target)
		}
		fmt.Fprintln(w)
	}
}

func pageRank(allLinks map[keys.Key][]keys.Key, repeat int, epsilon float32) (int, []Score) {
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

// Filter unknown key and key pointed to the page key.
func pageRankFilter(allLinks map[keys.Key][]keys.Key) {
	for key, links := range allLinks {
		writeIndex := 0
		for _, linkKey := range links {
			if key == linkKey {
				continue
			} else if _, exist := allLinks[linkKey]; !exist {
				continue
			}
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
