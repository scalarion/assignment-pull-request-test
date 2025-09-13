package main

import (
	"fmt"
	"os"

	"assignment-pull-request/internal/creator"
)

func main() {
	prCreator, err := creator.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := prCreator.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
