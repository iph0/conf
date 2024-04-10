package main

import (
	"fmt"

	"github.com/iph0/conf/v2"
	"github.com/iph0/conf/v2/envconf"
	"github.com/iph0/conf/v2/fileconf"
)

// MyAppConfig example type
type MyAppConfig struct {
	RootDir      string
	TemplatesDir string
	SessionsDir  string
	MediaFormats []string
	MediaDirs    []string
}

// DBConfig example type
type DBConfig map[string]DBConnector

// DBConnector example type
type DBConnector struct {
	Host     string
	Port     int
	DBName   string
	Username string
	Password string
}

// GenericConfig example type
type LogConfig struct {
	Tag    string
	Level  string
	Format string
}

func main() {
	fileLdr := fileconf.NewLoader("etc")
	envLdr := envconf.NewLoader()

	configProc := conf.NewProcessor(
		conf.ProcessorConfig{
			Loaders: map[string]conf.Loader{
				"file": fileLdr,
				"env":  envLdr,
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

	fmt.Printf("%+v\n\n", myAppConfig)

	var dbConfig DBConfig
	err = conf.Decode(configRaw["db"], &dbConfig)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%+v\n\n", dbConfig)

	var genericConfig LogConfig
	err = conf.Decode(configRaw["log"], &genericConfig)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%+v\n", genericConfig)
}
