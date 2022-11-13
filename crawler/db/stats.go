package db

import (
	"io/fs"
	"os"
)

type Stats struct {
	Done, Ban, Todo int64
}

func Statistics[T any](urlsDB *URLsDB, kvDB KeyValueDB[T]) (nb, size Stats, err error) {
	urlsDB.mutex.Lock()
	defer urlsDB.mutex.Unlock()

	for key, value := range urlsDB.keysMap {
		if value == existValue {
			nb.Todo++
			continue
		}

		stat := os.FileInfo(nil)
		stat, err = os.Stat(key.path(kvDB.base))
		if err != nil {
			return
		}

		if value > 0 {
			nb.Done++
			size.Done += stat.Size()
		} else {
			nb.Ban++
			size.Ban += stat.Size()
		}
	}

	nbDir := int64(0)
	withDir := int64(0)
	noDir := int64(0)
	blocks := int64(0)
	err = fs.WalkDir(os.DirFS(kvDB.base), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		stat, err := d.Info()
		if err != nil {
			return err
		}

		withDir += stat.Size()
		if !stat.IsDir() {
			noDir += stat.Size()
			blocks += (stat.Size()/4096 + 1) * 4096
		} else {
			nbDir++
		}

		return nil
	})

	println("withDir:", withDir)
	println("  noDir:", noDir)
	println(" blocks:", blocks)
	println()
	println("  nbDir:", nbDir)
	println("    sum:", blocks+nbDir*4096)
	println()

	return
}
