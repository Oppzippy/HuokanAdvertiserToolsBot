package main

import (
	"log"
	"os"
)

var logger = log.New(os.Stdout, "", log.Flags())
var errLogger = log.New(os.Stderr, "", log.Flags())
