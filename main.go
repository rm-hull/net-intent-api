package main

import (
	"os"

	"github.com/rm-hull/net-intent-api/internal"
)

func main() {
	if err := internal.Server(); err != nil {
		os.Exit(-1)
	}
}
