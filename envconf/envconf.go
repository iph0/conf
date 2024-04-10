// Copyright (c) 2024, Eugene Ponizovsky, <ponizovsky@gmail.com>. All rights
// reserved. Use of this source code is governed by a MIT License that can
// be found in the LICENSE file.

/*
Package envconf is configuration loader for the conf package. It loads
configuration layers from environment variables. Configuration locators for this
loader are regular expressions. Here some examples:

	env:^MYAPP_
	env:.*
*/
package envconf

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/iph0/conf/v2"
)

const errPref = "envconf"

// Loader loads configuration layers from environment variables.
type Loader struct{}

// NewLoader method creates new loader instance.
func NewLoader() *Loader {
	return &Loader{}
}

// Load method loads configuration layer from environment variables.
func (l *Loader) Load(re string) ([]any, error) {
	reObj, err := regexp.Compile(re)

	if err != nil {
		return nil, fmt.Errorf("%s: %s", errPref, err)
	}

	envs := os.Environ()
	layer := make(conf.M)

	for _, envStr := range envs {
		tokens := strings.SplitN(envStr, "=", 2)

		if reObj.MatchString(tokens[0]) {
			layer[tokens[0]] = tokens[1]
		}
	}

	return []any{layer}, nil
}
