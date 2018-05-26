// Copyright (c) 2018, Eugene Ponizovsky, <ponizovsky@gmail.com>. All rights
// reserved. Use of this source code is governed by a MIT License that can
// be found in the LICENSE file.

/*
Package envconf is the loader driver for the conf package, that imports
environment variables to configuration tree. After import, environment variables
will be available under the ENV key and can be interpolated into other
configuration parameters. Path pattern for this driver represents a regular
expression and must begins with "env:". Here some examples:

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
	rootKey = "ENV"
)

// EnvDriver type represents configuration loader driver instance
type EnvDriver struct{}

// Name method returns the driver name, that used by loader to determine, which
// configuration section must be loaded by this driver.
func (d *EnvDriver) Name() string {
	return drvName
}

// Load method imports environment variables to configuration tree.
func (d *EnvDriver) Load(pattern string) (interface{}, error) {
	if pattern == "" {
		return nil, fmt.Errorf("%s: empty pattern specified", errPref)
	}

	tokens := strings.SplitN(pattern, ":", 2)

	if len(tokens) < 2 || tokens[0] == "" {
		return nil, fmt.Errorf("%s: driver name not specified: %s",
			errPref, pattern)
	} else if tokens[0] != drvName {
		return nil, fmt.Errorf("%s: unknown driver name: %s", errPref,
			tokens[0])
	}

	re, err := regexp.Compile(tokens[1])

	if err != nil {
		return nil, fmt.Errorf("%s: %s", errPref, err)
	}

	pairs := os.Environ()
	environ := make(map[string]interface{})

	for _, pairRaw := range pairs {
		pair := strings.SplitN(pairRaw, "=", 2)

		key := pair[0]
		value := pair[1]

		if re.MatchString(key) {
			environ[key] = value
		}
	}

	config := make(map[string]interface{})
	config[rootKey] = environ

	return config, nil
}
