package running

import (
	"flag"
	"haystack/utils"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	userHomeDir string
	daemonMode  = flag.Bool("daemon", false, "Run in daemon mode")
	version     string
)

func SetVersion(ver string) {
	if len(version) > 0 {
		return
	}
	version = ver
}

func Version() string {
	return version
}

func IsDevVersion() bool {
	return version == "dev"
}

func Init() error {
	var err error
	userHomeDir, err = os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get user's home directory: %v", err)
		return err
	}

	return nil
}

func UserHomeDir() string {
	return userHomeDir
}

func IsDaemonMode() bool {
	return *daemonMode
}

func ExecutableName() string {
	return filepath.Base(Executable())
}

var once sync.Once
var executable string

func Executable() string {
	once.Do(func() {
		path, err := os.Executable()
		if err != nil {
			log.Fatalf("Failed to get executable path: %v", err)
			return
		}
		executable = utils.NormalizePath(path)
	})
	return executable
}

func ExecutablePath() string {
	return filepath.Dir(Executable())
}

func StartNewServer() {
	executable, err := os.Executable()
	if err != nil {
		log.Printf("Failed to get executable path: %v", err)
		return
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Printf("Failed to get working directory: %v", err)
		return
	}

	args := os.Args[1:]
	procAttr := &os.ProcAttr{
		Dir:   wd,
		Files: []*os.File{nil, os.Stdout, os.Stderr},
		Env:   os.Environ(),
	}

	if args[0] != "--daemon" {
		// starting from client, need to start server with --daemon flag and set working directory to executable directory
		args = []string{"--daemon"}
		procAttr.Dir = filepath.Dir(executable)
	}

	process, err := os.StartProcess(executable, append([]string{executable}, args...), procAttr)
	if err != nil {
		log.Printf("Failed to start new process: %v", err)
		return
	}

	log.Printf("Started new process with PID: %d", process.Pid)
}
