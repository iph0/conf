// Copyright (c) 2024, Eugene Ponizovsky, <ponizovsky@gmail.com>. All rights
// reserved. Use of this source code is governed by a MIT License that can
// be found in the LICENSE file.

/*
Package conf is an extensible solution for cascading configuration. Package conf
provides configuration processor, that can load configuration layers from
different sources and merges them into the one configuration tree. In addition
configuration processor can expand references on configuration parameters in
string values, and process $ref and _include directives in resulting configuration
tree (see below). Package conf comes with built-in configuration loaders: fileconf
and envconf, maploader and can be extended by third-party configuration loaders.
Package conf do not watch for configuration changes, but you can implement this
feature in the custom configuration loader. You can find full example in repository.

Configuration processor can expand references to configuration parameters in string
values (if you need reference to complex structures, see $ref directive). For
example, you have a YAML file:

	myapp:
	  mediaFormats: ["images", "audio", "video"]

	  dirs:
	    rootDir: "/myapp"
	    templatesDir: "${myapp.dirs.rootDir}/templates"
	    sessionsDir: "${myapp.dirs.rootDir}/sessions"

	    mediaDirs:
	      - "${myapp.dirs.rootDir}/media/${myapp.mediaFormats.0}"
	      - "${myapp.dirs.rootDir}/media/${myapp.mediaFormats.1}"
	      - "${myapp.dirs.rootDir}/media/${myapp.mediaFormats.2}"

After processing of the file you get a map:

	"myapp": conf.M{
	  "mediaFormats": conf.A{"images", "audio", "video"},

	  "dirs": conf.M{
	    "rootDir": "/myapp",
	    "templatesDir": "/myapp/templates",
	    "sessionsDir": "/myapp/sessions",

	    "mediaDirs": conf.A{
	      "/myapp/media/images",
	      "/myapp/media/audio",
	      "/myapp/media/video",
	    },
	  },
	}

To escape expansion of references, add one more "$" symbol. For example:

	templatesDir: "$${myapp.dirs.rootDir}/templates"

After processing we will get:

	templatesDir: "${myapp.dirs.rootDir}/templates"

Package conf supports special directives in configuration layers: $ref and
$include. $ref directive tries to get a value of a configuration parameter by
his name. $ref directive can take three forms:

	$ref: <name>
	$ref: { name: <name>, default: <value> }
	$ref: { firstDefined: [ <name1>, ... ], default: <value> }

In the first form $ref directive just try to get a value by name. In the second
form $ref directive tries to get a value by name and if no value is found, uses
default value. In the third form $ref directive tries to get a value of a first
defined configuration parameter and, if no value is found, uses default value.
Default value in second and third forms can be omitted. Below is an example of
TOML file with $ref directives:

	db:
	  defaultOptions:
	    serverPrepare: true
	    expandArray: true
	    errorLevel: 2

	  connectors:
	    stat:
	      host: "stat.mydb.com"
	      port: 1234
	      dbname: "stat"
	      username: "stat"
	      password:
	        $ref: {name: "MYAPP_DB_STAT_PASSWORD", default: "stat_pass"}
	      options: {$ref: "db.defaultOptions"}

	    metrics:
	      host: "metrics.mydb.com"
	      port: 1234
	      dbname: "metrics"
	      username: "metrics"
	      password:
	        $ref: {name: "MYAPP_DB_METRICS_PASSWORD", default: "metrics_pass"}
	      options: {$ref: "db.defaultOptions"}

$include directive loads configuration layer from external sources and inserts
it to configuration tree. $include directive accepts a list of
configuration locators as argument.

	db:
	  defaultOptions:
	    serverPrepare: true
	    expandArray: true
	    errorLevel: 2

	  connectors: { $include: [ "file:connectors.yml" ] }
*/
package conf
