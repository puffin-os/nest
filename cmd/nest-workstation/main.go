// Binary nest-workstation is the Puffin workstation derivative CLI.
package main

import (
	"os"

	"github.com/puffin/nest/internal/cli"
	_ "github.com/puffin/nest/internal/cli/workstation"
)

func main() {
	root := cli.NewRootCmd(cli.FlavorWorkstation)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}