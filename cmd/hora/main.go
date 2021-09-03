package main

import (
	"os"

	"github.com/deislabs/hora/cmd/hora/cmd"
)

func main() {
	if err := cmd.Root.Execute(); err != nil {
		os.Exit(1)
	}
	//cmd.Test("localhost:5000/net-monitor@sha256:05361724f0daf5054ac05150ba980ab69c704290cffe013cfa518663d0e2c956")
}
