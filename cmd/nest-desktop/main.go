// Binary nest-desktop is the Puffin desktop derivative CLI.
package main

import (
	"os"

	"github.com/puffin/nest/internal/cli"
	_ "github.com/puffin/nest/internal/cli/desktop"
)

func main() {
	root := cli.NewRootCmd(cli.FlavorDesktop)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}