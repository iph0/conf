package main

import (
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/iph0/conf"
	"github.com/iph0/conf/envconf"
	"github.com/iph0/conf/fileconf"
)

type myappConfig struct {
	MediaFormats []string
	Dirs         dirsConfig
}

type dirsConfig struct {
	RootDir      string
	TemplatesDir string
	SessionsDir  string
	MediaDirs    []string
}

type mysqlConfig struct {
	DefaultOptions mysqlOptions
	Connectors     map[string]mysqlConnector
}

type mysqlConnector struct {
	Host     string
	Port     int
	DBName   string
	Username string
	Password string
	Options  mysqlOptions
}

type mysqlOptions struct {
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
		"file:mysql.json",
		"env:^MYAPP_",
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	var myappConf myappConfig
	err = conf.Decode(configRaw["myapp"], &myappConf)

	if err != nil {
		fmt.Println(err)
		return
	}

	spew.Dump(myappConf)

	var mysqlConf mysqlConfig
	err = conf.Decode(configRaw["mysql"], &mysqlConf)

	if err != nil {
		fmt.Println(err)
		return
	}

	spew.Dump(mysqlConf)
}
