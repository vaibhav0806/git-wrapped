// main.go
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: gh-wrapped <username> [--auto]")
		os.Exit(1)
	}
	username := os.Args[1]
	auto := len(os.Args) > 2 && os.Args[2] == "--auto"

	_ = username
	_ = auto
	fmt.Println("gh-wrapped: not yet implemented")
}
