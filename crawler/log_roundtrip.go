package crawler

import (
	"golang.org/x/exp/slog"
	"net/http"
)

// A simple http.RoundTripper that log all request.
type logRoundTripper struct {
	logger       *slog.Logger
	roundTripper http.RoundTripper
}

func newlogRoundTripper(roundTripper http.RoundTripper, logger *slog.Logger) http.RoundTripper {
	if roundTripper == nil {
		roundTripper = http.DefaultTransport
	}
	return &logRoundTripper{
		logger:       logger,
		roundTripper: roundTripper,
	}
}

func (r *logRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := r.roundTripper.RoundTrip(request)

	if err == nil {
		r.logger.Info("fetch.ok", "status", response.StatusCode, "url", request.URL.String())
	} else {
		r.logger.Info("fetch.err", "err", err.Error(), "url", request.URL.String())
	}

	return response, err
}
