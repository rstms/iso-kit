package options

import (
	"github.com/bgrewell/iso-kit/pkg/consts"
	"github.com/go-logr/logr"
)

// ProgressCallback defines the signature for progress update functions.
type ProgressCallback func(
	currentFilename string,
	bytesTransferred int64,
	totalBytes int64,
	currentFileNumber int,
	totalFileCount int,
)

// Options represents the options for opening an ISO image
type Options struct {
	IsoType          consts.ISOType
	ParseOnOpen      bool
	StripVersionInfo bool
	RockRidgeEnabled bool
	ElToritoEnabled  bool
	BootFileLocation string
	PreferEnhancedVD bool
	Logger           logr.Logger
	ProgressCallback ProgressCallback
}

// Option represents a function that modifies the Options
type Option func(*Options)

// WithProgress sets a progress callback function that will be called with progress updates.
// Parameters:
// - currentFilename: The name of the file currently being processed.
// - bytesTransferred: The number of bytes transferred so far for the current file.
// - totalBytes: The total number of bytes to be transferred for the current file.
// - currentFileNumber: The index of the current file being processed.
// - totalFileCount: The total number of files to be processed.
func WithProgress(callback ProgressCallback) Option {
	return func(o *Options) {
		o.ProgressCallback = callback
	}
}

// WithIsoType sets the ISO type for the image. Currently only ISO9660 is supported.
func WithIsoType(isoType consts.ISOType) Option {
	return func(o *Options) {
		o.IsoType = isoType
	}
}

// WithStripVersionInfo sets whether to strip version information from the ISO9660 file names
func WithStripVersionInfo(enabled bool) Option {
	return func(o *Options) {
		o.StripVersionInfo = enabled
	}
}

// WithRockRidgeEnabled sets whether to enable Rock Ridge extensions
func WithRockRidgeEnabled(enabled bool) Option {
	return func(o *Options) {
		o.RockRidgeEnabled = enabled
	}
}

// WithEltoritoEnabled sets whether to enable El Torito boot record support
func WithEltoritoEnabled(enabled bool) Option {
	return func(o *Options) {
		o.ElToritoEnabled = enabled
	}
}

// WithBootFileLocation sets the location to extract any boot files
func WithBootFileLocation(location string) Option {
	return func(o *Options) {
		o.BootFileLocation = location
	}
}

// WithLogger sets the Logger for the ISO image
func WithLogger(logger logr.Logger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}

// WithParseOnOpen sets whether to parse the ISO image when opening. If set to false then the image will need to be
// manually parsed before accessing the contents.
func WithParseOnOpen(parseOnOpen bool) Option {
	return func(o *Options) {
		o.ParseOnOpen = parseOnOpen
	}
}

func WithPreferEnhancedVD(preferEnhancedVD bool) Option {
	return func(o *Options) {
		o.PreferEnhancedVD = preferEnhancedVD
	}
}
