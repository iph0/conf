// Package baseconf is the loader driver for conf package, that loads
// configrartion from YAML and JSON files.
package baseconf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/iph0/conf"
	"github.com/iph0/conf/merger"
	yaml "gopkg.in/yaml.v2"
)

// BaseDriver type represents configuration base driver instance
type BaseDriver struct {
	dirs []string
}

var fileExts = []string{"yml", "json"}

// NewDriver method creates new configuration loader driver
func NewDriver() conf.Driver {
	rawDirs := os.Getenv("GOCONF_PATH")
	dirs := make([]string, 0, 5)

	if rawDirs != "" {
		dirs = append(dirs, strings.Split(rawDirs, ":")...)
	}
	dirs = append(dirs, ".")

	return &BaseDriver{dirs}
}

// Load method loads configuration sections form YAML files
func (d *BaseDriver) Load(sec string) (map[string]interface{}, error) {
	config := make(map[string]interface{})

	for _, dir := range d.dirs {
		for _, ext := range fileExts {
			path := fmt.Sprintf("%s.%s", sec, ext)
			path = filepath.Join(dir, path)

			_, err := os.Stat(path)

			if err != nil {
				if os.IsNotExist(err) {
					continue
				}

				return nil, err
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

			var data map[string]interface{}

			if ext == "yml" {
				data, err = unmarshalYAML(bytes)
			} else { // json
				data, err = unmarshalJSON(bytes)
			}

			if err != nil {
				return nil, err
			}

			iConfig := merger.Merge(config, data)
			config = iConfig.(map[string]interface{})
		}
	}

	if len(config) == 0 {
		return nil, nil
	}

	return config, nil
}

func unmarshalYAML(bytes []byte) (map[string]interface{}, error) {
	var iData interface{}
	err := yaml.Unmarshal(bytes, &iData)

	if err != nil {
		return nil, err
	}

	iDataV := reflect.ValueOf(iData)

	if !iDataV.IsValid() {
		return nil, nil
	}

	return convertMap(iData.(map[interface{}]interface{})), nil
}

func unmarshalJSON(bytes []byte) (map[string]interface{}, error) {
	var iData interface{}
	err := json.Unmarshal(bytes, &iData)

	if err != nil {
		return nil, err
	}

	return iData.(map[string]interface{}), nil
}

func convertMap(f map[interface{}]interface{}) map[string]interface{} {
	fT := reflect.ValueOf(f).Type()
	t := make(map[string]interface{})

	for k, v := range f {
		if v == nil {
			continue
		}

		kS := fmt.Sprintf("%v", k)
		vT := reflect.ValueOf(v).Type()

		if fT == vT {
			t[kS] = convertMap(v.(map[interface{}]interface{}))
			continue
		}

		t[kS] = v
	}

	return t
}
