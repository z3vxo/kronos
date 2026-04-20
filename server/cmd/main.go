package main

import (
	"log"
	"os"

	"github.com/z3vxo/kronos/internal/kronos"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "setup" {
		if err := kronos.SetupConfig(); err != nil {
			log.Fatalf("Setup failed: %v", err)
		}
		os.Exit(0)
	}

	if err := kronos.Run(); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}
}
