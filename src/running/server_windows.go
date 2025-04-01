//go:build windows

package running

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	kernel32     = syscall.NewLazyDLL("kernel32.dll")
	createMutexW = kernel32.NewProc("CreateMutexW")
	releaseMutex = kernel32.NewProc("ReleaseMutex")
	closeHandle  = kernel32.NewProc("CloseHandle")
	lastError    = kernel32.NewProc("GetLastError")
)

func LockAndRunAsServer() (func(), error) {
	fmt.Println("Try LockAndRunAsServer")
	// Create a mutex with a unique name for current session
	mutexName := "Global\\SearchIndexerSingleInstance@20250401"
	namePtr, err := syscall.UTF16PtrFromString(mutexName)
	if err != nil {
		return nil, err
	}

	// Create mutex
	handle, _, err := createMutexW.Call(
		0, // Security attributes
		1, // Initial owner
		uintptr(unsafe.Pointer(namePtr)),
	)
	if handle == 0 {
		return nil, err
	}

	// Check if mutex already exists
	errCode, _, _ := lastError.Call()
	if syscall.Errno(errCode) == syscall.ERROR_ALREADY_EXISTS {
		closeHandle.Call(handle)
		return nil, syscall.Errno(syscall.ERROR_ALREADY_EXISTS)
	}

	fmt.Println("Mutex created", handle)

	// Return cleanup function
	cleanup := func() {
		releaseMutex.Call(handle)
		closeHandle.Call(handle)
	}

	return cleanup, nil
}
