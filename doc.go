// Copyright (c) 2018, Eugene Ponizovsky, <ponizovsky@gmail.com>. All rights
// reserved. Use of this source code is governed by a MIT License that can
// be found in the LICENSE file.

/*
Package conf is an extensible solution for cascading configuration. Package conf
provides configuration processor, that can load configuration layers from
different sources and merges them into the one configuration tree. In addition
configuration processor can expand references on configuration parameters in
string values, and process _var and _include directives in resulting configuration
tree (see below). Package conf comes with built-in configuration loaders: fileconf
and envconf, and can be extended by third-party configuration loaders. Package
conf do not watch for configuration changes, but you can implement this feature
in the custom configuration loader. You can find full example in repository.

Configuration processor can expand references on configuration parameters in
string values (if you need reference on complex structures see _var directive).
Reference names can be absolute or relative. Relative reference names begins
with "." (dot). The section, in which a value of relative reference will be
searched, determines by number of dots in reference name. For example, we have
a YAML file:

 myapp:
   mediaFormats: ["images", "audio", "video"]

   dirs:
     rootDir: "/myapp"
     templatesDir: "${myapp.dirs.rootDir}/templates"
     sessionsDir: "${.rootDir}/sessions"

     mediaDirs:
       - "${..rootDir}/media/${myapp.mediaFormats.0}"
       - "${..rootDir}/media/${myapp.mediaFormats.1}"
       - "${..rootDir}/media/${myapp.mediaFormats.2}"

After processing of the file we will get a map:

 "myapp": map[string]interface{}{
   "mediaFormats": []interface{}{"images", "audio", "video"},

   "dirs": map[string]interface{}{
     "rootDir": "/myapp",
     "templatesDir": "/myapp/templates",
     "sessionsDir": "/myapp/sessions",

     "mediaDirs": []interface{}{
       "/myapp/media/images",
       "/myapp/media/audio",
       "/myapp/media/video",
     },
   },
 }
To escape expansion of reference, add one more "$" symbol before reference name.

 templatesDir: "$${myapp.dirs.rootDir}/templates"

After processing we will get:

 templatesDir: "${myapp.dirs.rootDir}/templates"

Package conf supports special directives in configuration layers: _var and
_include. _var directive retrives a value by reference on configuration parameter
and assigns this value to another configuration parameter. _var directive can
take three forms:

 _var: <name>
 _var: {_name: <name>, _default: <value>}
 _var: {_firstDefined: [<name1>, ...], _default: <value>}

In the first form _var directive just assings a value retrived by reference.
In the second form _var directive tries to retrive a value by reference and, if
no value retrived, assigns default value. And in the third form _var directive
tries to retrive a value from the first defined reference and, if no value
retrived, assigns default value. Default value in second and third forms can be
omitted. Reference names in _var directive can be relative or absolute.

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
       username: "stat_writer"
       password:
        _var:
          _name: "MYAPP_DB_STAT_PASSWORD"
          _default: "stat_writer_pass"
       options: {_var: "myapp.db.defaultOptions"}

     metrics:
       host: "metrics.mydb.com"
       port: 1234
       dbname: "metrics"
       username: "metrics_writer"
       password: "metrics_writer_pass"
       password:
        _var:
          _firstDefined:
            - "TEST_DB_METRICS_PASSWORD"
            - "MYAPP_DB_METRICS_PASSWORD"
          _default: "metrics_writer_pass"
       options: {_var: "...defaultOptions"}

_include directive loads configuration layer from external sources and inserts
it to configuration tree. _include directive accepts as argument a list of
configuration locators.

 db:
   defaultOptions:
     serverPrepare: true
     expandArray: true
     errorLevel: 2

   connectors: {_include: ["file:connectors.yml"]}
*/
package conf
