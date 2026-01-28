package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"qr_generator/internal/cli"
	"qr_generator/internal/config"
)

const (
	exitUsage = 2
)

func main() {
	env := cli.OSEnv{}
	cli.LoadDotenv(".env", env)

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

	args, err := cli.ParseArgs(os.Args[1:], cfg, env, os.Stdin, isTerminal(os.Stdin))
	if err != nil {
		if usageErr, ok := err.(cli.ErrUsage); ok {
			fmt.Fprintln(os.Stderr, usageErr.Usage)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(exitUsage)
	}

	if args.ListVariants {
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

	fmt.Fprintln(os.Stderr, "error: rendering is not implemented in the Go CLI yet")
	os.Exit(exitUsage)
}

func isTerminal(file *os.File) bool {
	info, err := file.Stat()
	if err != nil {
		return true
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}
