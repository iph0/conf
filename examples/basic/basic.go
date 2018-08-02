package main

import (
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/iph0/conf"
	"github.com/iph0/conf/envconf"
	"github.com/iph0/conf/fileconf"
	mapstruct "github.com/mitchellh/mapstructure"
)

type myAppConfig struct {
	MediaFormats []string
	Dirs         dirsConfig
	DB           databaseConfig
}

type dirsConfig struct {
	RootDir      string
	TemplatesDir string
	SessionsDir  string
	MediaDirs    []string
}

type databaseConfig struct {
	DefaultOptions databaseOpts
	Connectors     map[string]databaseConnector
}

type databaseConnector struct {
	Host     string
	Port     string
	DBName   string
	Username string
	Password string
	Options  databaseOpts
}

type databaseOpts struct {
	PrintWarn  bool
	PrintError bool
	RaiseError bool
}

func init() {
	os.Setenv("MYAPP_ROOTDIR", "/myapp")
	os.Setenv("MYAPP_DBPASS_STAT", "stat_writer_pass")
	os.Setenv("MYAPP_DBPASS_METRIC", "metric_writer_pass")
}

func main() {
	envLdr := envconf.NewLoader()
	fileLdr, err := fileconf.NewLoader("etc")

	if err != nil {
		fmt.Println(err)
		return
	}

	configProc := conf.NewProcessor(
		conf.ProcessorConfig{
			Loaders: map[string]conf.Loader{
				"env":  envLdr,
				"file": fileLdr,
			},
		},
	)

	configRaw, err := configProc.Load(
		"file:dirs.yml",
		"file:db.json",
		"env:^MYAPP_",
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	var config myAppConfig
	err = mapstruct.Decode(configRaw["myapp"], &config)

	if err != nil {
		fmt.Println(err)
		return
	}

	spew.Dump(config)
}
