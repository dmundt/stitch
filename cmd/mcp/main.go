package main

import (
	"context"
	"log"

	"github.com/dmundt/stitch/internal/mcp"
)

func main() {
	if err := mcp.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
