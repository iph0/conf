# conf [![GoDoc](https://godoc.org/github.com/iph0/conf?status.svg)](https://godoc.org/github.com/iph0/conf) [![Build Status](https://travis-ci.org/iph0/conf.svg?branch=master)](https://travis-ci.org/iph0/conf) [![Go Report Card](https://goreportcard.com/badge/github.com/iph0/conf)](https://goreportcard.com/report/github.com/iph0/conf) [![Codecov](https://codecov.io/gh/iph0/conf/branch/master/graph/badge.svg)](https://codecov.io/gh/iph0/conf)

Package conf loads configuration sections from different sources and merges
them into the one configuration tree.

```go
package main

import (
  "fmt"
  "os"

  "github.com/iph0/conf"
  "github.com/iph0/conf/baseconf"
)

func init() {
  os.Setenv("GOCONF_PATH", "/etc/go")
}

func main() {
  loader := conf.NewLoader(
    baseconf.NewDriver(),
  )

  config, err := loader.Load("dirs", "db")

  if err != nil {
    fmt.Println("Failed to load configuration:", err)
    return
  }

  fmt.Printf("%v\n", config)
}
```