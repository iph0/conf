package main

import (
	"fmt"
	"os"

	"github.com/iph0/conf"
	"github.com/iph0/conf/envconf"
	"github.com/iph0/conf/fileconf"
)

// MyAppConfig example type
type MyAppConfig struct {
	MediaFormats []string
	Dirs         DirsConfig
}

// DirsConfig example type
type DirsConfig struct {
	RootDir      string
	TemplatesDir string
	SessionsDir  string
	MediaDirs    []string
}

// DBConfig example type
type DBConfig struct {
	DefaultOptions DBOptions
	Connectors     map[string]DBConnector
}

// DBConnector example type
type DBConnector struct {
	Host     string
	Port     int
	DBName   string
	Username string
	Password string
	Options  DBOptions
}

// DBOptions example type
type DBOptions struct {
	ServerPrepare bool
	ExpandArray   bool
	ErrorLevel    int
}

func init() {
	os.Setenv("MYAPP_DB_STAT_PASS", "stat_writer_pass")
	os.Setenv("MYAPP_DB_METRICS_PASS", "metrics_writer_pass")
}

func main() {
	envLdr := envconf.NewLoader()
	fileLdr := fileconf.NewLoader("etc")

	configProc := conf.NewProcessor(
		conf.ProcessorConfig{
			Loaders: map[string]conf.Loader{
				"env":  envLdr,
				"file": fileLdr,
			},
		},
	)

	configRaw, err := configProc.Load(
		"file:myapp.yml",
		"file:db.json",
		"env:^MYAPP_",
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	var myAppConfig MyAppConfig
	err = conf.Decode(configRaw["myapp"], &myAppConfig)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%#v\n\n", myAppConfig)

	var dbConfig DBConfig
	err = conf.Decode(configRaw["db"], &dbConfig)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%#v\n", dbConfig)
}
