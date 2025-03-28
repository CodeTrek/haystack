package conf

import (
	"search-indexer/runtime"
)

func checkMode() {
	if !runtime.IsServerMode() {
		panic("server conf is not accessible in client mode!")
	}
}

func Load() error {
	checkMode()

	return nil
}
