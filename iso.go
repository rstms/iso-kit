package iso

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/bgrewell/iso-kit/pkg/consts"
	. "github.com/bgrewell/iso-kit/pkg/descriptor"
	"github.com/bgrewell/iso-kit/pkg/directory"
	"github.com/bgrewell/iso-kit/pkg/eltorito"
	"github.com/bgrewell/iso-kit/pkg/logging"
	. "github.com/bgrewell/iso-kit/pkg/path"
	. "github.com/bgrewell/iso-kit/pkg/systemarea"
	"github.com/go-logr/logr"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// ISOType represents the type of ISO image
type ISOType int

const (
	ISO9660 ISOType = iota
)

// Options represents the options for opening an ISO image
type Options struct {
	isoType          ISOType
	parseOnOpen      bool
	stripVersionInfo bool
	rockRidgeEnabled bool
	eltoritoEnabled  bool
	bootFileLocation string
	preferEnhancedVD bool
	logger           logr.Logger
}

// Option represents a function that modifies the Options
type Option func(*Options)

// WithIsoType sets the ISO type for the image. Currently only ISO9660 is supported.
func WithIsoType(isoType ISOType) Option {
	return func(o *Options) {
		o.isoType = isoType
	}
}

// WithStripVersionInfo sets whether to strip version information from the ISO9660 file names
func WithStripVersionInfo(enabled bool) Option {
	return func(o *Options) {
		o.stripVersionInfo = enabled
	}
}

// WithRockRidgeEnabled sets whether to enable Rock Ridge extensions
func WithRockRidgeEnabled(enabled bool) Option {
	return func(o *Options) {
		o.rockRidgeEnabled = enabled
	}
}

// WithEltoritoEnabled sets whether to enable El Torito boot record support
func WithEltoritoEnabled(enabled bool) Option {
	return func(o *Options) {
		o.eltoritoEnabled = enabled
	}
}

// WithBootFileLocation sets the location to extract any boot files
func WithBootFileLocation(location string) Option {
	return func(o *Options) {
		o.bootFileLocation = location
	}
}

// WithLogger sets the logger for the ISO image
func WithLogger(logger logr.Logger) Option {
	return func(o *Options) {
		o.logger = logger
	}
}

// WithParseOnOpen sets whether to parse the ISO image when opening. If set to false then the image will need to be
// manually parsed before accessing the contents.
func WithParseOnOpen(parseOnOpen bool) Option {
	return func(o *Options) {
		o.parseOnOpen = parseOnOpen
	}
}

func WithPreferEnhancedVD(preferEnhancedVD bool) Option {
	return func(o *Options) {
		o.preferEnhancedVD = preferEnhancedVD
	}
}

// Open opens an existing ISO image file
func Open(location string, opts ...Option) (Image, error) {
	// Set default options
	options := Options{
		isoType:          ISO9660,
		stripVersionInfo: true,
		rockRidgeEnabled: true,
		eltoritoEnabled:  true,
		bootFileLocation: "[BOOT]", // Default location for boot files, same as 7zip
		logger:           logr.Discard(),
	}

	// Apply options
	for _, opt := range opts {
		opt(&options)
	}

	// Validate ISO type
	switch options.isoType {
	case ISO9660:
		// Create the specific Image type and return it
		img := &ISO9660Image{options: options}
		return img, img.Open(location)
	default:
		return nil, fmt.Errorf("unsupported ISO type: %d", options.isoType)
	}
}

// Create creates a new ISO image file
func Create(location string, opts ...Option) (Image, error) {
	options := Options{
		isoType:          ISO9660,
		stripVersionInfo: true,
		rockRidgeEnabled: true,
		eltoritoEnabled:  true,
	}
	for _, opt := range opts {
		opt(&options)
	}

	image := &ISO9660Image{options: options}
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

// ISO9660Image represents an ISO 9660 image
type ISO9660Image struct {
	SystemArea                     SystemArea
	PrimaryVolumeDescriptor        *PrimaryVolumeDescriptor
	SupplementaryVolumeDescriptors []*SupplementaryVolumeDescriptor
	BootRecordVolumeDescriptor     *BootRecordVolumeDescriptor
	eltorito                       *eltorito.ElTorito
	isoFile                        *os.File
	rootDirectory                  *directory.DirectoryEntry
	options                        Options
	logger                         logr.Logger
	parsed                         bool
}

// Open opens an existing ISO 9660 image file
func (i *ISO9660Image) Open(isoLocation string) (err error) {

	// Pull the logger out of the options
	i.logger = i.options.logger

	// Read the ISO 9660 image from the specified location
	i.isoFile, err = os.Open(isoLocation)
	if err != nil {
		return err
	}

	// Parse the iso if requested
	if i.options.parseOnOpen {
		err = i.Parse()
		if err != nil {
			return err
		}
	}

	return nil
}

// Create creates a new ISO 9660 image file
func (i *ISO9660Image) Create(isoLocation string) (err error) {
	return errors.New("ISO 9660 image creation is not yet implemented")
}

// Close closes the ISO 9660 image file
func (i *ISO9660Image) Close() error {
	return i.isoFile.Close()
}

// Parse parses the structures within the ISO 9660 image
func (i *ISO9660Image) Parse() (err error) {

	// Ensure that the iso file is open
	if i.isoFile == nil {
		return errors.New("iso file is not open")
	}

	// SystemArea isn't used by iso9660 but is used for other things so while it isn't parsed it is sliced out and exposed
	saEnd := 16 * consts.ISO9660_SECTOR_SIZE
	sa := make([]byte, saEnd)
	_, err = i.isoFile.ReadAt(sa, 0)
	if err != nil {
		return fmt.Errorf("failed to read system area: %w", err)
	}
	i.SystemArea = SystemArea(sa)

	// Handle processing the volume descriptors
	done := false
	totalLength, _ := i.isoFile.Seek(0, io.SeekEnd)
	for idx := int64(saEnd); idx < totalLength; idx += consts.ISO9660_SECTOR_SIZE {
		vdBytes := make([]byte, consts.ISO9660_SECTOR_SIZE)
		_, err = i.isoFile.ReadAt(vdBytes, idx)
		if err != nil {
			return fmt.Errorf("failed to read volume descriptor at offset %d: %w", idx, err)
		}

		// Parse the volume descriptor
		vd, err := ParseVolumeDescriptor(vdBytes, i.logger)
		if err != nil {
			return fmt.Errorf("failed to parse volume descriptor at offset %d: %w", idx, err)
		}

		switch vd.Type() {
		case VolumeDescriptorPrimary:
			i.logger.V(logging.DEBUG).Info("Processing primary volume descriptor", "idx", idx)
			pvd, err := ParsePrimaryVolumeDescriptor(vd, i.isoFile, i.logger)
			if err != nil {
				return fmt.Errorf("failed to parse primary volume descriptor: %w", err)
			}

			err = parsePathTable(i.isoFile, pvd, i.logger)
			if err != nil {
				return fmt.Errorf("failed to parse path table: %w", err)
			}

			i.PrimaryVolumeDescriptor = pvd
			i.rootDirectory = pvd.RootDirectoryEntry
			i.logger.V(logging.DEBUG).Info("Processing complete for primary volume descriptor", "idx", idx)
		case VolumeDescriptorSupplementary:
			i.logger.V(logging.DEBUG).Info("Processing supplementary volume descriptor", "idx", idx)
			svd, err := ParseSupplementaryVolumeDescriptor(vd, i.isoFile, i.logger)
			if err != nil {
				return fmt.Errorf("failed to parse supplementary volume descriptor: %w", err)
			}
			if i.SupplementaryVolumeDescriptors == nil {
				i.SupplementaryVolumeDescriptors = make([]*SupplementaryVolumeDescriptor, 0)
				if i.options.preferEnhancedVD && svd.IsJoliet() {
					i.rootDirectory = svd.RootDirectoryEntry
				}
			}
			i.SupplementaryVolumeDescriptors = append(i.SupplementaryVolumeDescriptors, svd)
			i.logger.V(logging.DEBUG).Info("Processing complete for supplementary volume descriptor", "idx", idx)
		case VolumeDescriptorBootRecord:
			i.logger.V(logging.DEBUG).Info("Processing boot record volume descriptor", "idx", idx)
			i.BootRecordVolumeDescriptor, err = ParseBootRecordVolumeDescriptor(vd, i.logger)
			if err != nil {
				return fmt.Errorf("failed to parse boot record volume descriptor: %w", err)
			}
			if isElTorito(i.BootRecordVolumeDescriptor.BootSystemIdentifier) && i.options.eltoritoEnabled {
				i.logger.V(logging.DEBUG).Info("Processing El Torito boot record volume descriptor", "id", i.BootRecordVolumeDescriptor.BootSystemIdentifier)
				catalogPointer := binary.LittleEndian.Uint32(i.BootRecordVolumeDescriptor.BootSystemUse[0:4])
				catalogOffset := int64(catalogPointer) * int64(consts.ISO9660_SECTOR_SIZE)
				catalogBytes := make([]byte, consts.ISO9660_SECTOR_SIZE)
				_, err = i.isoFile.ReadAt(catalogBytes, catalogOffset)
				if err != nil {
					return fmt.Errorf("failed to read El Torito catalog at offset %d: %w", catalogOffset, err)
				}
				i.eltorito = &eltorito.ElTorito{}
				err = i.eltorito.UnmarshalBinary(catalogBytes)
				if err != nil {
					return fmt.Errorf("failed to unmarshal El Torito catalog: %w", err)
				}
			}
			i.logger.V(logging.DEBUG).Info("Processing complete for boot record volume descriptor", "idx", idx)
		case VolumeDescriptorPartition:
			i.logger.Error(fmt.Errorf("Volume Descriptor Partition parser not implemented"), "The volume descriptor partition type is not yet supported")
		case VolumeDescriptorSetTerminator:
			i.logger.V(logging.DEBUG).Info("Processing volume descriptor set terminator", "idx", idx)
			done = true
		default:
			i.logger.Error(nil, "WARNING: Unknown volume descriptor type", "type", vd.Type())
		}

		if done {
			break
		}
	}

	i.logger.V(logging.DEBUG).Info("Finished parsing ISO 9660 image")
	i.parsed = true

	return nil
}

// Parsed returns whether the ISO 9660 image has been parsed
func (i *ISO9660Image) Parsed() bool {
	return i.parsed
}

// String returns a string representation of the ISO 9660 image data
func (i *ISO9660Image) String() string {
	// TODO: Make function to print out the ISO 9660 information
	return fmt.Sprintf("ISO 9660 Image: %s", i.isoFile.Name())
}

// RootDirectory returns the root directory of the ISO 9660 image
func (i *ISO9660Image) RootDirectory() *directory.DirectoryEntry {
	return i.rootDirectory
}

// HasElTorito returns whether the ISO 9660 image has an El Torito boot record
func (i *ISO9660Image) HasElTorito() bool {
	return i.eltorito != nil
}

// HasRockRidge returns whether the ISO 9660 image has Rock Ridge extensions
func (i *ISO9660Image) HasRockRidge() bool {
	return i.rootDirectory.HasRockRidge()
}

// ExtractAll extracts all files and boot images from the ISO 9660 image
func (i *ISO9660Image) Extract(outputLocation string, includeBootImages bool) (err error) {
	if !i.Parsed() {
		err = i.Parse()
		if err != nil {
			return err
		}
	}

	err = i.ExtractFiles(outputLocation)
	if err != nil {
		return err
	}

	if includeBootImages {
		err = i.ExtractBootImages(filepath.Join(outputLocation, i.options.bootFileLocation))
		if err != nil {
			return err
		}
	}

	return nil
}

// ExtractFiles extracts all files from the ISO 9660 image
func (i *ISO9660Image) ExtractFiles(outputLocation string) error {
	// Ensure the ISO 9660 image has been parsed
	i.logger.V(logging.INFO).Info("Extracting files from ISO 9660 image", "outputLocation", outputLocation)
	if !i.Parsed() {
		if err := i.Parse(); err != nil {
			return fmt.Errorf("failed to parse ISO: %w", err)
		}
	}

	// Get all the entries
	entries, err := i.GetAllEntries()
	if err != nil {
		return fmt.Errorf("failed to get all entries: %w", err)
	}

	// Handle creating all directories first
	for _, entry := range entries {

		if entry.IsDir() {
			fullPath := filepath.Join(outputLocation, entry.FullPath())
			if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
			}
		}
	}

	// Handle extracting all files
	for _, entry := range entries {
		if !entry.IsDir() {
			if err := i.extractFile(entry, filepath.Join(outputLocation, entry.FullPath())); err != nil {
				return fmt.Errorf("failed to extract file %s: %w", entry.FullPath(), err)
			}
		}
	}

	return nil
}

// ExtractBootImages extracts all boot images from the ISO 9660 image
func (i *ISO9660Image) ExtractBootImages(outputLocation string) (err error) {
	// Ensure the output directory exists
	i.logger.V(logging.INFO).Info("Extracting boot images from ISO 9660 image", "outputLocation", outputLocation)
	var stat os.FileInfo
	if stat, err = os.Stat(outputLocation); err != nil && os.IsNotExist(err) {
		if stat != nil && !stat.IsDir() {
			return fmt.Errorf("output location %s exists and is not a directory", outputLocation)
		}
		if err := os.MkdirAll(outputLocation, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create output location %s: %w", outputLocation, err)
		}
	}

	return i.eltorito.ExtractBootImages(i.isoFile, outputLocation)
}

// GetAllEntries returns all the entries in the actively selected Volume Descriptors Root Directory Entry
func (i *ISO9660Image) GetAllEntries() ([]*directory.DirectoryEntry, error) {
	if !i.Parsed() {
		if err := i.Parse(); err != nil {
			return nil, fmt.Errorf("failed to parse ISO: %w", err)
		}
	}

	// Start extracting from the root directory
	return walkAllEntries(i.RootDirectory())
}

// extractFile is a utility function to extract the contents of a file
func (i *ISO9660Image) extractFile(file *directory.DirectoryEntry, fullPath string) error {
	// Strip versioning information if requested
	if i.options.stripVersionInfo {
		fullPath = stripVersion(fullPath)
	}

	name := file.Name()
	// Don't attempt to write files that have no name
	if name == "" || name == "." || name == ".." {
		return nil
	}

	// Open or create the output file
	outFile, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", fullPath, err)
	}
	defer outFile.Close()

	// Read file contents from the ISO
	start := int64(file.Record.LocationOfExtent) * consts.ISO9660_SECTOR_SIZE // Replace logicalBlockSize as needed
	size := int64(file.Record.DataLength)
	buffer := make([]byte, size)

	if _, err = file.IsoReader.ReadAt(buffer, start); err != nil {
		return fmt.Errorf("failed to read file %s from ISO: %w", name, err)
	}

	// Write contents to the output file
	if _, err = outFile.Write(buffer); err != nil {
		return fmt.Errorf("failed to write file %s: %w", fullPath, err)
	}

	return nil
}

// isEltorito is a utility function to determine if the boot system identifier is El Torito
func isElTorito(bootSystemIdentifier string) bool {
	trimmed := strings.TrimRight(bootSystemIdentifier, "\x00")
	return trimmed == consts.EL_TORITO_BOOT_SYSTEM_ID
}

// parsePathTable is a utility function to parse the path table
func parsePathTable(isoReader io.ReaderAt, vd VolumeDescriptor, logger logr.Logger) error {
	// Walk the path table

	start := int(vd.PathTableLocation() * consts.ISO9660_SECTOR_SIZE)
	end := start + int(vd.PathTableSize())

	pathTable := vd.PathTable()

	offset := start
	for offset < end {
		// First, read the fixed 8-byte header to determine the record's actual length.
		header := make([]byte, 8)
		n, err := isoReader.ReadAt(header, int64(offset))
		if err != nil {
			return fmt.Errorf("failed to read path table header at offset %d: %w", offset, err)
		}
		if n < 8 {
			return fmt.Errorf("unexpected EOF reading path table header at offset %d", offset)
		}

		// The 1st byte = DirectoryIdentifierLength
		dirLen := header[0]

		// Total record length = 8 (fixed) + dirLen + padding (1 if dirLen is odd)
		recordLen := 8 + int(dirLen)
		if dirLen%2 != 0 {
			recordLen++
		}

		// Make sure we don't go beyond the table boundary
		if offset+recordLen > end {
			return fmt.Errorf("path table record at offset %d would exceed path table size", offset)
		}

		// Read the entire record in one shot
		buf := make([]byte, recordLen)
		n, err = isoReader.ReadAt(buf, int64(offset))
		if err != nil {
			return fmt.Errorf("failed to read path table record at offset %d: %w", offset, err)
		}
		if n < recordLen {
			return fmt.Errorf("unexpected EOF reading path table record at offset %d", offset)
		}

		// Unmarshal the record (assumes your Unmarshal can parse exactly one record)
		record := NewPathTableRecord(logger)
		if err := record.Unmarshal(buf); err != nil {
			return fmt.Errorf("failed to unmarshal path table record at offset %d: %w", offset, err)
		}

		// Append to your slice of records
		*pathTable = append(*pathTable, record)

		// Advance offset by the size of this record
		offset += recordLen
	}
	return nil
}

// stripVersion is a utility function to strip the version suffix from a file name
func stripVersion(filename string) string {
	if idx := strings.Index(filename, ";"); idx != -1 {
		return filename[:idx]
	}
	return filename
}

// walkAllEntries is a utility function to walk all entries in the directory tree
func walkAllEntries(root *directory.DirectoryEntry) ([]*directory.DirectoryEntry, error) {
	var (
		result []*directory.DirectoryEntry
		queue  = []*directory.DirectoryEntry{root} // initial queue with just the root
	)

	// Perform BFS
	for len(queue) > 0 {
		// Dequeue first element
		current := queue[0]
		queue = queue[1:]

		// Add current entry to result
		result = append(result, current)

		// If it's a directory, fetch children
		if current.IsDir() {
			children, err := current.GetChildren()
			if err != nil {
				return nil, err
			}
			// Enqueue all children
			queue = append(queue, children...)
		}
	}

	return result, nil
}

// BFSAllEntriesParallel performs a parallel breadth-first search of all entries in the directory tree
// TODO: This isn't used, still need to benchmark if it's worth using (probably not)
func BFSAllEntriesParallel(root *directory.DirectoryEntry, maxWorkers int) ([]*directory.DirectoryEntry, error) {
	var (
		result []*directory.DirectoryEntry
		queue  = []*directory.DirectoryEntry{root}
		mu     sync.Mutex
		wg     sync.WaitGroup
		sem    = make(chan struct{}, maxWorkers) // concurrency limiter
		errCh  = make(chan error, 1)
	)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Collect current
		result = append(result, current)

		// Only launch goroutine if it's a directory
		if current.IsDir() {
			wg.Add(1)
			go func(dir *directory.DirectoryEntry) {
				defer wg.Done()
				sem <- struct{}{} // acquire a worker slot
				defer func() { <-sem }()

				children, err := dir.GetChildren()
				if err != nil {
					// signal first error
					select {
					case errCh <- err:
					default:
					}
					return
				}

				// Safely append to queue
				mu.Lock()
				queue = append(queue, children...)
				mu.Unlock()
			}(current)
		}
	}

	// Wait for all parallel fetches
	wg.Wait()

	// Check if any error was reported
	select {
	case err := <-errCh:
		// Return the first encountered error
		return nil, err
	default:
	}

	return result, nil
}
