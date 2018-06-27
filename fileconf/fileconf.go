// Copyright (c) 2018, Eugene Ponizovsky, <ponizovsky@gmail.com>. All rights
// reserved. Use of this source code is governed by a MIT License that can
// be found in the LICENSE file.

/*
Package fileconf TODO
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

	"github.com/BurntSushi/toml"
	"github.com/iph0/conf"
	"github.com/iph0/merger"
	yaml "gopkg.in/yaml.v2"
)

const errPref = "fileconf"

var (
	parsers = map[string]func(bytes []byte) (interface{}, error){
		"yml":  unmarshalYAML,
		"yaml": unmarshalYAML,
		"json": unmarshalJSON,
		"toml": unmarshalTOML,
	}

	fileExtRe = regexp.MustCompile("\\.([^.]+)$")
)

// FileProvider loads configuration layers from YAML, JSON and TOML
// configuration files.
type FileProvider struct {
	dirs []string
}

// NewProvider method creates new FileProvider instance. Method accepts a list
// of directories, in which FileProvider will search configuration files. The
// merge priority of loaded configuration layers depends on the order of
// directories. Layers loaded from rightmost directory have highest priority.
func NewProvider(dirs ...string) (conf.Provider, error) {
	if len(dirs) == 0 {
		panic(fmt.Errorf("%s: no directories specified", errPref))
	}

	return &FileProvider{
		dirs: dirs,
	}, nil
}

// Load method loads configuration layer.
func (p *FileProvider) Load(loc *conf.Locator) (interface{}, error) {
	var config interface{}
	globPattern := loc.BareLocator

	for _, dir := range p.dirs {
		absPattern := filepath.Join(dir, globPattern)
		pathes, err := filepath.Glob(absPattern)

		if err != nil {
			return nil, fmt.Errorf("%s: %s", errPref, err)
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

			defer f.Close()
			bytes, err := ioutil.ReadAll(f)

			if err != nil {
				return nil, fmt.Errorf("%s: %s", errPref, err)
			}

			data, err := parser(bytes)

			if err != nil {
				return nil, fmt.Errorf("%s: %s", errPref, err)
			}

			config = merger.Merge(config, data)
		}
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
		return adaptYAMLMap(data), nil
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

func unmarshalTOML(bytes []byte) (interface{}, error) {
	var iData interface{}
	err := toml.Unmarshal(bytes, &iData)

	if err != nil {
		return nil, err
	}

	return iData, nil
}

func adaptYAMLMap(from map[interface{}]interface{}) map[string]interface{} {
	fromType := reflect.ValueOf(from).Type()
	to := make(map[string]interface{})

	for key, value := range from {
		if value == nil {
			continue
		}

		keyStr := fmt.Sprintf("%v", key)
		valType := reflect.ValueOf(value).Type()

		if fromType == valType {
			to[keyStr] = adaptYAMLMap(value.(map[interface{}]interface{}))
			continue
		}

		to[keyStr] = value
	}

	return to
}
