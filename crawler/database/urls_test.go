package crawldatabase

import (
	"github.com/HuguesGuilleus/isty-search/common"
	"github.com/HuguesGuilleus/isty-search/sloghandlers"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
	"testing"
)

func TestLoadURLs(t *testing.T) {
	records, handler := sloghandlers.NewHandlerRecords(slog.DebugLevel)

	received := loadURLs(
		slog.New(handler),
		[]byte(
			"https://www.google.com/\n"+
				"https://www.wikipedia.org/\n"+
				"https://www.wikipedia.fr/\n"),
		map[Key]metavalue{
			NewKeyString("https://www.google.com/"):    metavalue{Type: TypeKnow},
			NewKeyString("https://www.wikipedia.org/"): metavalue{Type: TypeErrorNetwork},
			NewKeyString("https://www.wikipedia.fr/"):  metavalue{Type: TypeRedirect},
		},
		[]byte{TypeKnow, TypeErrorNetwork})

	assert.Equal(t, common.ParseURLs(
		"https://www.google.com/",
		"https://www.wikipedia.org/",
	), received)
	assert.Nil(t, *records)
}
