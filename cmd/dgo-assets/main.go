// Command dgo-assets checks the graphics asset manifest, lists it and converts
// legacy Diablo graphics containers to CLX.
package main

import (
	"os"

	"github.com/bmcszk/devilutionGo/cmd/dgo-assets/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
