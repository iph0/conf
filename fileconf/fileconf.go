// Copyright (c) 2018, Eugene Ponizovsky, <ponizovsky@gmail.com>. All rights
// reserved. Use of this source code is governed by a MIT License that can
// be found in the LICENSE file.

/*Package fileconf is the loader driver for the conf package, that loads
configuration data from YAML and JSON files.

fileconf searches configuration files in directories specified by GOCONF_PATH
environment variable. In GOCONF_PATH you can specify one or more directories
separated by ":" symbol.

 GOCONF_PATH=/home/username/etc/go:/etc/go

If no directories specified in GOCONF_PATH, then fileconf searches
configuration files in the current directory.
*/
package fileconf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/iph0/conf"
	"github.com/iph0/conf/merger"
	yaml "gopkg.in/yaml.v2"
)

// FileLoader type represents configuration loader driver instance
type FileLoader struct {
	dirs      []string
	mandatory bool
}

const (
	driverName = "file"
	errPref    = "fileconf"
)

var (
	parsers = map[string]func(bytes []byte) (interface{}, error){
		"yml":  unmarshalYAML,
		"yaml": unmarshalYAML,
		"json": unmarshalJSON,
	}

	fileExtRe = regexp.MustCompile("\\.([^.]+)$")
)

// NewLoaderDriver method creates new configuration loader driver
func NewLoaderDriver(mandatory bool) conf.LoaderDriver {
	rawDirs := os.Getenv("GOCONF_PATH")
	var dirs []string

	if rawDirs != "" {
		dirs = strings.Split(rawDirs, ":")
	} else {
		dirs = []string{"."}
	}

	return &FileLoader{
		dirs:      dirs,
		mandatory: mandatory,
	}
}

// Name method returns the driver name, that used by loader to determine, which
// configuration section must be loaded by this driver.
func (d *FileLoader) Name() string {
	return driverName
}

// Load method loads configuration sections form YAML and JSON files
func (d *FileLoader) Load(uriAddr *url.URL) (interface{}, error) {
	if uriAddr.Scheme != driverName {
		return nil, fmt.Errorf("%s: unknown URL scheme %s://", errPref,
			uriAddr.Scheme)
	}

	relPath := uriAddr.Host + uriAddr.Path
	matches := fileExtRe.FindStringSubmatch(relPath)

	if matches == nil {
		return nil, fmt.Errorf("%s: file extension not specified in %s",
			errPref, uriAddr)
	}

	ext := matches[1]
	notFoundCnt := 0

	var config interface{}

	for _, dir := range d.dirs {
		pattern := filepath.Join(dir, relPath)
		pathes, err := filepath.Glob(pattern)

		if err != nil {
			return nil, err
		}

		if len(pathes) == 0 {
			notFoundCnt++
			continue
		}

		for _, path := range pathes {
			parser, ok := parsers[ext]

			if !ok {
				return nil, fmt.Errorf("%s: unknown file extension .%s",
					errPref, ext)
			}

			f, err := os.Open(path)

			if err != nil {
				return nil, err
			}

			bytes, err := ioutil.ReadAll(f)

			if err != nil {
				return nil, err
			}

			f.Close()

			data, err := parser(bytes)

			if err != nil {
				return nil, err
			}

			config = merger.Merge(config, data)
		}
	}

	if d.mandatory && notFoundCnt == len(d.dirs) {
		return nil, fmt.Errorf(
			"%s: configuration data not found for URL %s in %s",
			errPref, uriAddr, strings.Join(d.dirs, ", "),
		)
	}

	return config, nil
}

func unmarshalYAML(bytes []byte) (interface{}, error) {
	var iData interface{}
	err := yaml.Unmarshal(bytes, &iData)

	if err != nil {
		return nil, err
	}

	if iData == nil {
		return nil, nil
	}

	switch data := iData.(type) {
	case map[interface{}]interface{}:
		return convertMap(data), nil
	default:
		return data, nil
	}
}

func unmarshalJSON(bytes []byte) (interface{}, error) {
	var iData interface{}
	err := json.Unmarshal(bytes, &iData)

	if err != nil {
		return nil, err
	}

	return iData, nil
}

func convertMap(from map[interface{}]interface{}) map[string]interface{} {
	fromType := reflect.ValueOf(from).Type()
	to := make(map[string]interface{})

	for key, value := range from {
		if value == nil {
			continue
		}

		keyStr := fmt.Sprintf("%v", key)
		valType := reflect.ValueOf(value).Type()

		if fromType == valType {
			to[keyStr] = convertMap(value.(map[interface{}]interface{}))
			continue
		}

		to[keyStr] = value
	}

	return to
}
