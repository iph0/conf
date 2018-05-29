// Copyright (c) 2018, Eugene Ponizovsky, <ponizovsky@gmail.com>. All rights
// reserved. Use of this source code is governed by a MIT License that can
// be found in the LICENSE file.

/*
Package envconf is configuration provider for the conf package, that imports
environment variables to the root of configuration tree. Source patterns for
this provider are regular expressions and must begins with "env:".

 env:^MYAPP_.*"
 env:.*
*/
package envconf

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	drvName = "env"
	errPref = "envconf"
)

// EnvProvider type represents configuration provider instance.
type EnvProvider struct{}

// Name method returns the provider name.
func (d *EnvProvider) Name() string {
	return drvName
}

// Load method imports environment variables to configuration tree.
func (d *EnvProvider) Load(pattern string) (interface{}, error) {
	if pattern == "" {
		return nil, fmt.Errorf("%s: empty pattern specified", errPref)
	}

	patParsed := strings.SplitN(pattern, ":", 2)
	var reStr string

	if len(patParsed) < 2 {
		reStr = patParsed[0]
	} else if patParsed[0] != "" && patParsed[0] != drvName {
		return nil, fmt.Errorf("%s: unknown pattern specified: %s", errPref,
			patParsed[0])
	} else {
		reStr = patParsed[1]
	}

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
