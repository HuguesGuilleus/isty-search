package crawldatabase

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
