package main

import (
	"fmt"
	"github.com/bgrewell/iso-kit"
	"github.com/bgrewell/iso-kit/pkg/option"
	"github.com/bgrewell/iso-kit/pkg/version"
	"github.com/bgrewell/usage"
	"github.com/theckman/yacspin"
	"golang.org/x/term"
	"os"
	"time"
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
	// Initialize usage handler
	u := usage.NewUsage(
		usage.WithApplicationVersion(version.Version()),
		usage.WithApplicationBranch(version.Branch()),
		usage.WithApplicationBuildDate(version.Date()),
		usage.WithApplicationCommitHash(version.Revision()),
		usage.WithApplicationName("isoextract"),
		usage.WithApplicationDescription("isoextract is a command-line tool for extracting files and boot images from ISO9660 images, including support for Rock Ridge, Joliet, and El Torito extensions."),
	)

	// Define CLI options
	help := u.AddBooleanOption("h", "help", false, "Show this help message", "optional", nil)
	verbose := u.AddBooleanOption("v", "verbose", false, "Enable verbose (debug) logging", "", nil)
	trace := u.AddBooleanOption("vv", "trace", false, "Enable trace logging", "", nil)
	bootImages := u.AddBooleanOption("b", "boot", false, "Extract boot images (El Torito)", "", nil)
	rockRidge := u.AddBooleanOption("rr", "rockridge", true, "Enable Rock Ridge support", "", nil)
	enhancedVol := u.AddBooleanOption("eh", "enhanced", true, "Use Enhanced Volume Descriptors", "", nil)
	stripVer := u.AddBooleanOption("s", "strip", true, "Strip version info from filenames", "", nil)

	// Output directories
	outputDir := u.AddStringOption("o", "output", "./extracted", "Output directory for extracted files", "", nil)
	bootDir := u.AddStringOption("bd", "bootdir", "[BOOT]", "Output directory for boot images", "", nil)

	// ISO file path argument
	isoPath := u.AddArgument(1, "iso-path", "Path to the ISO file", "")

	// Parse arguments
	parsed := u.Parse()
	if !parsed {
		u.PrintError(fmt.Errorf("failed to parse arguments"))
		os.Exit(1)
	}

	// Handle help flag
	if *help {
		u.PrintUsage()
		os.Exit(0)
	}

	// Ensure an ISO path was provided
	if isoPath == nil || *isoPath == "" {
		u.PrintError(fmt.Errorf("path to the ISO file must be provided"))
		os.Exit(1)
	}

	// Setup logging level
	if *trace {
		fmt.Println("Trace logging enabled")
	} else if *verbose {
		fmt.Println("Verbose logging enabled")
	}

	// Setup callback for progress updates
	spinner, err := InitializeSpinner()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize spinner: %v\n", err)
		fmt.Fprintf(os.Stderr, "Progress updates will be disabled.\n")
	}

	// Create progress callback
	progressCallback := CreateProgressCallback(spinner)

	// Open the ISO image with the specified flags
	img, err := iso.Open(
		*isoPath,
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
