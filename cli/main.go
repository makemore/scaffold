package main

import (
	"os"

	"github.com/makemore/scaffold/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

