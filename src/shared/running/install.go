//go:build !windows

package running

import "path/filepath"

func InstallPath() string {
	return filepath.Join(UserHomeDir(), ".local", "bin")
}
