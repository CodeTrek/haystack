package conf

import (
	"haystack/shared/running"
	"haystack/shared/types"
	fsutils "haystack/utils/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

const (
	DefaultMaxFileSize  = 2 * 1024 * 1024
	DefaultIndexWorkers = 4
	DefaultPort         = 13134

	DefaultMaxResults        = 5000
	DefaultMaxResultsPerFile = 1000
)

var (
	DefaultInclude = []string{"*.cc", "*.c", "*.hpp", "*.cpp", "*.h", "*.md", "*.js", "*.ts", "*.txt", "*.mm", "*.java",
		"*.cs", "*.py", "*.kt", "*.go", "*.rb", "*.php", "*.html", "*.css", "*.yaml", "*.yml", "*.toml", "*.xml", "*.sql",
		"*.sh", "Makefile", "*.bat", "*.ps1", "*.sln", "*.json", "*.vcxproj", "*.vcproj", "*.vcxproj.filters"}
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
	MaxFileSize  int64             `yaml:"max_file_size"`
	IndexWorkers int               `yaml:"index_workers"`
	Filters      Filters           `yaml:"filters"`
	SearchLimit  types.SearchLimit `yaml:"search_limit"`

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

func Load() error {
	checkMode()
	homePath := filepath.Join(running.UserHomeDir(), ".haystack")

	search := []string{
		"./config.local.yaml",
		"./config.yaml",
		filepath.Join(homePath, "config.yaml"),
		"./config.example.yaml",
	}

	var confFile *string
	for _, path := range search {
		confFile = &path
		if _, err := os.Stat(path); err == nil {
			break
		}
	}

	conf = &Conf{
		Global: Global{
			HomePath: homePath,
			Port:     DefaultPort,
		},
		Client: Client{},
		Server: Server{
			MaxFileSize:  DefaultMaxFileSize,
			IndexWorkers: DefaultIndexWorkers,
			SearchLimit: types.SearchLimit{
				MaxResults:        DefaultMaxResults,
				MaxResultsPerFile: DefaultMaxResultsPerFile,
			},
		},
	}

	confBytes := fsutils.ReadFileWithDefault(*confFile, []byte(``))
	if err := yaml.Unmarshal(confBytes, conf); err != nil {
		return err
	}

	if conf.Global.HomePath == "" {
		conf.Global.HomePath = homePath
	}

	if err := os.Mkdir(conf.Global.HomePath, 0755); err != nil {
		if !os.IsExist(err) {
			log.Fatalf("Failed to create home directory: %v", err)
			return err
		}
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

	if conf.Server.SearchLimit.MaxResults <= 0 || conf.Server.SearchLimit.MaxResults > DefaultMaxResults {
		conf.Server.SearchLimit.MaxResults = DefaultMaxResults
	}

	if conf.Server.SearchLimit.MaxResultsPerFile <= 0 || conf.Server.SearchLimit.MaxResultsPerFile > DefaultMaxResultsPerFile {
		conf.Server.SearchLimit.MaxResultsPerFile = DefaultMaxResultsPerFile
	}

	return nil
}
