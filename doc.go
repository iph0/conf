// Copyright (c) 2018, Eugene Ponizovsky, <ponizovsky@gmail.com>. All rights
// reserved. Use of this source code is governed by a MIT License that can
// be found in the LICENSE file.

/*
Package conf is an extensible solution for application configuration. It loads
configuration sections from different sources and merges them into the one
configuration tree. Can be extended by third-party configuration providers.
Package conf comes with built-in providers: fileconf and envconf.

 package main

 import (
   "fmt"
   "os"

   "github.com/iph0/conf"
   "github.com/iph0/conf/envconf"
   "github.com/iph0/conf/fileconf"
 )

 func init() {
   os.Setenv("GOCONF_PATH", "/etc/go")
 }

 func main() {
   loader := conf.NewLoader(
     &envconf.EnvProvider{},
     fileconf.NewProvider(),
   )

   config, err := loader.Load(
     "env:^MYAPP_.*",
     "file:myapp/dirs.yml",
     "file:myapp/*.json",
   )

   if err != nil {
     fmt.Println("Loading failed:", err)
     return
   }

   fmt.Printf("%v\n", config)
 }

Package conf can expand variables in string values (if you need alias for
complex structures see @var directive). Variable names can be absolute or
relative. Relative variable names begins with "." (dot). The number of dots
depends on the nesting level of the current configuration parameter relative to
referenced configuration parameter. For example, we have a YAML file:

 myapp:
   mediaFormats: ["images", "audio", "video"]

   dirs:
     rootDir: "/myapp"
     templatesDir: "${myapp.dirs.root_dir}/templates"
     sessionsDir: "${.root_dir}/sessions"
     mediaDirs:
       - "${..root_dir}/media/${myapp.media_formats.0}"
       - "${..root_dir}/media/${myapp.media_formats.1}"
       - "${..root_dir}/media/${myapp.media_formats.2}"

After processing of the file we will get:

 "myapp": map[string]interface{}{
   "mediaFormats": []interface{}{"images", "audio", "video"},

   "dirs": map[string]interface{}{
     "rootDir":      "/myapp",
     "templatesDir": "/myapp/templates",
     "sessionsDir": "/myapp/sessions",
     "mediaDirs": []interface{}{
       "/myapp/media/images",
       "/myapp/media/audio",
       "/myapp/media/video",
     },
   },
 }

To escape variable expansion add one more "$" symbol before variable name.

 templatesDir: "$${myapp.dirs.rootDir}/templates"

After processing we will get:

 templatesDir: "${myapp.dirs.rootDir}/templates"

Package conf support directives: @var and @include. @var directive assigns
configuration parameter value to another configuration parameter. Argument of
the @var directive is a variabale name, absolute or relative.

 myapp:
   db:
     defaultOptions:
       PrintWarn:  0
       PrintError: 0
       RaiseError: 1

     connectors:
       stat:
         host:     "stat.mydb.com"
         port:     "1234"
         dbname:   "stat"
         username: "stat_writer"
         password: "stat_writer_pass"
         options:  {"@var": "myapp.db.defaultOptions"}

       metrics:
         host:     "metrics.mydb.com"
         port:     "1234"
         dbname:   "metrics"
         username: "metrics_writer"
         password: "metrics_writer_pass"
         options:  {"@var": "...defaultOptions"}

@include directive loads configuration section from external sources and assigns
it to specified configuration parameter. Argument of the @include directive is a
list of source patterns.

 myapp:
   db:
     defaultOptions:
       PrintWarn:  0
       PrintError: 0
       RaiseError: 1

     connectors: {"@include": ["conf.d/*.yml", "conf.d/*.json"]}
*/
package conf
