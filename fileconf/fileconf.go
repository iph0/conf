// Copyright (c) 2018, Eugene Ponizovsky, <ponizovsky@gmail.com>. All rights
// reserved. Use of this source code is governed by a MIT License that can
// be found in the LICENSE file.

/*
Package fileconf is the loader driver for the conf package, that loads
configuration data from YAML and JSON files. fileconf searches configuration
files in directories specified by GOCONF_PATH environment variable. In
GOCONF_PATH you can specify one or more directories separated by ":" symbol.

 GOCONF_PATH=/home/username/etc/go:/etc/go

If no directories specified in GOCONF_PATH, then driver searches
configuration files in the current directory. Loading patterns for this driver
must begins with "file:". The syntax of loading patterns is the same as for
patterns in Match method of the standart package path/filepath. Here some
examples:

 file:myapp/dirs.yml
 file:myapp/*.json
 file:myapp/*.*
*/
package fileconf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/iph0/conf"
	"github.com/iph0/conf/merger"
	yaml "gopkg.in/yaml.v2"
)

// FileDriver type represents configuration loader driver instance
type FileDriver struct {
	dirs      []string
	mandatory bool
}

const (
	drvName = "file"
	errPref = "fileconf"
)

var (
	parsers = map[string]func(bytes []byte) (interface{}, error){
		"yml":  unmarshalYAML,
		"yaml": unmarshalYAML,
		"json": unmarshalJSON,
	}

	fileExtRe = regexp.MustCompile("\\.([^.]+)$")
)

// NewDriver method creates new configuration loader driver. If mandatory flag is
// true, then not found configuration sections will raise errors.
func NewDriver(mandatory bool) conf.LoaderDriver {
	rawDirs := os.Getenv("GOCONF_PATH")
	var dirs []string

	if rawDirs != "" {
		dirs = strings.Split(rawDirs, ":")
	} else {
		dirs = []string{"."}
	}

	return &FileDriver{
		dirs:      dirs,
		mandatory: mandatory,
	}
}

// Name method returns the driver name, that used by loader to determine, which
// configuration section must be loaded by this driver.
func (d *FileDriver) Name() string {
	return drvName
}

// Load method loads configuration sections form YAML and JSON files
func (d *FileDriver) Load(pattern string) (interface{}, error) {
	if pattern == "" {
		return nil, fmt.Errorf("%s: empty pattern specified", errPref)
	}

	patParsed := strings.SplitN(pattern, ":", 2)
	var globPattern string

	if len(patParsed) < 2 {
		globPattern = patParsed[0]
	} else if patParsed[0] != "" && patParsed[0] != drvName {
		return nil, fmt.Errorf("%s: unknown pattern specified: %s", errPref,
			patParsed[0])
	} else {
		globPattern = patParsed[1]
	}

	var config interface{}
	notFoundCnt := 0

	for _, dir := range d.dirs {
		absPattern := filepath.Join(dir, globPattern)
		pathes, err := filepath.Glob(absPattern)

		if err != nil {
			return nil, fmt.Errorf("%s: %s", errPref, err)
		}

		if len(pathes) == 0 {
			notFoundCnt++
			continue
		}

		for _, path := range pathes {
			matches := fileExtRe.FindStringSubmatch(path)

			if matches == nil {
				return nil, fmt.Errorf("%s: file extension not specified: %s",
					errPref, path)
			}

			ext := matches[1]
			parser, ok := parsers[ext]

			if !ok {
				return nil, fmt.Errorf("%s: unknown file extension .%s",
					errPref, ext)
			}

			f, err := os.Open(path)

			if err != nil {
				return nil, fmt.Errorf("%s: %s", errPref, err)
			}

			bytes, err := ioutil.ReadAll(f)

			if err != nil {
				return nil, fmt.Errorf("%s: %s", errPref, err)
			}

			f.Close()

			data, err := parser(bytes)

			if err != nil {
				return nil, fmt.Errorf("%s: %s", errPref, err)
			}

			config = merger.Merge(config, data)
		}
	}

	if d.mandatory && notFoundCnt == len(d.dirs) {
		return nil, fmt.Errorf("%s: nothing found by pattern %s in %s", errPref,
			pattern, strings.Join(d.dirs, ", "))
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
