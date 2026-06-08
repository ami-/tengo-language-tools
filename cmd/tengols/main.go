package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	version      = "dev"
	tengoVersion = "unknown"
)

func main() {
	ver := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *ver {
		fmt.Printf("tengols %s (tengo %s)\n", version, tengoVersion)
		return
	}

	// placeholder — LSP server to be implemented
	fmt.Fprintln(os.Stderr, "tengols: not yet implemented")
	os.Exit(1)
}
