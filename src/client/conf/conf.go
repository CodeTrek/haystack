package conf

import (
	"search-indexer/running"
)

func checkMode() {
	if running.IsServerMode() {
		panic("client conf is not accessible in server mode!")
	}
}

func Load() error {
	checkMode()

	return nil
}
