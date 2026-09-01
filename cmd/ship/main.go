package main

import (
	"os"

	"github.com/Dendro-X0/Orbit/internal/cli"
	_ "github.com/Dendro-X0/Orbit/internal/providers/cloudflare"
)

func main() {
	if err := cli.NewRoot().Execute(); err != nil {
		os.Exit(1)
	}
}
