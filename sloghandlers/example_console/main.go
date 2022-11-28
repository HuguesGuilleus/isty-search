package main

import (
	"github.com/HuguesGuilleus/isty-search/sloghandlers"
	"golang.org/x/exp/slog"
	"time"
)

func main() {
	logger := slog.New(sloghandlers.NewConsole(slog.DebugLevel))

	for i, l := 0, 50; i < l; i++ {
		logger.Info("%", "%i", i, "%len", l)
		logger.Warn("hello", "i", i)
		time.Sleep(time.Millisecond * 50)
	}
	logger.Info("%end")

	logger.Debug("coucou", "hello", "world!!!")
}
