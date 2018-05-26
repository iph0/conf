// Copyright (c) 2018, Eugene Ponizovsky, <ponizovsky@gmail.com>. All rights
// reserved. Use of this source code is governed by a MIT License that can
// be found in the LICENSE file.

/*
Package conf loads configuration sections from different sources and merges
them into the one configuration tree.

 package main

 import (
   "fmt"
   "os"

   "github.com/iph0/conf"
   "github.com/iph0/conf/fileconf"
 )

 func init() {
   os.Setenv("GOCONF_PATH", "/etc/go")
 }

 func main() {
   loader := conf.NewLoader(
     fileconf.NewDriver(true),
     &envconf.EnvDriver{},
   )

   config, err := loader.Load(
     "file:myapp/dirs.yml",
     "file:myapp/*.json",
     "env:^MYAPP_.*",
   )

   if err != nil {
     fmt.Println("Loading failed:", err)
     return
   }

   fmt.Printf("%v\n", config)
 }

conf package can expand variables in string values. Variable names can be
absolute or relative. Relative variable names begins with "." (dot). The number
of dots depends on the nesting level of the current configuration parameter
relative to referenced configuration parameter. For example, we have a YAML file:

 myapp:
   mediaFormats: [ "images", "audio", "video" ]

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

To escape variable expansion add one more "$" symbol before variable.

 templatesDir: "$${myapp.dirs.root_dir}/templates"

After processing we will get:

 templatesDir: "${myapp.dirs.root_dir}/templates"
*/
package conf
