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
Load method loads configuration sections using specific loading patterns for
each destination and then merges them to the one configuration tree. Loading
patterns must begins with driver name. Format of the loading patterns depends on
the loader drivers. Here some examples:

 file:myapp/dirs.yml
 file:myapp/*.json
 file:myapp/*.*

 env:^MYAPP_.*"
 env:.*

Also you can specify configuration section as map[string]interface{}. In this
case configuration section will be simple merged to configuration tree as is.
*/
func (l *Loader) Load(sections ...interface{}) (interface{}, error) {
	var config interface{}

	for _, iSec := range sections {
		switch sec := iSec.(type) {
		case map[string]interface{}:
			iConfig := merger.Merge(config, sec)
			config = iConfig.(map[string]interface{})
		case string:
			pattern := sec

			if pattern == "" {
				return nil, fmt.Errorf("%s: empty pattern specified", errPref)
			}

			patParsed := strings.SplitN(pattern, ":", 2)

			if len(patParsed) < 2 || patParsed[0] == "" {
				return nil, fmt.Errorf("%s: missing driver name in pattern: %s",
					errPref, pattern)
			}

			driver, ok := l.drivers[patParsed[0]]

			if !ok {
				return nil, fmt.Errorf("%s: unknown pattern specified: %s", errPref,
					pattern)
			}

			data, err := driver.Load(patParsed[1])

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
