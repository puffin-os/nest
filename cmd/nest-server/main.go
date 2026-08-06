// Binary nest-server is the Puffin server derivative CLI.
package main

import (
	"os"

	"github.com/puffin/nest/internal/cli"
)

func main() {
	root := cli.NewRootCmd(cli.FlavorServer)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}