package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	SourceDir     string
	TargetDir     string
	DeleteMissing bool
	UseChecksum   bool
}

func reorderArgs() {
	var flags []string
	var positional []string

	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
		} else {
			positional = append(positional, arg)
		}
	}

	os.Args = append([]string{os.Args[0]}, append(flags, positional...)...)
}

func Parse() *Config {
	config := &Config{}

	flag.BoolVar(&config.DeleteMissing, "delete-missing", false, "Delete files in target that don't exist in source")
	flag.BoolVar(&config.DeleteMissing, "d", false, "Delete files in target that don't exist in source (shorthand)")
	flag.BoolVar(&config.UseChecksum, "checksum", false, "Compare files using SHA256 checksum (slower but more accurate)")
	flag.BoolVar(&config.UseChecksum, "c", false, "Compare files using SHA256 checksum (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <source> <target>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "One-way file synchronization tool (source -> target)\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  <source>  Source directory path\n")
		fmt.Fprintf(os.Stderr, "  <target>  Target directory path\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -d, --delete-missing  Delete files in target that don't exist in source\n")
		fmt.Fprintf(os.Stderr, "  -c, --checksum        Compare files using SHA256 checksum (slower but more accurate)\n")
		fmt.Fprintf(os.Stderr, "  -h, --help            Show this help message\n")
	}

	reorderArgs()
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Error: source and target directories are required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	config.SourceDir = args[0]
	config.TargetDir = args[1]

	return config
}
