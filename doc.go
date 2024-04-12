// Copyright (c) 2024, Eugene Ponizovsky, <ponizovsky@gmail.com>. All rights
// reserved. Use of this source code is governed by a MIT License that can
// be found in the LICENSE file.

/*
Module conf is an extensible solution for cascading configuration. Module conf
provides the configuration processor, that can load configuration layers from
different sources and merges them into the one configuration tree. Module conf
comes with built-in configuration loaders: fileconf and envconf, and can be
extended by third-party configuration loaders. Module conf do not watch for
configuration changes, but you can implement this feature in the custom
configuration loader. Configuration processor in conf module supports processing
directives $include, $ref, $underlay and $overlay. See more information about
directive below.

Configuration processor can include additional configuration sections to main
configuration tree from external sources using $include directive. $include
directives applies before the process of merging of configuration layers.
$include directive accepts a list of configuration locators as a value.

	db: { $include: [ "file:connectors.yml" ] }

Configuration processor can expand references, that can be specified in string
values, to configuration parameters within the same configuration tree (if you need
references to configuration sections, see $ref directive).

	myapp:
		mediaFormats: ["images", "audio", "video"]

		rootDir: "/var/lib/myapp"
		templatesDir: "${myapp.rootDir}/templates"
		sessionsDir: "${myapp.rootDir}/sessions"
		mediaDirs:
			- "${myapp.rootDir}/media/${myapp.mediaFormats.0}"
			- "${myapp.rootDir}/media/${myapp.mediaFormats.1}"
			- "${myapp.rootDir}/media/${myapp.mediaFormats.2}"

To escape expansion of references, add one more "$" symbol. For example:

	templatesDir: "$${myapp.rootDir}/templates"

Configuration processor supports internal references to configuration parameters
within the same configuration tree using $ref directive. $ref directive can take
three forms. The first form of the $ref directive just tries to get a value by
name. The second form of the $ref directive tries to get a value by name and if
no value is found, uses default one. The third form of the $ref directive tries
to get a first defined value, if no value is found, uses  default one. Default
value in second and third forms can be omitted.

	db:
		default:
			port: "5432"
			dbname: "stat"
			username: "guest"
			password: "guest_pass"
			options:
				PrintWarn: 0
				PrintError: 0
				RaiseError: 1

		stat_master:
			host: "stat-master.mydb.com"
			port: { $ref: "db.default.port" }
			dbname: { $ref: { name: "db.default.dbname", default: "default_stat" } }
			username:
				$ref:
					firstDefined: ["MYAPP_DB_STAT_USER", "db.default.username"]
					default: "stat_writer"
			password:
				$ref:
					firstDefined: ["MYAPP_DB_STAT_PASS", "db.default.password"]
					default: "stat_writer_pass"

Configuration processor supports internal merges between configuration sections
within the same configuration tree using $underlay and $overlay directives. Both
directives accept a parameter name or a list of parameter names as a value.

$underlay directive retrieves configuration sections by parameter names, merges
them between each other in the order in which the names was specified and places
the result under the configuration section where $underlay directive was specified.

	db:
		default:
			port: "5432"
			dbname: "stat"
			username: "guest"
			password: "guest_pass"
			options:
				PrintWarn: 0
				PrintError: 0
				RaiseError: 1

		stat_master:
			$underlay: "db.default"
			host: "stat-master.mydb.com"
			username: "stat_writer"
			password: "stat_writer_pass"

		stat_slave:
			$underlay: "db.default"
			host: "stat-slave.mydb.com"
			username: "stat_reader"
			password: "stat_reader_pass"

$overlay directive retrieves configuration sections by parameter names, merges
them between each other in the order in which the names was specified and overlays
the result on the configuration section where $overlay directive was specified.

	db:
		default:
			port: "5432"
			dbname: "stat"
			username: "guest"
			password: "guest_pass"
			options:
				PrintWarn: 0
				PrintError: 0
				RaiseError: 1

		stat_master:
			$underlay: "db.default"
			host: "stat-master.mydb.com"
			username: "stat_writer"
			password: "stat_writer_pass"
			$overlay: "db.test"

		stat_slave:
			$underlay: "db.default"
			host: "stat-slave.mydb.com"
			username: "stat_reader"
			password: "stat_reader_pass"
			$overlay: "db.test"

		test:
			host: "localhost"
			port: "54322"
*/
package conf
