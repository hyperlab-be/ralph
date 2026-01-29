package main

import (
	"os"

	"github.com/hyperlab/ralph/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
