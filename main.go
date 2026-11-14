// Command compa: A troubleshooting companion for support cases: a chat with a model that reads a case's artifacts through sandboxed, deterministic tools and loads domain experts on demand.
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	fmt.Println("Hello from compa!")
	return nil
}
