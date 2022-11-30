package sloghandlers

import (
	"testing"
)

func TestHandlerNull(t *testing.T) {
	fillLogger(NewNullHandler())
}
