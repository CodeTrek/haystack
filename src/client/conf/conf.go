package conf

import (
	"search-indexer/runtime"
)

func checkMode() {
	if runtime.IsServerMode() {
		panic("client conf is not accessible in server mode!")
	}
}

func Load() error {
	checkMode()

	return nil
}
