package main

import (
	"fmt"

	"github.com/z3vxo/nyx/internal/server"
)

func main() {
	if err := server.Setup(); err != nil {
		fmt.Println("failed")
	}
	return
}
