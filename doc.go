// Copyright (c) 2024, Eugene Ponizovsky, <ponizovsky@gmail.com>. All rights
// reserved. Use of this source code is governed by a MIT License that can
// be found in the LICENSE file.

/*
Package conf is an extensible solution for cascading configuration. Package conf
provides configuration processor, that can load configuration layers from
different sources and merges them into the one configuration tree. Package conf
comes with built-in configuration loaders: fileconf and envconf, and can be
extended by third-party configuration loaders. Package conf do not watch for
configuration changes, but you can implement this feature in the custom
configuration loader. You can find full example in repository.

Configuration processor can expand references, that can be specified in string
values, to configuration parameters in the same configuration tree (if you need
references to complex structures, see $ref directive).

	myapp:
		rootDir: "/myapp"
		templatesDir: "${myapp.rootDir}/templates"
		sessionsDir: "${myapp.rootDir}/sessions"

To escape expansion of references, add one more "$" symbol. For example:

	templatesDir: "$${myapp.rootDir}/templates"

Configuration processor supports $ref directive, that can be used to refer to
configuration parameters in the same configuration tree. $ref directive can take
three forms. In the first form $ref directive just try to get a value by name.
In the second form $ref directive tries to get a value by name and if no value
is found, uses default value. In the third form $ref directive tries to get a
value of a first defined configuration parameter and, if no value is found, uses
default value. Default value in second and third forms can be omitted.

	db:
		stat:
			host: { $ref: "db.generic.host" }
			port: { $ref: { name: "db.generic.port", default: 5432 } }
			dbname: "stat"
			username: { $ref: { name: "db.generic.username", default: "some_user" } }
			password:
				$ref: { firstDefined: [ "MYAPP_DB_STAT_PASS", "db.generic.password" ], default: "some_pass" }

Configuration processor can include additional configuration to main configuration
tree from external sources using $include directive. $include directive applies
before merge of configuration layers. $include directive accepts a list of
configuration locators as argument.

	db: { $include: [ "file:connectors.yml" ] }
*/
package conf
