package iso9660

import (
	"errors"
	"fmt"
	"github.com/bgrewell/iso-kit/pkg/consts"
	"github.com/bgrewell/iso-kit/pkg/filesystem"
	"github.com/bgrewell/iso-kit/pkg/iso9660/boot"
	"github.com/bgrewell/iso-kit/pkg/iso9660/descriptor"
	"github.com/bgrewell/iso-kit/pkg/iso9660/directory"
	"github.com/bgrewell/iso-kit/pkg/iso9660/info"
	"github.com/bgrewell/iso-kit/pkg/iso9660/parser"
	"github.com/bgrewell/iso-kit/pkg/iso9660/pathtable"
	"github.com/bgrewell/iso-kit/pkg/iso9660/systemarea"
	"github.com/bgrewell/iso-kit/pkg/logging"
	"github.com/bgrewell/iso-kit/pkg/option"
	"github.com/bgrewell/iso-kit/pkg/version"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// Open opens an ISO9660 filesystem from the specified reader.
func Open(isoReader io.ReaderAt, opts ...option.OpenOption) (*ISO9660, error) {

	// Set default open options
	emptyCallback := func(currentFilename string, bytesTransferred int64, totalBytes int64, currentFileNumber int, totalFileCount int) {
	}
	openOptions := &option.OpenOptions{
		ReadOnly:                   true,
		ParseOnOpen:                true,
		PreloadDir:                 true,
		StripVersionInfo:           true,
		RockRidgeEnabled:           true,
		ElToritoEnabled:            true,
		PreferJoliet:               false,
		BootFileExtractLocation:    "[BOOT]",
		ExtractionProgressCallback: emptyCallback,
		Logger:                     logging.DefaultLogger(),
	}

	for _, opt := range opts {
		opt(openOptions)
	}

	// Read the System Area
	saBuf := [consts.ISO9660_SECTOR_SIZE * consts.ISO9660_SYSTEM_AREA_SECTORS]byte{}
	if _, err := isoReader.ReadAt(saBuf[:], 0); err != nil {
		return nil, err
	}
	sa := systemarea.SystemArea{
		Contents: saBuf,
	}

	// Create a parser
	p := parser.NewParser(isoReader, openOptions)

	// Read the boot record
	bootRecord, err := p.GetBootRecord()
	if err != nil {
		return nil, err
	}

	// Check for El-Torito boot record
	var et *boot.ElTorito
	if bootRecord != nil && boot.IsElTorito(bootRecord.BootSystemIdentifier) && openOptions.ElToritoEnabled {
		et, err = p.GetElTorito(bootRecord)
		if err != nil {
			return nil, err
		}
	}

	// Read the primary volume descriptor
	pvd, err := p.GetPrimaryVolumeDescriptor()
	if err != nil {
		return nil, err
	}

	// Read the supplementary volume descriptors
	svds, err := p.GetSupplementaryVolumeDescriptors()
	if err != nil {
		return nil, err
	}

	// Read any partition volume descriptors
	partitionvds, err := p.GetVolumePartitionDescriptors()
	if err != nil {
		return nil, err
	}

	// Mark the end of the volume descriptors
	term, err := p.GetVolumeDescriptorSetTerminator()
	if err != nil {
		return nil, err
	}

	// Handle walking the pvd directory records
	pvd.DirectoryRecords, err = p.WalkDirectoryRecords(pvd.RootDirectoryRecord)
	if err != nil {
		return nil, err
	}

	// Handle walking the svd directory records
	for _, svd := range svds {
		svd.DirectoryRecords, err = p.WalkDirectoryRecords(svd.RootDirectoryRecord)
		if err != nil {
			return nil, err
		}
	}

	// Handle processing volume descriptor
	var filesystemEntries []*filesystem.FileSystemEntry
	if openOptions.PreferJoliet && len(svds) > 0 {
		// Open the Joliet filesystem
		filesystemEntries, err = p.BuildFileSystemEntries(svds[0].RootDirectoryRecord, false)
	} else {
		filesystemEntries, err = p.BuildFileSystemEntries(pvd.RootDirectoryRecord, openOptions.RockRidgeEnabled)
	}

	// Handle the path tables
	tables, err := p.GetPathTables(pvd)
	if err != nil {
		return nil, err
	}
	for _, svd := range svds {
		svdTables, err := p.GetPathTables(svd)
		if err != nil {
			return nil, err
		}
		tables = append(tables, svdTables...)
	}

	volumeDescSet := &descriptor.VolumeDescriptorSet{
		Primary:       pvd,
		Supplementary: svds,
		Partition:     partitionvds,
		Boot:          bootRecord,
		Terminator:    term,
	}

	iso := &ISO9660{
		isoReader:           isoReader,
		openOptions:         openOptions, //TODO: Work on making composite options that limit users ability to create based on context but have a single set behind the scenes
		systemArea:          sa,
		volumeDescriptorSet: volumeDescSet,
		pathTables:          tables,
		filesystemEntries:   filesystemEntries,
		elTorito:            et,
		logger:              openOptions.Logger,
		isPacked:            true,
		pendingFiles:        make(map[string][]byte),
	}

	return iso, nil
}

func Create(name string, opts ...option.CreateOption) (*ISO9660, error) {
	// Set default create options
	createOptions := &option.CreateOptions{
		Preparer: fmt.Sprintf("iso-kit %s %s (%s) %s", version.Version(), version.Revision(), version.Branch(), version.Date()),
	}

	for _, opt := range opts {
		opt(createOptions)
	}

	// Create a root directory record
	rootDir := &directory.DirectoryRecord{
		FileIdentifier:                "\x00",
		LengthOfDirectoryRecord:       34, // Standard size for root directory
		ExtendedAttributeRecordLength: 0,
		LocationOfExtent:              18, // Usually starts at sector 18
		DataLength:                    consts.ISO9660_SECTOR_SIZE, // One sector for root directory
		RecordingDateAndTime:          time.Now(),
		FileFlags:                     directory.FileFlags{Directory: true},
		VolumeSequenceNumber:          1,
	}

	// Create system area (First 16 sectors reserved)
	sa := systemarea.SystemArea{}

	// Create primary volume descriptor
	pvd := &descriptor.PrimaryVolumeDescriptor{
		VolumeDescriptorHeader: descriptor.VolumeDescriptorHeader{
			VolumeDescriptorType:    descriptor.TYPE_PRIMARY_DESCRIPTOR,
			StandardIdentifier:      consts.ISO9660_STD_IDENTIFIER,
			VolumeDescriptorVersion: consts.ISO9660_VOLUME_DESC_VERSION,
		},
		PrimaryVolumeDescriptorBody: descriptor.PrimaryVolumeDescriptorBody{
			SystemIdentifier:              "",
			VolumeIdentifier:              name,
			VolumeSpaceSize:               19, // Initially small - will grow as files are added
			VolumeSetSize:                 1,
			VolumeSequenceNumber:          1,
			LogicalBlockSize:              consts.ISO9660_SECTOR_SIZE,
			RootDirectoryRecord:           rootDir,
			VolumeSetIdentifier:           "",
			PublisherIdentifier:           "",
			DataPreparerIdentifier:        createOptions.Preparer,
			ApplicationIdentifier:         "",
			VolumeCreationDateAndTime:     time.Now(),
			VolumeModificationDateAndTime: time.Now(),
			VolumeExpirationDateAndTime:   time.Time{}, // No expiration
			VolumeEffectiveDateAndTime:    time.Now(),
			FileStructureVersion:          1,
		},
	}

	// Create volume descriptor set terminator
	term := descriptor.NewVolumeDescriptorSetTerminator()

	// Initialize supplementary volume descriptors (for Joliet)
	var svds []*descriptor.SupplementaryVolumeDescriptor
	if createOptions.JolietEnabled {
		svd := &descriptor.SupplementaryVolumeDescriptor{
			VolumeDescriptorHeader: descriptor.VolumeDescriptorHeader{
				VolumeDescriptorType:    descriptor.TYPE_SUPPLEMENTARY_DESCRIPTOR,
				StandardIdentifier:      consts.ISO9660_STD_IDENTIFIER,
				VolumeDescriptorVersion: consts.ISO9660_VOLUME_DESC_VERSION,
			},
			SupplementaryVolumeDescriptorBody: descriptor.SupplementaryVolumeDescriptorBody{
				VolumeFlags:                   0,
				SystemIdentifier:              "",
				VolumeIdentifier:              name,
				VolumeSpaceSize:               [8]byte{19, 0, 0, 0, 0, 0, 0, 19}, // BothByteOrder
				RootDirectoryRecord:           rootDir,
				VolumeSetIdentifier:           "",
				PublisherIdentifier:           "",
				DataPreparerIdentifier:        createOptions.Preparer,
				ApplicationIdentifier:         "",
				VolumeCreationDateAndTime:     time.Now(),
				VolumeModificationDateAndTime: time.Now(),
				VolumeExpirationDateAndTime:   time.Time{}, // No expiration
				VolumeEffectiveDateAndTime:    time.Now(),
				FileStructureVersion:          1,
			},
		}
		// Set Joliet escape sequence for Level 3
		copy(svd.SupplementaryVolumeDescriptorBody.EscapeSequences[:], []byte(consts.JOLIET_LEVEL_3_ESCAPE))
		svds = append(svds, svd)
	}

	// Create volume descriptor set
	volumeDescSet := &descriptor.VolumeDescriptorSet{
		Primary:       pvd,
		Supplementary: svds,
		Partition:     nil, // Not commonly used in basic ISO9660
		Boot:          nil, // Boot record would go here for El Torito
		Terminator:    term,
	}

	// Create empty filesystem entries (just root directory for now)
	filesystemEntries := []*filesystem.FileSystemEntry{}

	// Build ISO structure
	iso := &ISO9660{
		isoReader:           nil, // No reader for created ISOs
		createOptions:       createOptions,
		systemArea:          sa,
		volumeDescriptorSet: volumeDescSet,
		pathTables:          nil, // Will be generated during packing
		filesystemEntries:   filesystemEntries,
		elTorito:            nil, // No boot record initially
		logger:              createOptions.Logger,
		isPacked:            false, // Not packed yet
		pendingFiles:        make(map[string][]byte),
	}

	// Add files from root directory if specified
	if createOptions.RootDir != "" {
		err := iso.AddDirectory(createOptions.RootDir, "")
		if err != nil {
			return nil, fmt.Errorf("failed to add root directory: %w", err)
		}
	}

	return iso, nil
}

// ISO9660 represents an ISO9660 filesystem.
type ISO9660 struct {
	// ISO Reader
	isoReader io.ReaderAt
	// Open Options
	openOptions *option.OpenOptions
	// Create Options
	createOptions *option.CreateOptions
	// System Area
	systemArea systemarea.SystemArea
	// Volume Descriptor Set
	volumeDescriptorSet *descriptor.VolumeDescriptorSet
	// Path Tables
	pathTables []*pathtable.PathTable
	// ElTorito Boot Record
	elTorito *boot.ElTorito
	// FileSystemEntries
	filesystemEntries []*filesystem.FileSystemEntry
	// Logger
	logger *logging.Logger
	// isPacked represents if the ISO9660 filesystem is packed and ready to write to disk
	isPacked bool
	// pendingFiles stores data for newly added files that haven't been written to disk yet
	pendingFiles map[string][]byte
}

// GetVolumeID returns the volume identifier of the ISO9660 filesystem.
func (iso *ISO9660) GetVolumeID() string {
	if iso.openOptions.PreferJoliet && iso.volumeDescriptorSet.Supplementary != nil {
		return iso.volumeDescriptorSet.Supplementary[0].VolumeIdentifier()
	}
	return iso.volumeDescriptorSet.Primary.VolumeIdentifier()
}

// GetSystemID returns the system identifier of the ISO9660 filesystem.
func (iso *ISO9660) GetSystemID() string {
	if iso.openOptions.PreferJoliet && iso.volumeDescriptorSet.Supplementary != nil {
		return iso.volumeDescriptorSet.Supplementary[0].SystemIdentifier()
	}
	return iso.volumeDescriptorSet.Primary.SystemIdentifier()
}

// GetVolumeSize returns the size of the ISO9660 filesystem.
func (iso *ISO9660) GetVolumeSize() uint32 {
	return iso.volumeDescriptorSet.Primary.VolumeSpaceSize
}

// GetVolumeSetID returns the volume set identifier of the ISO9660 filesystem.
func (iso *ISO9660) GetVolumeSetID() string {
	if iso.openOptions.PreferJoliet && iso.volumeDescriptorSet.Supplementary != nil {
		return iso.volumeDescriptorSet.Supplementary[0].VolumeSetIdentifier()
	}
	return iso.volumeDescriptorSet.Primary.VolumeSetIdentifier()
}

// GetPublisherID returns the publisher identifier of the ISO9660 filesystem.
func (iso *ISO9660) GetPublisherID() string {
	if iso.openOptions.PreferJoliet && iso.volumeDescriptorSet.Supplementary != nil {
		return iso.volumeDescriptorSet.Supplementary[0].PublisherIdentifier()
	}
	return iso.volumeDescriptorSet.Primary.PublisherIdentifier()
}

// GetDataPreparerID returns the data preparer identifier of the ISO9660 filesystem.
func (iso *ISO9660) GetDataPreparerID() string {
	if iso.openOptions.PreferJoliet && iso.volumeDescriptorSet.Supplementary != nil {
		return iso.volumeDescriptorSet.Supplementary[0].DataPreparerIdentifier()
	}
	return iso.volumeDescriptorSet.Primary.DataPreparerIdentifier()
}

// GetApplicationID returns the application identifier of the ISO9660 filesystem.
func (iso *ISO9660) GetApplicationID() string {
	if iso.openOptions.PreferJoliet && iso.volumeDescriptorSet.Supplementary != nil {
		return iso.volumeDescriptorSet.Supplementary[0].ApplicationIdentifier()
	}
	return iso.volumeDescriptorSet.Primary.ApplicationIdentifier()
}

// GetCopyrightID returns the copyright identifier of the ISO9660 filesystem.
func (iso *ISO9660) GetCopyrightID() string {
	if iso.openOptions.PreferJoliet && iso.volumeDescriptorSet.Supplementary != nil {
		return iso.volumeDescriptorSet.Supplementary[0].CopyrightFileIdentifier()
	}
	return iso.volumeDescriptorSet.Primary.CopyrightFileIdentifier()
}

// GetAbstractID returns the abstract identifier of the ISO9660 filesystem.
func (iso *ISO9660) GetAbstractID() string {
	if iso.openOptions.PreferJoliet && iso.volumeDescriptorSet.Supplementary != nil {
		return iso.volumeDescriptorSet.Supplementary[0].AbstractFileIdentifier()
	}
	return iso.volumeDescriptorSet.Primary.AbstractFileIdentifier()
}

// GetBibliographicID returns the bibliographic identifier of the ISO9660 filesystem.
func (iso *ISO9660) GetBibliographicID() string {
	if iso.openOptions.PreferJoliet && iso.volumeDescriptorSet.Supplementary != nil {
		return iso.volumeDescriptorSet.Supplementary[0].BibliographicFileIdentifier()
	}
	return iso.volumeDescriptorSet.Primary.BibliographicFileIdentifier()
}

// GetCreationDateTime returns the creation date and time of the ISO9660 filesystem.
func (iso *ISO9660) GetCreationDateTime() time.Time {
	if iso.openOptions.PreferJoliet && iso.volumeDescriptorSet.Supplementary != nil {
		return iso.volumeDescriptorSet.Supplementary[0].VolumeCreationDateTime()
	}
	return iso.volumeDescriptorSet.Primary.VolumeCreationDateTime()
}

// GetModificationDateTime returns the modification date and time of the ISO9660 filesystem.
func (iso *ISO9660) GetModificationDateTime() time.Time {
	if iso.openOptions.PreferJoliet && iso.volumeDescriptorSet.Supplementary != nil {
		return iso.volumeDescriptorSet.Supplementary[0].VolumeModificationDateTime()
	}
	return iso.volumeDescriptorSet.Primary.VolumeModificationDateTime()
}

// GetExpirationDateTime returns the expiration date and time of the ISO9660 filesystem.
func (iso *ISO9660) GetExpirationDateTime() time.Time {
	if iso.openOptions.PreferJoliet && iso.volumeDescriptorSet.Supplementary != nil {
		return iso.volumeDescriptorSet.Supplementary[0].VolumeExpirationDateTime()
	}
	return iso.volumeDescriptorSet.Primary.VolumeExpirationDateTime()
}

// GetEffectiveDateTime returns the effective date and time of the ISO9660 filesystem.
func (iso *ISO9660) GetEffectiveDateTime() time.Time {
	if iso.openOptions.PreferJoliet && iso.volumeDescriptorSet.Supplementary != nil {
		return iso.volumeDescriptorSet.Supplementary[0].VolumeEffectiveDateTime()
	}
	return iso.volumeDescriptorSet.Primary.VolumeEffectiveDateTime()
}

// HasJoliet returns true if the ISO9660 filesystem has Joliet extensions.
func (iso *ISO9660) HasJoliet() bool {
	for _, svd := range iso.volumeDescriptorSet.Supplementary {
		if svd.HasJoliet() {
			return true
		}
	}
	return false
}

// HasRockRidge returns true if the ISO9660 filesystem has Rock Ridge extensions.
func (iso *ISO9660) HasRockRidge() bool {
	return iso.volumeDescriptorSet.Primary.HasRockRidge()
}

// HasElTorito returns true if the ISO9660 filesystem has El Torito boot extensions.
func (iso *ISO9660) HasElTorito() bool {
	return iso.elTorito != nil
}

// RootDirectoryLocation returns the location of the root directory in the ISO9660 filesystem.
func (iso *ISO9660) RootDirectoryLocation() uint32 {
	if iso.openOptions.PreferJoliet && iso.volumeDescriptorSet.Supplementary != nil {
		return iso.volumeDescriptorSet.Supplementary[0].RootDirectoryRecord.LocationOfExtent
	}
	return iso.volumeDescriptorSet.Primary.RootDirectoryRecord.LocationOfExtent
}

// ListBootEntries returns a list of all boot entries in the ISO9660 filesystem.
func (iso *ISO9660) ListBootEntries() ([]*filesystem.FileSystemEntry, error) {
	return iso.elTorito.BuildBootImageEntries()
}

// ListFiles returns a list of all files in the ISO9660 filesystem.
func (iso *ISO9660) ListFiles() ([]*filesystem.FileSystemEntry, error) {
	files := make([]*filesystem.FileSystemEntry, 0)
	for _, entry := range iso.filesystemEntries {
		if !entry.IsDir {
			files = append(files, entry)
		}
	}

	return files, nil
}

// ListDirectories returns a list of all directories in the ISO9660 filesystem.
func (iso *ISO9660) ListDirectories() ([]*filesystem.FileSystemEntry, error) {
	dirs := make([]*filesystem.FileSystemEntry, 0)
	for _, entry := range iso.filesystemEntries {
		if entry.IsDir {
			dirs = append(dirs, entry)
		}
	}

	return dirs, nil
}

func (iso *ISO9660) ReadFile(path string) ([]byte, error) {
	// Normalize the path by removing leading slash
	normalizedPath := strings.TrimPrefix(path, "/")
	
	// Check if it's a pending file first
	if iso.pendingFiles != nil {
		if data, exists := iso.pendingFiles[normalizedPath]; exists {
			return data, nil
		}
	}
	
	// Find the file in our filesystem entries
	for _, entry := range iso.filesystemEntries {
		if entry.FullPath == normalizedPath && !entry.IsDir {
			return entry.GetBytes()
		}
	}
	
	return nil, fmt.Errorf("file not found: %s", path)
}

func (iso *ISO9660) AddFile(path string, data []byte) error {
	// Normalize the path by removing leading slash
	normalizedPath := strings.TrimPrefix(path, "/")
	
	// Check if file already exists
	for _, entry := range iso.filesystemEntries {
		if entry.FullPath == normalizedPath {
			return fmt.Errorf("file already exists: %s", path)
		}
	}
	
	// Initialize pendingFiles map if it doesn't exist
	if iso.pendingFiles == nil {
		iso.pendingFiles = make(map[string][]byte)
	}
	
	// Store the file data
	iso.pendingFiles[normalizedPath] = data
	
	// Create a new file system entry
	fileName := filepath.Base(normalizedPath)
	
	// Create directory record for the new file
	record := &directory.DirectoryRecord{
		DataLength:              uint32(len(data)),
		RecordingDateAndTime:    time.Now(),
		FileFlags:               directory.FileFlags{}, // Regular file
		FileIdentifier:          fileName,
		LocationOfExtent:        0, // Will be set during packing/save
		ExtendedAttributeRecordLength: 0,
	}
	
	// Create filesystem entry
	entry := filesystem.NewFileSystemEntry(
		fileName,
		normalizedPath,
		false, // not a directory
		uint32(len(data)),
		0, // location will be set during packing
		nil, // uid
		nil, // gid
		0644, // default file mode
		time.Now(), // create time
		time.Now(), // mod time
		record,
		nil, // reader will be set during save
	)
	
	// Add it to the filesystem entries
	iso.filesystemEntries = append(iso.filesystemEntries, entry)
	
	// Mark as unpacked since we've added a new file
	iso.isPacked = false
	
	return nil
}

func (iso *ISO9660) RemoveFile(path string) error {
	// Normalize the path by removing leading slash
	normalizedPath := strings.TrimPrefix(path, "/")
	
	// Remove from pending files if it exists there
	if iso.pendingFiles != nil {
		if _, exists := iso.pendingFiles[normalizedPath]; exists {
			delete(iso.pendingFiles, normalizedPath)
		}
	}
	
	// Find and remove the file from our filesystem entries
	for i, entry := range iso.filesystemEntries {
		if entry.FullPath == normalizedPath && !entry.IsDir {
			// Remove the entry from the slice
			iso.filesystemEntries = append(iso.filesystemEntries[:i], iso.filesystemEntries[i+1:]...)
			// Mark as unpacked since we've modified the filesystem
			iso.isPacked = false
			return nil
		}
	}
	
	return fmt.Errorf("file not found: %s", path)
}

// AddDirectory recursively adds all files from a directory to the ISO
func (iso *ISO9660) AddDirectory(sourcePath, targetPath string) error {
	// Normalize paths
	sourcePath = filepath.Clean(sourcePath)
	targetPath = strings.TrimPrefix(filepath.Clean(targetPath), "/")
	
	// Check if source directory exists
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("source directory does not exist: %s", sourcePath)
	}
	
	if !sourceInfo.IsDir() {
		return fmt.Errorf("source path is not a directory: %s", sourcePath)
	}
	
	// Walk the directory tree
	return filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Calculate relative path from the source directory
		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}
		
		// Skip the root directory itself
		if relPath == "." {
			return nil
		}
		
		// Build the target path in the ISO
		isoPath := filepath.Join(targetPath, relPath)
		isoPath = filepath.ToSlash(isoPath) // Convert to forward slashes for ISO paths
		
		if info.IsDir() {
			// For directories, we could create them explicitly, but ISO9660 
			// typically creates them implicitly when files are added
			return nil
		} else {
			// Read the file content and add it to the ISO
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", path, err)
			}
			
			return iso.AddFile(isoPath, data)
		}
	})
}

// CreateDirectories creates all directories from the ISO in the specified path.
func (iso *ISO9660) CreateDirectories(path string) error {
	// Ensure output directory exists
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", path, err)
	}

	// Iterate over directory FileSystemEntries and create directories
	dirs, err := iso.ListDirectories()
	if err != nil {
		return fmt.Errorf("failed to list directories: %w", err)
	}
	for _, entry := range dirs {
		dirPath := filepath.Join(path, entry.FullPath)
		if err := os.MkdirAll(dirPath, entry.Mode); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
		}
	}

	return nil
}

// Extract extracts all files and directories from the ISO to the specified path.
func (iso *ISO9660) Extract(path string) error {
	// Create all directories first
	if err := iso.CreateDirectories(path); err != nil {
		return err
	}

	// Extract El Torito boot images if enabled
	if iso.elTorito != nil && iso.openOptions.ElToritoEnabled {
		err := iso.elTorito.ExtractBootImages(iso.isoReader, filepath.Join(path, iso.openOptions.BootFileExtractLocation))
		if err != nil {
			return fmt.Errorf("failed to extract El Torito boot images: %w", err)
		}
	}

	// Get list of files to extract
	files, err := iso.ListFiles()
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	totalFiles := len(files)

	// Extract files
	for i, entry := range files {
		outputPath := filepath.Join(path, entry.FullPath)

		// Ensure parent directories exist
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directories for %s: %w", outputPath, err)
		}

		// if the option to strip version info is enabled, enhanced and rr are not enabled then strip the version info
		if iso.openOptions.StripVersionInfo && !iso.openOptions.RockRidgeEnabled && !iso.openOptions.PreferJoliet {
			outputPath = strings.TrimRight(outputPath, ";1")
		}

		// Open output file for writing
		outFile, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", outputPath, err)
		}
		defer outFile.Close()

		// Stream the file from the ISO
		startOffset := int64(entry.Location) * int64(consts.ISO9660_SECTOR_SIZE)
		size := int64(entry.Size)
		bufferSize := 4096 // 4KB buffer
		buffer := make([]byte, bufferSize)

		var bytesTransferred int64
		for bytesTransferred < size {
			// Read chunk from ISO
			bytesToRead := bufferSize
			if remaining := size - bytesTransferred; remaining < int64(bufferSize) {
				bytesToRead = int(remaining)
			}

			n, err := entry.ReadAt(buffer[:bytesToRead], startOffset+bytesTransferred)
			if err != nil && err != io.EOF {
				return fmt.Errorf("failed to read file %s from ISO: %w", entry.FullPath, err)
			}

			if n == 0 {
				break // Reached EOF
			}

			// Write chunk to file
			if _, err := outFile.Write(buffer[:n]); err != nil {
				return fmt.Errorf("failed to write to file %s: %w", outputPath, err)
			}

			// Update bytes transferred
			bytesTransferred += int64(n)

			// Invoke progress callback
			if iso.openOptions.ExtractionProgressCallback != nil {
				iso.openOptions.ExtractionProgressCallback(outputPath, bytesTransferred, size, i+1, totalFiles)
			}
		}

		// Set correct file permissions
		if err := os.Chmod(outputPath, entry.Mode); err != nil {
			return fmt.Errorf("failed to set permissions on %s: %w", outputPath, err)
		}

		// Set timestamps
		if !entry.ModTime.IsZero() {
			if err := os.Chtimes(outputPath, entry.ModTime, entry.ModTime); err != nil {
				return fmt.Errorf("failed to set timestamps on %s: %w", outputPath, err)
			}
		}
	}

	return nil
}

// SetLogger sets the logger for the ISO9660 filesystem.
func (iso *ISO9660) SetLogger(logger *logging.Logger) {
	iso.logger = logger
}

// GetLogger returns the logger for the ISO9660 filesystem.
func (iso *ISO9660) GetLogger() *logging.Logger {
	return iso.logger
}

// GetLayout returns the layout information for the ISO9660 filesystem.
func (iso *ISO9660) GetLayout() *info.ISOLayout {
	objects := iso.GetObjects()

	return &info.ISOLayout{
		Objects: objects,
	}
}

func (iso *ISO9660) GetObjects() []info.ImageObject {
	var objects []info.ImageObject

	for _, objs := range []([]info.ImageObject){
		iso.systemArea.GetObjects(),
		iso.volumeDescriptorSet.Primary.GetObjects(),
		iso.volumeDescriptorSet.Terminator.GetObjects(),
	} {
		objects = append(objects, objs...)
	}

	if iso.volumeDescriptorSet.Boot != nil {
		objects = append(objects, iso.volumeDescriptorSet.Boot.GetObjects()...)
	}

	if iso.volumeDescriptorSet.Supplementary != nil {
		for _, svd := range iso.volumeDescriptorSet.Supplementary {
			objects = append(objects, svd.GetObjects()...)
		}
	}

	if iso.volumeDescriptorSet.Partition != nil {
		for _, pvd := range iso.volumeDescriptorSet.Partition {
			objects = append(objects, pvd.GetObjects()...)
		}
	}

	if iso.pathTables != nil {
		for _, pt := range iso.pathTables {
			objects = append(objects, pt.GetObjects()...)
		}
	}

	if iso.elTorito != nil {
		objects = append(objects, iso.elTorito.GetObjects()...)
	}
	return objects
}

// Pack prepares the ISO for writing by calculating file locations and preparing data structures
func (iso *ISO9660) Pack() error {
	if iso.isPacked {
		return nil // Already packed
	}
	
	// This is a simplified packing implementation
	// In a full implementation, this would calculate proper sector locations,
	// build path tables, create directory records, etc.
	
	// For now, just mark as packed to allow Save to work
	iso.isPacked = true
	return nil
}

func (iso *ISO9660) Save(writer io.WriterAt) error {
	// Ensure the ISO is packed and all objects have been assigned locations
	if !iso.isPacked {
		err := iso.Pack()
		if err != nil {
			return fmt.Errorf("failed to pack ISO: %w", err)
		}
	}

	// Get all objects
	objects := iso.GetObjects()

	// Sort objects by offset before writing
	slices.SortFunc(objects, func(a, b info.ImageObject) int {
		return int(a.Offset() - b.Offset())
	})

	// Write each object at its assigned offset
	for _, obj := range objects {
		// Get raw data for the object
		data, err := obj.Marshal()
		if err != nil {
			return fmt.Errorf("failed to marshal object %s: %w", obj.Name(), err)
		}

		// Write data at the correct offset
		_, err = writer.WriteAt(data, obj.Offset())
		if err != nil {
			return fmt.Errorf("failed to write object %s at offset %d: %w", obj.Name(), obj.Offset(), err)
		}
	}

	return nil

	//sectorSize := int64(consts.ISO9660_SECTOR_SIZE)
	//saOffset := int64(0)
	//
	//// Calculate offsets for descriptors
	//pvdSize := sectorSize
	//bootSize := int64(0)
	//if iso.bootRecord != nil {
	//	bootSize = sectorSize
	//}
	//svdSize := int64(len(iso.svds)) * sectorSize
	//ptvdSize := int64(len(iso.partitionvds)) * sectorSize
	//
	//pvdOffset := saOffset + consts.ISO9660_SYSTEM_AREA_SECTORS*sectorSize
	//bootOffset := pvdOffset + pvdSize
	//svdOffset := bootOffset + bootSize
	//ptvdOffset := svdOffset + svdSize
	//termOffset := ptvdOffset + ptvdSize
	//
	//type descriptorSetEntry struct {
	//	descriptor descriptor.VolumeDescriptor
	//	offset     int64
	//}
	//descriptorSet := []*descriptorSetEntry{
	//	{descriptor: iso.pvd, offset: pvdOffset},
	//}
	//if iso.bootRecord != nil {
	//	descriptorSet = append(descriptorSet,
	//		&descriptorSetEntry{descriptor: iso.bootRecord, offset: bootOffset},
	//	)
	//}
	//for i, svd := range iso.svds {
	//	descriptorSet = append(descriptorSet,
	//		&descriptorSetEntry{descriptor: svd, offset: svdOffset + int64(i)*sectorSize},
	//	)
	//}
	//for i, ptvd := range iso.partitionvds {
	//	descriptorSet = append(descriptorSet,
	//		&descriptorSetEntry{descriptor: ptvd, offset: ptvdOffset + int64(i)*sectorSize},
	//	)
	//}
	//descriptorSet = append(descriptorSet,
	//	&descriptorSetEntry{descriptor: descriptor.NewVolumeDescriptorSetTerminator(), offset: termOffset})
	//
	//// Write system area
	//_, err := writer.WriteAt(iso.systemArea.Contents[:], 0)
	//if err != nil {
	//	return err
	//}
	//
	//// Write descriptor set
	//for _, entry := range descriptorSet {
	//	if err = writeDescriptor(writer, entry.descriptor, entry.offset); err != nil {
	//		return err
	//	}
	//}
	//
	//// Write path tables according to their location in the volume descriptors
	//err = iso.writePathTables(writer)
	//if err != nil {
	//	return err
	//}

	// Write directory records

	//pathTableOffset := svdOffset + (len(iso.svds) * sectorSize)
	//// Directory Record offsets should be
	//
	//

	//// TODO: Clean up all of the offset calculations, instead this should be calculated somewhere
	////       else and the locations should just be used to write the data or this should be wrapped
	////       nicely in a function.
	//// Write volume descriptor set
	//pvdOffset := int64(16 * sectorSize)
	//err = writeDescriptor(writer, iso.pvd, pvdOffset)
	//if err != nil {
	//	return err
	//}
	//
	//svdOffset := pvdOffset + int64(1*sectorSize)
	//for i, svd := range iso.svds {
	//	err = writeDescriptor(writer, svd, svdOffset)
	//	if err != nil {
	//		return err
	//	}
	//	svdOffset = svdOffset+int64((i+1)*sectorSize)
	//}
	//
	//ptvdOffset := svdOffset
	//for i, pvd := range iso.partitionvds {
	//	if err = writeDescriptor(writer, pvd, ptvdOffset); err != nil {
	//		return err
	//	}
	//	ptvdOffset = ptvdOffset+int64((i+1)*sectorSize)
	//}
	//
	//bootOffset := ptvdOffset
	//if iso.bootRecord != nil {
	//	if err = writeDescriptor(writer, iso.bootRecord, bootOffset); err != nil {
	//		return err
	//	}
	//	bootOffset = bootOffset + int64(1*sectorSize)
	//}
	//
	//tr := descriptor.NewVolumeDescriptorSetTerminator()
	//if err := writeDescriptor(writer, tr, bootOffset); err != nil {
	//	return err
	//}

	//// 3: Write path tables (Little & Big Endian versions)
	//if err = iso.writePathTables(writer); err != nil {
	//	return err
	//}
	//
	//// 4: Write directory records (Root & Subdirectories)
	//totalRecords := append(iso.pvdDirectoryRecords, iso.svdDirectoryRecords...)
	//for i, dr := range totalRecords {
	//	drOffset :=
	//}
	//
	//
	//// 5: Write file contents (Ensuring correct logical block placement)
	//if err := iso.writeFileData(writer); err != nil {
	//	return err
	//}
	//
	//// 6: Align to sector size (Padding to 2048-byte boundaries)
	//if err := padToSector(writer); err != nil {
	//	return err
	//}

	return nil
}

// Close closes the ISO9660 filesystem.
func (iso *ISO9660) Close() error {
	if f, ok := iso.isoReader.(*os.File); ok {
		return f.Close()
	}
	return nil
}

// writePathTables writes the path tables to the ISO9660 filesystem.
func (iso *ISO9660) writePathTables(writer io.WriterAt) error {
	//if iso.pvd == nil {
	//	return errors.New("PVD is missing")
	//}
	//
	//buf, err := iso.pvdLPathTable.Marshal(true)
	//if err != nil {
	//	return err
	//}
	//
	//if _, err = writer.WriteAt(buf, int64(iso.pvd.LocationOfTypeLPathTable)*consts.ISO9660_SECTOR_SIZE); err != nil {
	//	return err
	//}
	//
	//buf, err = iso.pvdMPathTable.Marshal(false)
	//if err != nil {
	//	return err
	//}
	//
	//if _, err = writer.WriteAt(buf, int64(iso.pvd.LocationOfTypeMPathTable)*consts.ISO9660_SECTOR_SIZE); err != nil {
	//	return err
	//}
	//
	//if iso.svdLPathTable != nil {
	//	buf, err = iso.svdLPathTable.Marshal(true)
	//	if err != nil {
	//		return err
	//	}
	//
	//	if _, err = writer.WriteAt(buf, int64(iso.svds[0].LocationOfTypeLPathTable)*consts.ISO9660_SECTOR_SIZE); err != nil {
	//		return err
	//	}
	//}
	//
	//if iso.svdMPathTable != nil {
	//	buf, err = iso.svdMPathTable.Marshal(false)
	//	if err != nil {
	//		return err
	//	}
	//
	//	if _, err = writer.WriteAt(buf, int64(iso.svds[0].LocationOfTypeMPathTable)*consts.ISO9660_SECTOR_SIZE); err != nil {
	//		return err
	//	}
	//}

	return nil
}

func (iso *ISO9660) writeDirectoryRecords(writer io.WriterAt) error {

	//var rootDirOffset uint32
	//
	//// TODO: Write PVD Directory records starting with the root
	//rootDirOffset = iso.pvd.RootDirectoryRecord.LocationOfExtent

	// TODO: Write SVD Directory records starting with the root

	//sectorSize := consts.ISO9660_SECTOR_SIZE
	//
	//// Root Directory
	//rootDirData, err := iso.
	//if err != nil {
	//	return err
	//}
	//if _, err := writer.Write(rootDirData[:]); err != nil {
	//	return err
	//}
	//
	//// Write Subdirectories
	//for _, dir := range iso.directories {
	//	dirData, err := dir.Marshal()
	//	if err != nil {
	//		return err
	//	}
	//
	//	// Ensure correct sector alignment
	//	padding := sectorSize - (len(dirData) % sectorSize)
	//	if padding < sectorSize {
	//		dirData = append(dirData, make([]byte, padding)...)
	//	}
	//
	//	if _, err := writer.Write(dirData[:]); err != nil {
	//		return err
	//	}
	//}
	//
	//return nil
	return errors.New("not implemented")
}

func (iso *ISO9660) writeFileData(writer io.Writer) error {
	//sectorSize := consts.ISO9660_SECTOR_SIZE
	//
	//for _, file := range iso. {
	//	fileData, err := file.ReadData()
	//	if err != nil {
	//		return err
	//	}
	//
	//	if _, err := writer.Write(fileData); err != nil {
	//		return err
	//	}
	//
	//	// Align to sector size
	//	padding := sectorSize - (len(fileData) % sectorSize)
	//	if padding < sectorSize {
	//		if _, err := writer.Write(make([]byte, padding)); err != nil {
	//			return err
	//		}
	//	}
	//}
	//
	//return nil
	return errors.New("not implemented")
}

func padToSector(writer io.Writer) error {
	//sectorSize := consts.ISO9660_SECTOR_SIZE
	//
	//// Get current file position
	//offset, err := writer.Seek(0, io.SeekCurrent)
	//if err != nil {
	//	return err
	//}
	//
	//// Compute padding
	//padding := sectorSize - (offset % sectorSize)
	//if padding < sectorSize {
	//	_, err := writer.Write(make([]byte, padding))
	//	return err
	//}
	//
	//return nil
	return errors.New("not implemented")
}

func writeDescriptor(writer io.WriterAt, descriptor descriptor.VolumeDescriptor, offset int64) error {
	data, err := descriptor.Marshal()
	if err != nil {
		return err
	}

	n, err := writer.WriteAt(data[:], offset)
	if err != nil {
		return err
	}
	if n != len(data) {
		return errors.New("failed to write descriptor")
	}

	return nil
}

func writePathTable(writer io.Writer, location uint32, littleEndian bool) error {
	//if location == 0 {
	//	return nil // Skip if no path table is set
	//}
	//
	//pathTable := generatePathTable(littleEndian) // Implement the path table generator
	//data := pathTable.Marshal()
	//
	//_, err := writer.Write(data)
	//return err
	return errors.New("not implemented")
}
