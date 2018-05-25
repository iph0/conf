// Copyright (c) 2018, Eugene Ponizovsky, <ponizovsky@gmail.com>. All rights
// reserved. Use of this source code is governed by a MIT License that can
// be found in the LICENSE file.

package conf

import (
	"fmt"
	"net/url"

	"github.com/iph0/conf/merger"
)

// Loader loads configuration sections from different sources using different
// loader drivers.
type Loader struct {
	drivers map[string]LoaderDriver
}

// LoaderDriver interface is the interface for all configuration loader drivers.
type LoaderDriver interface {
	Name() string
	Load(*url.URL) (interface{}, error)
}

const errPref = "conf"

// NewLoader method creates a new configuration loader.
func NewLoader(drivers ...LoaderDriver) *Loader {
	if len(drivers) == 0 {
		panic(fmt.Errorf("%s: no drivers specified", errPref))
	}

	driverMap := make(map[string]LoaderDriver)

	for _, driver := range drivers {
		name := driver.Name()
		driverMap[name] = driver
	}

	return &Loader{driverMap}
}

/*
Load method loads configuration sections and merges them to the one
configuration tree. Location of configuration section must be specified as
URI. URI scheme and form depends on the loader driver.

 file:///myapp/dirs.yml
 file:///myapp/*.json
 env:///MYAPP_*
 env:///*

Also you can specify configuration section as map[string]interface{}. In this
case configuration section will be loaded and merged to configuration tree as is.
*/
func (l *Loader) Load(sections ...interface{}) (interface{}, error) {
	var config interface{}

	for _, iSec := range sections {
		switch sec := iSec.(type) {
		case map[string]interface{}:
			iConfig := merger.Merge(config, sec)
			config = iConfig.(map[string]interface{})
		case string:
			uriAddr, err := url.Parse(sec)

			if err != nil {
				return nil, err
			}

			if uriAddr.Scheme == "" {
				return nil, fmt.Errorf("%s: URI scheme not specified: %s", errPref,
					uriAddr)
			}

			driver, ok := l.drivers[uriAddr.Scheme]

			if !ok {
				return nil, fmt.Errorf("%s: unknown URI scheme %s://", errPref,
					uriAddr.Scheme)
			}

			data, err := driver.Load(uriAddr)

			if err != nil {
				return nil, err
			}

			if data == nil {
				continue
			}

			config = merger.Merge(config, data)
		default:
			panic(fmt.Errorf("%s: %T is invalid type for configuration section",
				errPref, sec))
		}
	}

	processor := &Processor{}
	processor.Process(config)

	return config, nil
}
