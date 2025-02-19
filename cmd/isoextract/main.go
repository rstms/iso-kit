package main

import (
	"flag"
	"fmt"
	"github.com/bgrewell/iso-kit"
	"github.com/bgrewell/iso-kit/pkg/option"
	"github.com/theckman/yacspin"
	"golang.org/x/term"
	"os"
	"time"
)

var (
	version = "dev"
)

// truncateString truncates the input string to the specified max length.
// If truncation occurs, it prepends "..." to indicate the string has been shortened.
func truncateString(input string, maxLength int) string {
	if len(input) <= maxLength {
		return input
	}
	if maxLength <= 3 {
		return input[len(input)-maxLength:]
	}
	return "..." + input[len(input)-(maxLength-3):]
}

// CreateProgressCallback returns a ProgressCallback that updates the spinner's message.
func CreateProgressCallback(spinner *yacspin.Spinner) option.ExtractionProgressCallback {
	return func(
		currentFilename string,
		bytesTransferred int64,
		totalBytes int64,
		currentFileNumber int,
		totalFileCount int,
	) {
		// Fetch terminal width
		width, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			width = 80 // Default width
		}

		// Define fixed parts of the message
		fixedPart := fmt.Sprintf(" [%d/%d] ", currentFileNumber, totalFileCount)
		suffixPart := fmt.Sprintf(" - %.2f%%", float64(bytesTransferred)/float64(totalBytes)*100)

		// Calculate available space for the filename
		availableSpace := width - len(fixedPart) - len(suffixPart) - 6
		if availableSpace < 10 { // Minimum space to display meaningful filename
			availableSpace = 10
		}

		// Truncate the filename if necessary
		adjustedFilename := truncateString(currentFilename, availableSpace)

		percent := float64(bytesTransferred) / float64(totalBytes) * 100
		message := fmt.Sprintf(" [%d/%d] %s - %.2f%%",
			currentFileNumber, totalFileCount, adjustedFilename, percent)

		// Update spinner's suffix (message)
		spinner.Message(message)
	}
}

// InitializeSpinner sets up and starts the yacspin spinner.
func InitializeSpinner() (*yacspin.Spinner, error) {
	// Define spinner options
	settings := yacspin.Config{
		Frequency:         100 * time.Millisecond,
		ShowCursor:        false,
		SpinnerAtEnd:      false,
		CharSet:           yacspin.CharSets[14],
		Colors:            []string{"fgHiCyan"},
		StopColors:        []string{"fgHiGreen"},
		StopFailColors:    []string{"fgHiRed"},
		StopFailCharacter: "✗",
		StopCharacter:     "✓",
	}

	// Create a new spinner
	spinner, err := yacspin.New(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create spinner: %w", err)
	}

	// Start the spinner
	if err := spinner.Start(); err != nil {
		return nil, fmt.Errorf("failed to start spinner: %w", err)
	}

	return spinner, nil
}

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

	_ = debug
	_ = trace

	// Setup callback for progress updates
	spinner, err := InitializeSpinner()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize spinner: %v\n", err)
		fmt.Fprintf(os.Stderr, "Progress updates will be disabled.\n")
	}

	// Create progress callback
	progressCallback := CreateProgressCallback(spinner)

	// Ensure we have an ISO path
	if flag.NArg() < 1 {
		fmt.Println("isoextract v" + version)
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
		option.WithElToritoEnabled(*bootImages),
		option.WithRockRidgeEnabled(*rockRidge),
		option.WithParseOnOpen(*enhancedVol),
		option.WithBootFileExtractLocation(*bootDir),
		option.WithPreferJoliet(*enhancedVol),
		option.WithStripVersionInfo(*stripVer),
		option.WithExtractionProgress(progressCallback),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open ISO: %v\n", err)
		os.Exit(1)
	}
	defer img.Close()

	// Extract the contents
	running := true
	go func() {
		err = img.Extract(*outputDir)
		if err != nil {
			spinner.StopFailMessage(fmt.Sprintf("Failed to extract image: %v", err))
			spinner.StopFail()
			os.Exit(1)
		}
		running = false
		spinner.StopMessage(fmt.Sprintf(" All files extracted successfully to %s!", *outputDir))
		spinner.Stop()
	}()

	// Wait for extraction to complete
	for running {
		time.Sleep(10 * time.Millisecond)
	}

	//fmt.Printf("Extraction completed successfully to '%s'.\n", *outputDir)
}
