package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/ami-/tengo-language-tools/internal/formatter"
)

var (
	version      = "dev"
	tengoVersion = "unknown"
)

func main() {
	write := flag.Bool("w", false, "write result to source file instead of stdout")
	ver := flag.Bool("version", false, "print version and exit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: tengofmt [flags] [file...]\n\n")
		fmt.Fprintf(os.Stderr, "Formats Tengo source files. Reads from stdin if no files are given.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *ver {
		fmt.Printf("tengofmt %s (tengo %s)\n", version, tengoVersion)
		return
	}

	if flag.NArg() == 0 {
		src, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		out, err := formatter.Format(src)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Stdout.Write(out)
		return
	}

	exitCode := 0
	for _, path := range flag.Args() {
		if err := processFile(path, *write); err != nil {
			fmt.Fprintln(os.Stderr, err)
			exitCode = 1
		}
	}
	os.Exit(exitCode)
}

func processFile(path string, write bool) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	out, err := formatter.Format(src)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if write {
		return os.WriteFile(path, out, 0o644)
	}
	os.Stdout.Write(out)
	return nil
}
