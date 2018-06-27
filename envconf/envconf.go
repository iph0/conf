// Copyright (c) 2018, Eugene Ponizovsky, <ponizovsky@gmail.com>. All rights
// reserved. Use of this source code is governed by a MIT License that can
// be found in the LICENSE file.

/*
Package envconf TODO
*/
package envconf

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/iph0/conf"
)

const errPref = "envconf"

// EnvSource TODO
type EnvSource struct{}

// NewSource TODO
func NewSource() conf.Source {
	return &EnvSource{}
}

// Load TODO
func (p *EnvSource) Load(loc *conf.Locator) (interface{}, error) {
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
