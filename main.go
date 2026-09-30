package main

import (
	"os"

	"github.com/yakovlef/totext/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
