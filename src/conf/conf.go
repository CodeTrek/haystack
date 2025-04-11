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
	DefaultMaxResultsPerFile = 500

	DefaultClientMaxResults        = 500
	DefaultClientMaxResultsPerFile = 50

	DefaultMaxSearchWildcardLength  = 24
	DefaultMaxSearchKeywordDistance = 32
)

var (
	DefaultInclude = []string{"*.cc", "*.c", "*.hpp", "*.cpp", "*.h", "*.md", "*.js", "*.ts", "*.txt", "*.mm", "*.java",
		"*.cs", "*.py", "*.kt", "*.go", "*.rb", "*.php", "*.html", "*.css", "*.yaml", "*.yml", "*.toml", "*.xml", "*.sql",
		"*.sh", "Makefile", "*.bat", "*.ps1", "*.sln", "*.json", "*.vcxproj", "*.vcproj", "*.vcxproj.filters"}
)

type Exclude struct {
	UseGitIgnore bool     `yaml:"use_git_ignore,omitempty" json:"use_git_ignore,omitempty"`
	Customized   []string `yaml:"customized,omitempty"     json:"customized,omitempty"` // Won't be used if enable_git_ignore is true
}

type Filters struct {
	Exclude Exclude  `yaml:"exclude,omitempty" json:"exclude,omitempty"`
	Include []string `yaml:"include,omitempty" optional:"true" json:"include,omitempty"`
}

type Global struct {
	HomePath string `yaml:"home_path,omitempty"`
	Port     int    `yaml:"port,omitempty"`
}

type Client struct {
	DefaultWorkspace string            `yaml:"default_workspace,omitempty"`
	DefaultLimit     types.SearchLimit `yaml:"default_limit,omitempty"`
}

type Search struct {
	MaxWildcardLength  int               `yaml:"max_wildcard_length,omitempty"`
	MaxKeywordDistance int               `yaml:"max_keyword_distance,omitempty"`
	Limit              types.SearchLimit `yaml:"limit,omitempty"`
}

type Server struct {
	MaxFileSize  int64   `yaml:"max_file_size,omitempty"`
	IndexWorkers int     `yaml:"index_workers,omitempty"`
	Filters      Filters `yaml:"filters,omitempty"`
	Search       Search  `yaml:"search,omitempty"`

	LoggingStdout bool `yaml:"logging_stdout,omitempty"`
}

type Conf struct {
	Global Global `yaml:"global,omitempty"`
	Client Client `yaml:"client,omitempty"`
	Server Server `yaml:"server,omitempty"`

	ForTest struct {
		Path string `yaml:"path,omitempty"`
	} `yaml:"for_test,omitempty"`
}

var conf *Conf

var confFile string

func Get() *Conf {
	return conf
}

func Load() error {
	homePath := filepath.Join(running.UserHomeDir(), ".haystack")

	search := []string{
		filepath.Join(running.ExecutablePath(), "config.local.yaml"),
		filepath.Join(homePath, "config.yaml"),
		filepath.Join(running.ExecutablePath(), "config.yaml"),
		filepath.Join(running.InstallPath(), "config.yaml"),
		filepath.Join(running.ExecutablePath(), "config.example.yaml"),
	}

	for _, path := range search {
		if _, err := os.Stat(path); err == nil {
			confFile = path
			break
		}
	}

	if confFile == "" {
		// Create a new config file
		confFile = filepath.Join(homePath, "config.yaml")
	}

	conf = &Conf{
		Global: Global{
			HomePath: homePath,
			Port:     DefaultPort,
		},
		Client: Client{
			DefaultWorkspace: "",
			DefaultLimit: types.SearchLimit{
				MaxResults:        DefaultClientMaxResults,
				MaxResultsPerFile: DefaultClientMaxResultsPerFile,
			},
		},
		Server: Server{
			MaxFileSize:  DefaultMaxFileSize,
			IndexWorkers: DefaultIndexWorkers,
			Search: Search{
				MaxWildcardLength:  DefaultMaxSearchWildcardLength,
				MaxKeywordDistance: DefaultMaxSearchKeywordDistance,
				Limit: types.SearchLimit{
					MaxResults:        DefaultMaxResults,
					MaxResultsPerFile: DefaultMaxResultsPerFile,
				},
			},
		},
	}

	confBytes := fsutils.ReadFileWithDefault(confFile, []byte(``))
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

	if conf.Server.Search.Limit.MaxResults <= 0 || conf.Server.Search.Limit.MaxResults > DefaultMaxResults {
		conf.Server.Search.Limit.MaxResults = DefaultMaxResults
	}

	if conf.Server.Search.Limit.MaxResultsPerFile <= 0 ||
		conf.Server.Search.Limit.MaxResultsPerFile > DefaultMaxResultsPerFile {
		conf.Server.Search.Limit.MaxResultsPerFile = DefaultMaxResultsPerFile
	}

	if conf.Server.Search.MaxWildcardLength <= 0 ||
		conf.Server.Search.MaxWildcardLength > 64 { // 64 is the maximum length of a wildcard
		conf.Server.Search.MaxWildcardLength = DefaultMaxSearchWildcardLength
	}

	if conf.Server.Search.MaxKeywordDistance <= 0 ||
		conf.Server.Search.MaxKeywordDistance > 128 { // 128 is the maximum distance of a keyword
		conf.Server.Search.MaxKeywordDistance = DefaultMaxSearchKeywordDistance
	}

	if conf.Client.DefaultLimit.MaxResults <= 0 ||
		conf.Client.DefaultLimit.MaxResults > conf.Server.Search.Limit.MaxResults {
		conf.Client.DefaultLimit.MaxResults = DefaultClientMaxResults
	}

	if conf.Client.DefaultLimit.MaxResultsPerFile <= 0 ||
		conf.Client.DefaultLimit.MaxResultsPerFile > conf.Server.Search.Limit.MaxResultsPerFile {
		conf.Client.DefaultLimit.MaxResultsPerFile = DefaultClientMaxResultsPerFile
	}

	return nil
}
