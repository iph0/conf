// Copyright (c) 2024, Eugene Ponizovsky, <ponizovsky@gmail.com>. All rights
// reserved. Use of this source code is governed by a MIT License that can
// be found in the LICENSE file.

/*
Package fileconf is configuration loader for the conf package. It loads
configuration layers from YAML, JSON or TOML files. Configuration locators for
this loader are relative pathes or glob patterns. See standart package
path/filepath for more information about syntax of glob patterns. Here some
examples:

	file:myapp/dirs.yml
	file:myapp/servers.toml
	file:myapp/*.json
	file:myapp/*.*
*/
package fileconf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/BurntSushi/toml"
	"github.com/iph0/conf/v2"
	yaml "gopkg.in/yaml.v3"
)

const errPref = "fileconf"

var (
	parsers = map[string]func(bytes []byte) ([]any, error){
		"yml":  unmarshalYAML,
		"yaml": unmarshalYAML,
		"json": unmarshalJSON,
		"toml": unmarshalTOML,
	}

	fileExtRe = regexp.MustCompile("\\.([^.]+)$")
)

// Loader loads configuration layers from YAML, JSON and TOML configuration files.
type Loader struct {
	dirs []string
}

// NewLoader method creates new loader instance. Method accepts a list of
// directories, in which the loader will search configuration files. The merge
// priority of loaded configuration layers depends on the order of directories.
// Layers loaded from rightmost directory have highest priority.
func NewLoader(dirs ...string) *Loader {
	if len(dirs) == 0 {
		panic(fmt.Errorf("%s: no directories specified", errPref))
	}

	return &Loader{
		dirs: dirs,
	}
}

// Load method loads configuration layer from YAML, JSON and TOML configuration files.
func (l *Loader) Load(pattern string) ([]any, error) {
	var allLayers []any

	for _, dir := range l.dirs {
		absPattern := filepath.Join(dir, pattern)
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
			bytes, err := io.ReadAll(f)

			if err != nil {
				return nil, fmt.Errorf("%s: %s", errPref, err)
			}

			layers, err := parser(bytes)

			if err != nil {
				return nil, fmt.Errorf("%s: %s", errPref, err)
			}

			for _, layer := range layers {
				if layer != nil {
					allLayers = append(allLayers, layer)
				}
			}
		}
	}

	return allLayers, nil
}

func unmarshalYAML(rawData []byte) ([]any, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(rawData))
	var layers []any

	for {
		var layer any
		err := decoder.Decode(&layer)

		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		if d, ok := layer.(map[any]any); ok {
			layer = conformMap(d)
		}

		layers = append(layers, layer)
	}

	return layers, nil
}

func unmarshalJSON(bytes []byte) ([]any, error) {
	var layer any
	err := json.Unmarshal(bytes, &layer)

	if err != nil {
		return nil, err
	}

	return []any{layer}, nil
}

func unmarshalTOML(bytes []byte) ([]any, error) {
	var layer any
	err := toml.Unmarshal(bytes, &layer)

	if err != nil {
		return nil, err
	}

	return []any{layer}, nil
}

func conformMap(m map[any]any) conf.M {
	cm := make(conf.M)

	for key, value := range m {
		if v, ok := value.(map[any]any); ok {
			value = conformMap(v)
		}

		keyStr := fmt.Sprintf("%v", key)
		cm[keyStr] = value
	}

	return cm
}
