package search

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"github.com/HuguesGuilleus/isty-search/common"
	"github.com/HuguesGuilleus/isty-search/crawler"
	crawldatabase "github.com/HuguesGuilleus/isty-search/crawler/database"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"github.com/HuguesGuilleus/isty-search/index"
	"github.com/HuguesGuilleus/isty-search/keys"
)

// The number of element in a chunck (aka the result page)
const ChunckLen = 10

var ErrPageTooLong = errors.New("Page argument too long")

type DB struct {
	// Get a page from the crawler database.
	CrawlerDB *crawldatabase.Database[crawler.Page]
	// Get for a word key, all pages with the word and occurence coeficient.
	// The sub slice are sorted by key.
	ReverseIndex map[keys.Key][]index.KeyFloat32
	// Get a global score, like a page rank.
	GlobalScore map[keys.Key]float32
}

type Result struct {
	// Parsed query, the keywords
	Queries []Query
	// The (at most ChunckLen) page details (title, URL...)
	Results []PageResult
	// The number of founded page
	NumberOfResults int
	// The number of chunck (result page)
	NumberOfChunck int
}

type PageResult struct {
	// The page key and his URL.
	Key keys.Key
	URL url.URL
	// Metadata of the page.
	Title       string
	Description string
}

func FakeDB() *DB {
	_, crawlerDB, _ := crawldatabase.OpenMemory[crawler.Page](nil, "", false)

	wordIndex := make([]index.KeyFloat32, 0)
	globalScore := make(map[keys.Key]float32)
	for i := 0; i < 51; i++ {
		k := keys.Key{byte(i)}
		istr := strconv.Itoa(i)

		wordIndex = append(wordIndex, index.KeyFloat32{Key: k})
		globalScore[k] = float32(100 - i)

		crawlerDB.SetValue(k, &crawler.Page{
			URL:  *common.ParseURL("https://example.com/page-" + istr),
			Html: &htmlnode.Root{Meta: htmlnode.Meta{Title: "page:" + istr, Description: "desc:" + istr}},
		}, crawldatabase.TypeFileHTML)
	}

	return &DB{
		CrawlerDB: crawlerDB,
		ReverseIndex: map[keys.Key][]index.KeyFloat32{
			keys.NewString("word"): wordIndex,
		},
		GlobalScore: globalScore,
	}
}

func (db *DB) Search(queryString string, chunck int) (*Result, error) {
	queries, pages := search(queryString, db.ReverseIndex)
	score(pages, db.GlobalScore)

	numberOfChunck := len(pages) / ChunckLen
	if len(pages)%ChunckLen != 0 {
		numberOfChunck++
	}

	if chunck*ChunckLen > len(pages) {
		return nil, ErrPageTooLong
	}

	results := make([]PageResult, 0, ChunckLen)
	for i := chunck * ChunckLen; i < (chunck+1)*ChunckLen && i < len(pages); i++ {
		p, err := db.pageResult(pages[i].Key)
		if err != nil {
			return nil, err
		}
		results = append(results, *p)
	}

	return &Result{
		Queries:         queries,
		Results:         results,
		NumberOfResults: len(pages),
		NumberOfChunck:  numberOfChunck,
	}, nil
}

func (db *DB) pageResult(key keys.Key) (*PageResult, error) {
	page, _, err := db.CrawlerDB.GetValue(key)
	if err != nil {
		return nil, err
	} else if page.Html == nil {
		return nil, fmt.Errorf("Not HTML Page for: %s", key)
	}

	return &PageResult{
		Key:         key,
		URL:         page.URL,
		Title:       page.Html.Meta.Title,
		Description: page.Html.Meta.Description,
	}, nil
}
