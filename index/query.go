package index

import (
	"github.com/HuguesGuilleus/isty-search/crawler"
	"github.com/HuguesGuilleus/isty-search/crawler/database"
	"github.com/HuguesGuilleus/isty-search/keys"
	"net/url"
	"time"
)

// The system that get the result for a query.
type Querier interface {
	QueryText(query string) ([]string, []*PageResult)
	// Query(query string) []keys.Key
	// PageResult(key keys.Key) *PageResult
}

type PageResult struct {
	Key                keys.Key
	Title, Description string
	LastModification   time.Time
	URL                url.URL
}

type RealQuerier struct {
	DB    *crawldatabase.Database[crawler.Page]
	Words ReverseIndex
}

func (rq *RealQuerier) QueryText(query string) ([]string, []*PageResult) {
	founded := rq.Query(query)
	results := make([]*PageResult, len(founded))
	for i, key := range founded {
		results[i] = rq.PageResult(key)
	}
	return nil, results
}

func (rq *RealQuerier) Query(query string) []keys.Key {
	pages := rq.Words[keys.NewString(query)]
	list := make([]keys.Key, len(pages))
	for i, item := range pages {
		list[i] = item.Key
	}
	return list
}

func (rq *RealQuerier) PageResult(key keys.Key) *PageResult {
	page, lastSave, _ := rq.DB.GetValue(key)
	if page == nil || page.Html == nil {
		return nil
	}

	return &PageResult{
		Title:            page.Html.Meta.Title,
		Description:      page.Html.Meta.Description,
		URL:              page.URL,
		LastModification: lastSave,
	}
}
