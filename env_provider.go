package conf

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type envProvider struct{}

func newEnvProvider() Provider {
	return &envProvider{}
}

func (p *envProvider) Watch(notifier UpdatesNotifier) {
	// TODO
}

func (p *envProvider) Load(loc *Locator) (interface{}, error) {
	reStr := loc.BareLocator
	re, err := regexp.Compile(reStr)

	if err != nil {
		return nil, fmt.Errorf("%s: %s", errPref, err)
	}

	pairs := os.Environ()
	config := make(map[string]interface{})

	for _, pairRaw := range pairs {
		pair := strings.SplitN(pairRaw, "=", 2)

		key := pair[0]
		value := pair[1]

		if re.MatchString(key) {
			config[key] = value
		}
	}

	return config, nil
}

func (p *envProvider) Close() {
	// TODO
}
