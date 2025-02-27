package main

import (
	"os"
	
	"github.com/lugnut42/openipam/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		// Error is already logged and printed to stderr by Execute()
		// Just exit with non-zero status
		os.Exit(1)
	}
}