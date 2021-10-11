package main

import (
	"os"

	"github.com/deislabs/hora/cmd/hora/cmd"
	_ "github.com/deislabs/hora/pkg/referrerstore/oras"
	_ "github.com/deislabs/hora/pkg/verifier/notaryv2"
)

func main() {
	if err := cmd.Root.Execute(); err != nil {
		os.Exit(1)
	}
}
