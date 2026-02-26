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
	IdentityFile  string
	Port          int
	Password      string
}

func reorderArgs() {
	var flags []string
	var positional []string

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			flags = append(flags, args[i])
			// If this flag takes a value (not a boolean flag), include the next arg too
			switch args[i] {
			case "-i", "--identity", "-p", "--port", "--password":
				if i+1 < len(args) {
					i++
					flags = append(flags, args[i])
				}
			}
		} else {
			positional = append(positional, args[i])
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
	flag.StringVar(&config.IdentityFile, "identity", "", "Path to SSH private key")
	flag.StringVar(&config.IdentityFile, "i", "", "Path to SSH private key (shorthand)")
	flag.IntVar(&config.Port, "port", 22, "SSH port")
	flag.IntVar(&config.Port, "p", 22, "SSH port (shorthand)")
	flag.StringVar(&config.Password, "password", "", "SSH password (prefer key-based auth)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <source> <target>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "One-way file synchronization tool (source -> target)\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  <source>  Source directory path or [user@]host:/path\n")
		fmt.Fprintf(os.Stderr, "  <target>  Target directory path or [user@]host:/path\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -d, --delete-missing  Delete files in target that don't exist in source\n")
		fmt.Fprintf(os.Stderr, "  -c, --checksum        Compare files using SHA256 checksum (slower but more accurate)\n")
		fmt.Fprintf(os.Stderr, "  -i, --identity FILE   Path to SSH private key (default: ~/.ssh/id_ed25519, ~/.ssh/id_rsa)\n")
		fmt.Fprintf(os.Stderr, "  -p, --port PORT       SSH port (default: 22)\n")
		fmt.Fprintf(os.Stderr, "      --password PASS   SSH password (prefer key-based auth)\n")
		fmt.Fprintf(os.Stderr, "  -h, --help            Show this help message\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s /local/src /local/dst                        Local to local\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s /local/src user@host:/remote/dst             Local to remote\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s user@host:/remote/src /local/dst             Remote to local\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s user@host1:/path user@host2:/path            Remote to remote\n", os.Args[0])
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
