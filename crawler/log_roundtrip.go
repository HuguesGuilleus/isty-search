package crawler

import (
	"github.com/HuguesGuilleus/isty-search/bytesrecycler"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// A simple http.RoundTripper that log all request.
type logRoundTripper struct {
	logOutput    io.Writer
	logMutex     sync.Mutex
	roundTripper http.RoundTripper
}

func newlogRoundTripper(roundTripper http.RoundTripper, logOutput io.Writer) http.RoundTripper {
	if roundTripper == nil {
		roundTripper = http.DefaultTransport
	}
	return &logRoundTripper{
		logOutput:    logOutput,
		roundTripper: roundTripper,
	}
}

func (r *logRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := r.roundTripper.RoundTrip(request)

	if r.logOutput != nil {
		buff := recycler.Get()
		defer recycler.Recycle(buff)
		buff.WriteString(time.Now().Format("2006-01-02 15:04:05 "))
		if err == nil {
			buff.WriteString("(")
			buff.WriteString(strconv.Itoa(response.StatusCode))
			buff.WriteString(") ")
			buff.WriteString(request.URL.String())
		} else {
			buff.WriteString("(err) ")
			buff.WriteString(request.URL.String())
			buff.WriteString(" // ")
			buff.WriteString(err.Error())
		}
		buff.WriteByte('\n')

		r.logMutex.Lock()
		defer r.logMutex.Unlock()
		buff.WriteTo(r.logOutput)
	}

	return response, err
}
