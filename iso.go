package iso

import (
	"fmt"
	"github.com/bgrewell/iso-kit/pkg"
	"github.com/bgrewell/iso-kit/pkg/consts"
	"github.com/bgrewell/iso-kit/pkg/directory"
	"github.com/bgrewell/iso-kit/pkg/options"
	"github.com/go-logr/logr"
)

// Open initializes and returns an Image object representing the given ISO file.
//
// The function accepts variadic options to customize the opening behavior. If no
// options are provided, default values are used as specified in the Options struct.
//
// Parameters:
//   - location: The file path to the ISO image to open.
//   - opts: Optional functions that modify the opening behavior using the Options struct.
//
// Returns:
//   - Image: A pointer to an Image object if successful; nil if an error occurs.
//   - error: An error describing any issue encountered during opening, or nil on success.
//
// Example usage:
//
//	// Open ISO with default options
//	img, err := Open("image.iso")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Open ISO with progress updates
//	callback := func(currentFilename string, bytesTransferred int64, totalBytes int64, currentFileNumber int, totalFileCount int) {
//	    fmt.Printf("Processing %s: %d/%d files (%d%%)\n", currentFilename, currentFileNumber, totalFileCount, (bytesTransferred * 100)/totalBytes)
//	}
//	img, err := Open("image.iso", WithProgress(callback))
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Options:
//   - WithIsoType: Sets the ISO type to use. Currently, supports ISO9660.
//   - WithStripVersionInfo: Enables or disables stripping of version information from file names.
//   - WithRockRidgeEnabled: Enables Rock Ridge support for better handling of Unix-specific features.
//   - WithElToritoEnabled: Enables El Torito support for bootable images.
//   - WithProgress: Sets a callback function to track the opening progress.
//   - WithLogger: Specifies a custom logger to use during processing.
//   - ParseOnOpen: Determines whether the image should be parsed immediately upon opening (default is true).
//
// Note:
//
//	If ParseOnOpen is set to false, you must manually call Parse() on the Image object before accessing its contents.
func Open(location string, opts ...options.Option) (Image, error) {
	// Set default options
	options := options.Options{
		IsoType:          consts.TYPE_ISO9660,
		StripVersionInfo: true,
		RockRidgeEnabled: true,
		ElToritoEnabled:  true,
		BootFileLocation: "[BOOT]", // Default location for boot files, same as 7zip
		Logger:           logr.Discard(),
	}

	// Apply options
	for _, opt := range opts {
		opt(&options)
	}

	// Validate ISO type
	switch options.IsoType {
	case consts.TYPE_ISO9660:
		// Create the specific Image type and return it
		img := &pkg.ISO9660Image{Options: options}
		return img, img.Open(location)
	default:
		return nil, fmt.Errorf("unsupported ISO type: %d", options.IsoType)
	}
}

// Create creates a new ISO image file
func Create(location string, opts ...options.Option) (Image, error) {
	options := options.Options{
		IsoType:          consts.TYPE_ISO9660,
		StripVersionInfo: true,
		RockRidgeEnabled: true,
		ElToritoEnabled:  true,
	}
	for _, opt := range opts {
		opt(&options)
	}

	image := &pkg.ISO9660Image{Options: options}
	if err := image.Create(location); err != nil {
		return nil, fmt.Errorf("failed to create ISO: %w", err)
	}
	return image, nil
}

// Image represents an ISO image
type Image interface {
	Open(isoLocation string) error
	Create(isoLocation string) error
	Parse() error
	Parsed() bool
	Close() error
	String() string
	HasRockRidge() bool
	HasElTorito() bool
	RootDirectory() *directory.DirectoryEntry
	ExtractFiles(outputLocation string) error
	ExtractBootImages(outputLocation string) error
	Extract(outputLocation string, includeBootImages bool) error
	GetAllEntries() ([]*directory.DirectoryEntry, error)
}
