package conf

import (
	"fmt"
	"strings"

	"github.com/iph0/conf/merger"
)

// Loader loads configuration sections from different sources using different
// configuration providers.
type Loader struct {
	providers map[string]Provider
}

// Provider interface is the interface for all configuration providers.
type Provider interface {
	Name() string
	Load(string) (interface{}, error)
}

const errPref = "conf"

// NewLoader method creates a new configuration loader.
func NewLoader(providers ...Provider) *Loader {
	if len(providers) == 0 {
		panic(fmt.Errorf("%s: no providers specified", errPref))
	}

	providerMap := make(map[string]Provider)

	for _, provider := range providers {
		name := provider.Name()
		providerMap[name] = provider
	}

	return &Loader{providerMap}
}

/*
Load method loads configuration sections using specific patterns for each source
and then merges them to the one configuration tree. Source patterns must begins
with provider name. Format of the source patterns depends on configuration
provider.

 file:myapp/dirs.yml
 file:myapp/*.json
 file:myapp/*.*

 env:^MYAPP_.*"
 env:.*

Also you can specify configuration section as map[string]interface{}. In this
case configuration section will be simple merged to configuration tree as is.
Priority of the configuration sections, listed in the Load method, increases
from the left to the right. Therefore non-zero high priority values overrides
low priority values during merging process.
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
				return nil, fmt.Errorf("%s: missing provider name in pattern: %s",
					errPref, pattern)
			}

			provider, ok := l.providers[patParsed[0]]

			if !ok {
				return nil, fmt.Errorf("%s: unknown pattern specified: %s", errPref,
					pattern)
			}

			data, err := provider.Load(patParsed[1])

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

	if config != nil {
		var err error
		proc := newProcessor(l)
		config, err = proc.Process(config)

		if err != nil {
			return nil, err
		}
	}

	return config, nil
}
