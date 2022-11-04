package crawler

import (
	"bytes"
	"io"
	"net/http"
)

// A simple http.RoundTripper, map is indexex by URL and give the body bytes.
type mapRoundTripper map[string][]byte

func (m mapRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	makeResponse := func(status int, body []byte) (*http.Response, error) {
		return &http.Response{
			Status:     http.StatusText(status),
			StatusCode: status,

			Proto:      request.Proto,
			ProtoMajor: request.ProtoMajor,
			ProtoMinor: request.ProtoMinor,

			Body:          io.NopCloser(bytes.NewReader(body)),
			ContentLength: int64(len(body)),
			Request:       request,
		}, nil
	}

	if request.Method != http.MethodGet {
		return makeResponse(http.StatusMethodNotAllowed, []byte("Method not allowed"))
	}

	if body := m[request.URL.String()]; len(body) != 0 {
		return makeResponse(http.StatusOK, body)
	}

	return makeResponse(http.StatusNotFound, []byte("Not found"))
}
