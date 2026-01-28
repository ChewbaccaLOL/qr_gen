package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"qr_generator/internal/config"
)

const (
	exitUsage = 2
)

func main() {
	var listVariants bool

	flag.BoolVar(&listVariants, "list-variants", false, "List available variants and exit.")
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Minimal Go CLI prototype for the QR generator.")
		fmt.Fprintln(flag.CommandLine.Output())
		fmt.Fprintln(flag.CommandLine.Output(), "Usage:")
		fmt.Fprintln(flag.CommandLine.Output(), "  qr_generator --list-variants")
		fmt.Fprintln(flag.CommandLine.Output())
		flag.PrintDefaults()
	}

	flag.Parse()

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: unable to resolve working directory")
		os.Exit(exitUsage)
	}
	configPath := filepath.Join(cwd, "variants.json")

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(exitUsage)
	}

	if listVariants {
		names := make([]string, 0, len(cfg.Variants))
		for name := range cfg.Variants {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			fmt.Println(name)
		}
		fmt.Println()
		fmt.Println("Animations:")
		for _, name := range cfg.AnimationVariants {
			fmt.Println(name)
		}
		return
	}

	fmt.Fprintln(os.Stderr, "error: only --list-variants is implemented in the Go prototype")
	os.Exit(exitUsage)
}
