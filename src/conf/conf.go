package conf

import (
	"os"
	"path/filepath"
	"runtime"
	"search-indexer/running"
	fsutils "search-indexer/utils/fs"

	"gopkg.in/yaml.v3"
)

const (
	DefaultMaxFileSize  = 2 * 1024 * 1024
	DefaultIndexWorkers = 4
	DefaultPort         = 13134
)

type Exclude struct {
	UseGitIgnore bool     `yaml:"use_git_ignore" json:"use_git_ignore"`
	Customized   []string `yaml:"customized" json:"customized"` // Won't be used if enable_git_ignore is true
}

type Filters struct {
	Exclude Exclude  `yaml:"exclude" json:"exclude"`
	Include []string `yaml:"include" optional:"true" json:"include"`
}

type Global struct {
	HomePath string `yaml:"home_path"`
	Port     int    `yaml:"port"`
}

type Client struct {
}

type Server struct {
	MaxFileSize  int64   `yaml:"max_file_size"`
	IndexWorkers int     `yaml:"index_workers"`
	Filters      Filters `yaml:"filters"`

	LoggingStdout bool `yaml:"logging_stdout"`
}

type Conf struct {
	Global Global `yaml:"global"`
	Client Client `yaml:"client"`
	Server Server `yaml:"server"`

	ForTest struct {
		Path string `yaml:"path"`
	} `yaml:"for_test"`
}

var conf *Conf

func checkMode() {
	if !running.IsServerMode() {
		panic("server conf is not accessible in client mode!")
	}
}

func Get() *Conf {
	checkMode()
	return conf
}

var serverConf *string

func Load() error {
	checkMode()

	search := []string{
		"./config.local.yaml",
		"./config.yaml",
		filepath.Join(running.DefaultRootPath(), "config.yaml"),
		"./config.example.yaml",
	}

	for _, path := range search {
		serverConf = &path
		if _, err := os.Stat(path); err == nil {
			break
		}
	}

	conf = &Conf{
		Global: Global{
			HomePath: running.DefaultRootPath(),
			Port:     DefaultPort,
		},
		Client: Client{},
		Server: Server{
			MaxFileSize:  DefaultMaxFileSize,
			IndexWorkers: DefaultIndexWorkers,
		},
	}

	confBytes := fsutils.ReadFileWithDefault(*serverConf, []byte(``))
	if err := yaml.Unmarshal(confBytes, conf); err != nil {
		return err
	}

	if conf.Server.IndexWorkers <= 0 || conf.Server.IndexWorkers > runtime.NumCPU() {
		conf.Server.IndexWorkers = runtime.NumCPU()
	}

	if conf.Server.MaxFileSize <= 0 {
		conf.Server.MaxFileSize = DefaultMaxFileSize
	}

	if conf.Global.Port <= 0 || conf.Global.Port > 65535 {
		conf.Global.Port = DefaultPort
	}

	return nil
}
