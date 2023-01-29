package display

import (
	"github.com/HuguesGuilleus/isty-search/common"
	"github.com/HuguesGuilleus/isty-search/index"
	"net/http"
	"strconv"
)

func sendResult(w http.ResponseWriter, r *http.Request, query string, p int, querier index.Querier) {
	_, result := querier.QueryText(query)
	resultLen := len(result)
	if len(result) > 10 {
		result = result[:10]
	}

	nodeResults := make([]node, len(result))
	for i, p := range result {
		u := p.URL.String()
		nodeResults[i] = np("li.search-results-item",
			nap("a.search-results-item", []string{`href="` + u + `"`},
				nt("div.search-results-item-title", limitString(p.Title, 50)),
				np("div.search-results-item-info",
					ntime(".search-results-item-info-time", p.LastModification),
					nt("span.search-results-item-info-url", limitString(u, 70)),
				),
				nt("div.search-results-item-desc", limitString(p.Description, 150)),
			),
		)
	}

	buff := common.GetBuffer()
	defer common.RecycleBuffer(buff)
	page2html(buff, page{
		Title: "Résultat",
		Body: np("body.search",
			np("form.search-top",
				na(".search-top-home.pixelated", "/", nap("img.search-top-home-img", []string{
					"src=/image/tree.png",
					"width=96", "height=96",
					`title="Home"`,
				})),

				nap(`input.search-top-searchbar`, []string{
					"type=search",
					"name=q",
					"value=" + strconv.Quote(query),
					`placeholder="Mots clés de recherche"`}),

				np("div.search-top-kind",
					np("button.search-top-kind-buttonText",
						nap("img.search-top-kind-buttonText-img.pixelated", []string{
							"src=/image/search-text.png",
							"width=13", "height=13",
						}),
						nt("span", "Text"),
					),
				),
			),
			np("div.search-query",
				nt("div.search-query-resultLen", strconv.Itoa(resultLen)),
				nt("div.search-query-page", strconv.Itoa(p)),
			),
			np("ul.search-results", nodeResults...),
			nt("footer.search-footer", "Hugues GUILLEUS, Projet ISTY-Search, 2022-2023"),
		),
	})

	w.Header().Add("Content-Length", strconv.Itoa(buff.Len()))
	w.Header().Add("Content-Type", "text/html")
	w.Write(buff.Bytes())
}

func limitString(s string, limit int) string {
	if len(s) <= limit {
		return s
	}

	l := 0
	for i := range s {
		l++
		if l > limit {
			return s[:i] + "..."
		}
	}

	return s
}
