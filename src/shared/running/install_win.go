//go:build windows

package running

import (
	"haystack/utils"
	"os"
)

func InstallPath() string {
	return utils.NormalizePath(os.Getenv("LocalAppData")) + "\\Haystack"
}
