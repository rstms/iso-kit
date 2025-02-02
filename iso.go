package iso

import (
	"fmt"
	"github.com/bgrewell/iso-kit/pkg"
	"github.com/bgrewell/iso-kit/pkg/consts"
	"github.com/bgrewell/iso-kit/pkg/directory"
	"github.com/bgrewell/iso-kit/pkg/options"
	"github.com/go-logr/logr"
)

// Open opens an existing ISO image file
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
