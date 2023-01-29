package index

import (
	"net/url"
	"time"
)

// The system that get the result for a query.
type Querier interface {
	QueryText(query string) (keywords []string, pages []PageResult)
}

type PageResult struct {
	Title, Description string
	LastModification   time.Time
	URL                url.URL
}
