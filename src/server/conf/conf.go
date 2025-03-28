package conf

import (
	"search-indexer/runtime"
	fsutils "search-indexer/utils/fs"

	"gopkg.in/yaml.v3"
)

type Exclude struct {
	UseGitIgnore bool     `yaml:"use_git_ignore"`
	Customized   []string `yaml:"customized"` // Won't be used if enable_git_ignore is true
}

type Index struct {
	Path    string   `yaml:"path"`
	Exclude Exclude  `yaml:"exclude"`
	Files   []string `yaml:"files"`
}

type Conf struct {
	Indexes []Index `yaml:"indexes"`
	Port    int     `yaml:"port"`
}

var conf *Conf

func checkMode() {
	if !runtime.IsServerMode() {
		panic("server conf is not accessible in client mode!")
	}
}

func Get() *Conf {
	checkMode()
	return conf
}

func Load() error {
	checkMode()

	conf = &Conf{
		Port: runtime.DefaultListenPort(),
	}

	confBytes := fsutils.ReadFileWithDefault(runtime.ServerConf(), []byte(``))
	if err := yaml.Unmarshal(confBytes, conf); err != nil {
		return err
	}

	return nil
}
