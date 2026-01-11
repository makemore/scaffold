package main

import (
	"os"

	"github.com/christophercochran/scaffold/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

