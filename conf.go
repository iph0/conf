// Copyright (c) 2018, Eugene Ponizovsky, <ponizovsky@gmail.com>. All rights
// reserved. Use of this source code is governed by a MIT License that can
// be found in the LICENSE file.

// Package conf loads configuration sections from different sources and merges
// them into the one configuration tree.
package conf

import (
	"fmt"

	"github.com/iph0/conf/merger"
)

// Loader loads configuration sections from different sources using different
// loader drivers.
type Loader struct {
	drivers []Driver
}

// Driver interace is the interface for all configuration loader drivers.
type Driver interface {
	Load(string) (map[string]interface{}, error)
}

// NewLoader method creates a new configuration loader.
func NewLoader(drivers ...Driver) *Loader {
	if len(drivers) == 0 {
		panic("no drivers specified")
	}

	return &Loader{drivers}
}

// Load method loads configuration sections and merges them to the one
// configuration tree.
func (l *Loader) Load(sections ...interface{}) (map[string]interface{}, error) {
	config := make(map[string]interface{})

	for _, iSec := range sections {
		switch sec := iSec.(type) {
		case map[string]interface{}:
			iConfig := merger.Merge(config, sec)
			config = iConfig.(map[string]interface{})
		case string:
			var notFoundCnt int

			for _, driver := range l.drivers {
				data, err := driver.Load(sec)

				if err != nil {
					return nil, err
				}

				if data == nil {
					notFoundCnt++
					continue
				}

				iConfig := merger.Merge(config, data)
				config = iConfig.(map[string]interface{})
			}

			if notFoundCnt == len(l.drivers) {
				return nil, fmt.Errorf("configuration section \"%s\" not found", sec)
			}
		default:
			panic(fmt.Sprintf("%T is invalid type for configuration section", sec))
		}
	}

	return config, nil
}
