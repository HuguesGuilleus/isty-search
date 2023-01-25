package display

import (
	"golang.org/x/exp/slog"
	"net/http"
	"strconv"
)

func DemoServ(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("http", "url", r.URL.String())

		w.Header().Add("Content-Length", strconv.Itoa(len(home)))
		w.Header().Add("Content-Type", "text/html")
		w.Write(home)
	})
}
