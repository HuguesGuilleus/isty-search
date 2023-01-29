package display

import (
	"github.com/HuguesGuilleus/isty-search/index"
	"golang.org/x/exp/slog"
	"net/http"
	"strconv"
)

func Handler(logger *slog.Logger, querier index.Querier) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/favicon.ico":
			serveStatic(w, "image/x-icon", imageFavicon)
		case "/image/search-text.png":
			serveStatic(w, "image/png", imageSearchText)
		case "/image/tree.png":
			serveStatic(w, "image/png", imageTree)
		case "/":
			logger.Info("serv.static.home")
			serveStatic(w, "text/html", home)

		case "/result":
			query := r.URL.Query()
			q := query.Get("q")
			if q == "" {
				logger.Info("serv.noquey", "url", r.URL.String())
				http.Redirect(w, r, "/", http.StatusPermanentRedirect)
				return
			}

			pageString := query.Get("page")
			page := 0
			if pageString != "" {
				parsedPage, err := strconv.Atoi(pageString)
				if err != nil {
					logger.Info("serv.wrongpage.syntax", "url", r.URL.String())
					http.Error(w, "can not parsing page number: "+err.Error(), http.StatusBadRequest)
					return
				} else if parsedPage < 0 {
					logger.Info("serv.wrongpage.negative", "url", r.URL.String())
					http.Error(w, "page number can be negative", http.StatusBadRequest)
					return
				}
				page = parsedPage
			}

			logger.Info("serv.search", "page", page, "query", q)
			sendResult(w, r, q, page, querier)

		default:
			logger.Info("serv.404", "url", r.URL.String())
			http.NotFound(w, r)
		}
	})
}

func serveStatic(w http.ResponseWriter, mime string, content []byte) {
	w.Header().Add("Content-Length", strconv.Itoa(len(content)))
	w.Header().Add("Content-Type", mime)
	w.Write(content)
}
