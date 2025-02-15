package main

import (
	"fmt"
	"github.com/bgrewell/iso-kit"
	"github.com/bgrewell/usage"
	"os"
)

func main() {

	u := usage.NewUsage()
	help := u.AddBooleanOption("h", "help", false, "Show this help message", "optional", nil)
	verbose := u.AddBooleanOption("v", "verbose", false, "Print verbose output", "", nil)
	path := u.AddArgument(1, "iso-path", "Path to the file within the ISO to read", "")
	parsed := u.Parse()

	if !parsed {
		u.PrintError(fmt.Errorf("failed to parse arguments"))
		os.Exit(1)
	}

	if *help {
		u.PrintUsage()
		os.Exit(0)
	}

	if path == nil || *path == "" {
		u.PrintError(fmt.Errorf("location of the iso file <path> must be provided"))
		os.Exit(1)
	}

	_ = verbose

	i, err := iso.Open(*path)
	if err != nil {
		u.PrintError(err)
		os.Exit(1)
	}

	_ = i

}
