# conf

[![GoDoc](https://godoc.org/github.com/iph0/conf?status.svg)](https://godoc.org/github.com/iph0/conf/v2) [![Go Report Card](https://goreportcard.com/badge/github.com/iph0/conf)](https://goreportcard.com/report/github.com/iph0/conf)

Module conf is an extensible solution for cascading configuration. Module conf
provides the configuration processor, that can load configuration layers from
different sources and merges them into the one configuration tree. Module conf
comes with built-in configuration loaders fileconf and envconf, and can be
extended by third-party configuration loaders. Module conf do not watch for
configuration changes, but you can implement this feature in the custom
configuration loader. Configuration processor in conf module supports processing
directives $include, $ref, $underlay and $overlay. See more information about
directives in documentation.

See full documentation on [GoDoc](https://godoc.org/github.com/iph0/conf/v2) for
more information.
