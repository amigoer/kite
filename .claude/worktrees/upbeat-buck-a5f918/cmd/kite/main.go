package main

import (
	"fmt"
	"os"

	"github.com/amigoer/kite/internal/cli"
)

var version = "dev"

func main() {
	if err := cli.NewRootCommand(version).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
