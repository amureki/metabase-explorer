package main

import (
	"os"

	"github.com/amureki/metabase-explorer/pkg/cli"
)

var version = "dev" // Will be overridden by ldflags during release builds

func main() {
	cli.Execute(os.Args[1:], version)
}
