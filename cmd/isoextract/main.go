package main

import (
	"flag"
	"fmt"
	"github.com/bgrewell/iso-kit"
	"github.com/bgrewell/iso-kit/pkg/logging"
	"os"
)

func main() {
	// Logging level flags
	debug := flag.Bool("v", false, "Enable verbose (debug) logging")
	trace := flag.Bool("vv", false, "Enable trace logging")

	// Extraction options
	bootImages := flag.Bool("boot", false, "Extract boot images (El Torito)")
	rockRidge := flag.Bool("rockridge", true, "Enable Rock Ridge support")
	enhancedVol := flag.Bool("enhanced", true, "Use Enhanced Volume Descriptors")
	stripVer := flag.Bool("strip", true, "Strip version info from filenames")

	// Output directory
	outputDir := flag.String("o", "./extracted", "Output directory for extracted files")
	bootDir := flag.String("bootdir", "[BOOT]", "Output directory for boot images")

	// Parse flags
	flag.Parse()

	// Configure logging
	if *trace {
		level := "trace"
		logging.InitLogger(&level)
	} else if *debug {
		level := "debug"
		logging.InitLogger(&level)
	}

	// Ensure we have an ISO path
	if flag.NArg() < 1 {
		fmt.Println("Usage: isoextract [options] <path-to-iso>")
		fmt.Println("  -v               Enable verbose (debug) logging")
		fmt.Println("  -vv              Enable trace logging")
		fmt.Println("  -boot            Extract boot images (El Torito)")
		fmt.Println("  -rockridge       Enable Rock Ridge support (default: true)")
		fmt.Println("  -enhanced        Use Enhanced Volume Descriptors (default: true)")
		fmt.Println("  -strip           Strip version info from filenames (default: true)")
		fmt.Println("  -o <directory>   Output directory (default './extracted')")
		fmt.Println("  -bootdir <dir>   Output directory for boot images (default './extracted/boot')")
		os.Exit(1)
	}

	// Grab the ISO path from arguments
	isoPath := flag.Arg(0)

	// Open the ISO image with the specified flags
	img, err := iso.Open(
		isoPath,
		iso.WithEltoritoEnabled(*bootImages),
		iso.WithRockRidgeEnabled(*rockRidge),
		iso.WithParseOnOpen(*enhancedVol),
		iso.WithBootFileLocation(*bootDir),
		iso.WithPreferEnhancedVD(*enhancedVol),
		iso.WithStripVersionInfo(*stripVer),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open ISO: %v\n", err)
		os.Exit(1)
	}
	defer img.Close()

	// Extract the contents
	err = img.Extract(*outputDir, *bootImages)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to extract image: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Extraction completed successfully to '%s'.\n", *outputDir)
}
