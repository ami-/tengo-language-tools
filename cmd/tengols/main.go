package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ami-/tengo-language-tools/internal/lsp"
)

var (
	version      = "dev"
	tengoVersion = "unknown"
)

func main() {
	ver := flag.Bool("version", false, "print version and exit")
	flag.Bool("stdio", false, "use stdio transport (default, accepted for LSP client compatibility)")
	flag.Parse()

	if *ver {
		fmt.Printf("tengols %s (tengo %s)\n", version, tengoVersion)
		return
	}

	if err := lsp.Serve(os.Stdin, os.Stdout, version); err != nil {
		fmt.Fprintln(os.Stderr, "tengols:", err)
		os.Exit(1)
	}
}
