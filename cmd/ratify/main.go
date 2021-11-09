package main

import (
	"os"

	"github.com/deislabs/ratify/cmd/ratify/cmd"
	_ "github.com/deislabs/ratify/pkg/referrerstore/oras"
	_ "github.com/deislabs/ratify/pkg/verifier/notaryv2"
)

func main() {
	if err := cmd.Root.Execute(); err != nil {
		os.Exit(1)
	}
}
