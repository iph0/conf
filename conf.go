package conf

import (
	"fmt"
	"strings"

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
	Load(string) (interface{}, error)
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
configuration tree. Path pattern to configuration sections is specified as a
string and must begins with driver name. Format of the path pattern depends on
the loader driver. Here some examples:

 file:myapp/dirs.yml
 file:myapp/*.json
 file:myapp/*.*

 env:MYAPP_ROOTDIR
 env:MYAPP_.*
 env:.*

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
			if sec == "" {
				return nil, fmt.Errorf("%s: empty pattern specified", errPref)
			}

			tokens := strings.SplitN(sec, ":", 2)

			if len(tokens) < 2 || tokens[0] == "" {
				return nil, fmt.Errorf("%s: driver name not specified: %s",
					errPref, sec)
			}

			drvName := tokens[0]
			pattern := tokens[1]

			driver, ok := l.drivers[drvName]

			if !ok {
				return nil, fmt.Errorf("%s: unknown driver name: %s", errPref,
					drvName)
			}

			data, err := driver.Load(pattern)

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
