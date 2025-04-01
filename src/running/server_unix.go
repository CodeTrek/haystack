//go:build !windows

package running

func LockAndRunAsServer() (func(), error) {
	return func() {}, nil
}
