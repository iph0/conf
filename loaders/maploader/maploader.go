// Copyright (c) 2024, Eugene Ponizovsky, <ponizovsky@gmail.com>. All rights
// reserved. Use of this source code is governed by a MIT License that can
// be found in the LICENSE file.

/*
Package maploader is configuration loader for the conf package. It loads
configuration layers from a map. Configuration locators for this loader are just
keys of the map. For example:

	map:foo
	map:bar
*/

package maploader

import "github.com/iph0/conf/v2"

// Loader loads configuration layers from a map.
type Loader struct {
	m conf.M
}

// NewLoader method creates new loader instance.
func NewLoader(m conf.M) conf.Loader {
	return &Loader{
		m: m,
	}
}

// Load method loads configuration layer from a map.
func (l *Loader) Load(loc *conf.Locator) (any, error) {
	return l.m[loc.Value], nil
}
