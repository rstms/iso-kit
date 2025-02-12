package pkg

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/bgrewell/iso-kit/pkg/consts"
	"github.com/bgrewell/iso-kit/pkg/descriptor"
	"github.com/bgrewell/iso-kit/pkg/directory"
	"github.com/bgrewell/iso-kit/pkg/eltorito"
	"github.com/bgrewell/iso-kit/pkg/logging"
	"github.com/bgrewell/iso-kit/pkg/options"
	"github.com/bgrewell/iso-kit/pkg/path"
	"github.com/bgrewell/iso-kit/pkg/systemarea"
	"github.com/go-logr/logr"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// ISO9660Image represents an ISO 9660 image
type ISO9660Image struct {
	SystemArea                     systemarea.SystemArea
	PrimaryVolumeDescriptor        *descriptor.PrimaryVolumeDescriptor
	SupplementaryVolumeDescriptors []*descriptor.SupplementaryVolumeDescriptor
	BootRecordVolumeDescriptor     *descriptor.BootRecordVolumeDescriptor
	eltorito                       *eltorito.ElTorito
	isoFile                        *os.File
	rootDirectory                  *directory.DirectoryEntry
	Options                        options.Options
	logger                         logr.Logger
	parsed                         bool
}

// Open opens an existing ISO 9660 image file
func (i *ISO9660Image) Open(isoLocation string) (err error) {

	// Pull the logger out of the Options
	i.logger = i.Options.Logger

	// Read the ISO 9660 image from the specified location
	i.isoFile, err = os.Open(isoLocation)
	if err != nil {
		return err
	}

	// Parse the iso if requested
	if i.Options.ParseOnOpen {
		err = i.Parse()
		if err != nil {
			return err
		}
	}

	return nil
}

// Create creates a new ISO 9660 image file
func (i *ISO9660Image) Create(rootPath string) (err error) {

	// Parse the directory, create directory records and path tables
	// If Rock Ridge is enabled, populate the Rock Ridge extensions
	dirRecords, err := directory.BuildDirectoryRecords(rootPath, i.logger)
	for _, record := range dirRecords {
		i.logger.V(logging.TRACE).Info("Directory Record", "record", record)
	}

	// Populate a primary volume descriptor

	// If el torito is enabled, look at the boot file location and populate the boot record volume descriptor

	// If enhanced volume descriptors are enabled, populate the supplementary volume descriptors

	return nil
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
	i.SystemArea = systemarea.SystemArea(sa)

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
		vd, err := descriptor.ParseVolumeDescriptor(vdBytes, i.logger)
		if err != nil {
			return fmt.Errorf("failed to parse volume descriptor at offset %d: %w", idx, err)
		}

		switch vd.Type() {
		case descriptor.VolumeDescriptorPrimary:
			i.logger.V(logging.DEBUG).Info("Processing primary volume descriptor", "idx", idx)
			pvd, err := descriptor.ParsePrimaryVolumeDescriptor(vd, i.isoFile, i.Options.RockRidgeEnabled, i.logger)
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
		case descriptor.VolumeDescriptorSupplementary:
			if !i.Options.PreferEnhancedVD {
				i.logger.V(logging.INFO).Info("Enhanced parsing disabled. Skipping Enhanced Volume Descriptor")
				continue
			}
			i.logger.V(logging.DEBUG).Info("Processing supplementary volume descriptor", "idx", idx)
			svd, err := descriptor.ParseSupplementaryVolumeDescriptor(vd, i.isoFile, i.Options.RockRidgeEnabled, i.logger)
			if err != nil {
				return fmt.Errorf("failed to parse supplementary volume descriptor: %w", err)
			}
			if i.SupplementaryVolumeDescriptors == nil {
				i.SupplementaryVolumeDescriptors = make([]*descriptor.SupplementaryVolumeDescriptor, 0)
				if i.Options.PreferEnhancedVD && svd.IsJoliet() {
					i.rootDirectory = svd.RootDirectoryEntry
				}
			}
			i.SupplementaryVolumeDescriptors = append(i.SupplementaryVolumeDescriptors, svd)
			i.logger.V(logging.DEBUG).Info("Processing complete for supplementary volume descriptor", "idx", idx)
		case descriptor.VolumeDescriptorBootRecord:
			i.logger.V(logging.DEBUG).Info("Processing boot record volume descriptor", "idx", idx)
			i.BootRecordVolumeDescriptor, err = descriptor.ParseBootRecordVolumeDescriptor(vd, i.logger)
			if err != nil {
				return fmt.Errorf("failed to parse boot record volume descriptor: %w", err)
			}
			if isElTorito(i.BootRecordVolumeDescriptor.BootSystemIdentifier) && i.Options.ElToritoEnabled {
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
		case descriptor.VolumeDescriptorPartition:
			i.logger.Error(fmt.Errorf("Volume Descriptor Partition parser not implemented"), "The volume descriptor partition type is not yet supported")
		case descriptor.VolumeDescriptorSetTerminator:
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
		err = i.ExtractBootImages(filepath.Join(outputLocation, i.Options.BootFileLocation))
		if err != nil {
			return err
		}
	}

	return nil
}

// ExtractFiles extracts all files from the ISO 9660 image with progress updates.
func (i *ISO9660Image) ExtractFiles(outputLocation string) error {
	// Ensure the ISO 9660 image has been parsed
	i.logger.V(logging.DEBUG).Info("Extracting files from ISO 9660 image", "outputLocation", outputLocation)
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

	// Filter out all file entries and count them
	var fileEntries []*directory.DirectoryEntry
	for _, entry := range entries {
		if !entry.IsDir() {
			fileEntries = append(fileEntries, entry)
		}
	}
	totalFileCount := len(fileEntries)
	currentFileNumber := 0

	// Handle creating all directories first
	for _, entry := range entries {
		if entry.IsDir() {
			fullPath := filepath.Join(outputLocation, entry.FullPath())
			if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
			}
		}
	}

	// Handle extracting all files with progress updates
	for _, entry := range fileEntries {
		currentFileNumber++

		fullPath := filepath.Join(outputLocation, entry.FullPath())

		// Strip versioning information if requested
		if i.Options.StripVersionInfo {
			fullPath = stripVersion(fullPath)
		}

		// Extract the file with progress updates
		if err := i.extractFileWithProgress(entry, fullPath, currentFileNumber, totalFileCount); err != nil {
			return fmt.Errorf("failed to extract file %s: %w", entry.FullPath(), err)
		}
	}

	return nil
}

// ExtractBootImages extracts all boot images from the ISO 9660 image
func (i *ISO9660Image) ExtractBootImages(outputLocation string) (err error) {
	// Ensure the output directory exists
	i.logger.V(logging.DEBUG).Info("Extracting boot images from ISO 9660 image", "outputLocation", outputLocation)
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

// Write writes the ISO 9660 image to the specified path
func (i *ISO9660Image) Write(path string) error {

	// Open the file and set to i.isoFile
	// Write the system area
	// Write the primary volume descriptor
	// Write the path table
	// Write the secondary volume descriptors

	return errors.New("ISO 9660 image writing is not yet implemented")
}

// extractFileWithProgress is a utility function to extract the contents of a file with progress updates
func (i *ISO9660Image) extractFileWithProgress(file *directory.DirectoryEntry, fullPath string, currentFileNumber int, totalFileCount int) error {
	// Open or create the output file
	outFile, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", fullPath, err)
	}
	defer outFile.Close()

	// Calculate the byte offset and size
	start := int64(file.Record.LocationOfExtent) * consts.ISO9660_SECTOR_SIZE // Replace logicalBlockSize as needed
	size := int64(file.Record.DataLength)
	bufferSize := 4096 // 4KB buffer
	buffer := make([]byte, bufferSize)

	bytesTransferred := int64(0)

	for bytesTransferred < size {
		// Determine the number of bytes to read in this iteration
		bytesToRead := bufferSize
		remaining := size - bytesTransferred
		if remaining < int64(bufferSize) {
			bytesToRead = int(remaining)
		}

		// Read bytes from the ISO
		n, err := file.IsoReader.ReadAt(buffer[:bytesToRead], start+bytesTransferred)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read file %s from ISO: %w", fullPath, err)
		}

		if n == 0 {
			break // Reached EOF
		}

		// Write bytes to the output file
		if _, err := outFile.Write(buffer[:n]); err != nil {
			return fmt.Errorf("failed to write to file %s: %w", fullPath, err)
		}

		// Update bytes transferred
		bytesTransferred += int64(n)

		// Invoke the progress callback if it's set
		if i.Options.ProgressCallback != nil {
			i.Options.ProgressCallback(
				file.FullPath(),   // currentFilename
				bytesTransferred,  // bytesTransferred
				size,              // totalBytes
				currentFileNumber, // currentFileNumber
				totalFileCount,    // totalFileCount
			)
		}
	}

	return nil
}

// extractFile is a utility function to extract the contents of a file
func (i *ISO9660Image) extractFile(file *directory.DirectoryEntry, fullPath string) error {
	// Strip versioning information if requested
	if i.Options.StripVersionInfo {
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
func parsePathTable(isoReader io.ReaderAt, vd descriptor.VolumeDescriptor, logger logr.Logger) error {
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
		record := path.NewPathTableRecord(logger)
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
