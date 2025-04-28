package main

import "flag"

var logDirPath string

func init() {
	// path to the log directory
	// default is /var/log
	flag.StringVar(&logDirPath, "path", "var/log", "path to log directory")

	// parse the flags
	flag.Parse()
}
