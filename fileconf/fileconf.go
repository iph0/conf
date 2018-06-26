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
	"strings"

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

// FileProvider TODO
type FileProvider struct {
	dirs []string
}

// NewProvider TODO
func NewProvider() (conf.Provider, error) {
	rawDirs := os.Getenv("GOCONF_PATH")
	var dirs []string

	if rawDirs == "" {
		return nil, fmt.Errorf("%s: GOCONF_PATH not set", errPref)
	}

	dirs = strings.Split(rawDirs, ":")

	return &FileProvider{
		dirs: dirs,
	}, nil
}

// Load TODO
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

// Close TODO
func (p *FileProvider) Close() {
	// TODO
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
