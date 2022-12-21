package crawldatabase

import (
	"golang.org/x/exp/slog"
)

// All statistics of a database.
type Statistics struct {
	// Number element by type.
	Count [256]int

	// Total number of element.
	Total      int
	TotalFile  int
	TotalError int

	// The size of compressed chunck indexed by type of entry.
	FileSize [TypeError]int64

	// Sum of compressed chunck of data
	TotalFileSize int64
}

// Get the statistics from the metavalue map.
func getStatistics(m map[Key]metavalue) (stats Statistics) {
	stats.Total = len(m)

	for _, meta := range m {
		stats.Count[meta.Type]++
		if t := meta.Type; TypeFile <= t && t < TypeError {
			stats.FileSize[t] += int64(meta.Length)
		}
	}

	for _, n := range stats.Count[TypeFile:TypeError] {
		stats.TotalFile += n
	}
	for _, n := range stats.Count[TypeError:] {
		stats.TotalError += n
	}

	for _, size := range stats.FileSize {
		stats.TotalFileSize += size
	}

	return
}

// Log the total count and size.
func (stats Statistics) Log(logger *slog.Logger) {
	logger.LogAttrs(slog.InfoLevel, "db.stats.total",
		slog.Group("count",
			slog.Int("all", stats.Total),
			slog.Int("file", stats.TotalFile),
			slog.Int("error", stats.TotalError),
		),
		slog.Int64("size", stats.TotalFileSize),
	)
}

// Log the details of the statistics (total and each type) for count and size.
func (stats Statistics) LogAll(logger *slog.Logger) {
	stats.Log(logger)

	type2name := [...]string{
		TypeKnow:            "know",
		TypeRedirect:        "redirect",
		TypeFileRobots:      "fileRobots",
		TypeFileHTML:        "fileHTML",
		TypeFileRSS:         "fileRSS",
		TypeFileSitemap:     "fileSitemap",
		TypeFileFavicon:     "fileFavicon",
		TypeErrorNetwork:    "errorNetwork",
		TypeErrorParsing:    "errorParsing",
		TypeErrorFilterURL:  "errorFilterURL",
		TypeErrorFilterPage: "errorFilterPage",
	}

	for t, name := range type2name {
		if name == "" {
			continue
		}
		logger.Info("db.stats.count",
			"count", stats.Count[t],
			"percent", stats.Count[t]*100/stats.Total,
			"type", name)
	}

	logger.Info("db.stats.size", "total", 20)
	for t, name := range type2name[:TypeErrorNetwork] {
		if byte(t) < TypeFileRobots || name == "" {
			continue
		}
		logger.Info("db.stats.size",
			"size", stats.FileSize[t],
			"percent", stats.FileSize[t]*100/stats.TotalFileSize,
			"type", name)
	}
}
