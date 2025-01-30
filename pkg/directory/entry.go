package directory

import (
	"fmt"
	"github.com/bgrewell/iso-kit/pkg/consts"
	"github.com/bgrewell/iso-kit/pkg/encoding"
	"github.com/bgrewell/iso-kit/pkg/logging"
	"github.com/go-logr/logr"
	"io"
	"io/fs"
	"os"
	"path"
	"time"
)

// Ensure that DirectoryEntry implements the os.FileInfo interface.
var _ fs.FileInfo = DirectoryEntry{}

// NewEntry creates a new DirectoryEntry instance.
func NewEntry(record *DirectoryRecord, reader io.ReaderAt, logger logr.Logger) *DirectoryEntry {
	return &DirectoryEntry{
		Record:    record,
		IsoReader: reader,
		logger:    logger,
	}
}

// DirectoryEntry is an os.FileInfo compatible wrapper around a DirectoryRecord.
type DirectoryEntry struct {
	Record     *DirectoryRecord  // Reference to the underlying DirectoryRecord
	IsoReader  io.ReaderAt       // Reference to the underlying ISO image reader
	children   []*DirectoryEntry // Lazily populated children
	parentPath string            // Parent path of the directory entry
	logger     logr.Logger       // Logger
}

// Name returns the name of the directory entry. If the entry has Rock Ridge extensions, the Rock Ridge name is
// returned. Otherwise, the FileIdentifier is returned.
func (d DirectoryEntry) Name() string {
	if d.HasRockRidge() && d.Record.rockRidgeName != nil {
		d.logger.V(logging.TRACE).Info("Using Rock Ridge name",
			"name", *d.Record.rockRidgeName, "identifier", d.Record.FileIdentifier)
		return *d.Record.rockRidgeName
	}

	switch d.Record.FileIdentifier { //TODO: Revisit, should just be returning '.' and '..'?
	case "\x00":
		return ""
	case "\x01":
		return "<parent>"
	default:
		return d.Record.FileIdentifier
	}
}

// Size returns the size of the directory entry.
func (d DirectoryEntry) Size() int64 {
	return int64(d.Record.DataLength)
}

// Mode returns the file mode bits for the directory entry.
func (d DirectoryEntry) Mode() fs.FileMode {
	if d.HasRockRidge() && d.Record.rockRidgePermissions != nil {
		d.logger.V(logging.TRACE).Info("Using Rock Ridge permissions",
			"permissions", d.Record.rockRidgePermissions, "identifier", d.Record.FileIdentifier)
		return d.Record.rockRidgePermissions.Mode
	}

	var mode os.FileMode
	if d.IsDir() {
		mode |= os.ModeDir
	}
	return mode
}

// ModTime returns the recording date and time of the directory entry.
func (d DirectoryEntry) ModTime() time.Time {
	if t, err := encoding.DecodeDirectoryTime(d.Record.RecordingDateAndTime); err == nil {
		return t
	}
	return time.Time{}
}

// IsDir returns true if the directory entry represents a directory.
func (d DirectoryEntry) IsDir() bool {
	if d.HasRockRidge() {
		if perms := d.Record.rockRidgePermissions; perms != nil {
			d.logger.V(logging.TRACE).Info("Using Rock Ridge permissions",
				"IsDir", perms.Mode.IsDir(), "identifier", d.Record.FileIdentifier)
			return perms.Mode.IsDir()
		}
	}
	return d.Record.FileFlags.Directory
}

// Sys returns the underlying system-specific data.
func (d DirectoryEntry) Sys() any {
	d.logger.V(logging.TRACE).Info("Sys() called but it is not implemented", "return", nil, "name", d.Name())
	return nil
}

// FullPath returns the full path of the directory entry.
func (d DirectoryEntry) FullPath() string {
	return path.Join(d.parentPath, d.Name())
}

// HasRockRidge returns true if the directory entry has Rock Ridge extensions.
func (d DirectoryEntry) HasRockRidge() bool {
	hasRR := d.Record.HasRockRidge()
	d.logger.V(logging.TRACE).Info("DirectoryEntry has Rock Ridge", "hasRR", hasRR, "identifier", d.Record.FileIdentifier)
	return hasRR
}

// IsRootEntry returns true if the directory entry is the root entry.
func (d DirectoryEntry) IsRootEntry() bool {
	return d.Record.FileIdentifier == "\x00"
}

// GetChildren returns the children of the directory entry.
func (d *DirectoryEntry) GetChildren() ([]*DirectoryEntry, error) {
	// If children are already populated, return them early
	if d.children != nil {
		return d.children, nil
	}

	// Track nodes that have been visited to prevent infinite recursion
	visited := make(map[uint32]bool)

	// Populate the children
	if err := d.PopulateChildren(visited, path.Join(d.parentPath, d.Name())); err != nil {
		return nil, err
	}

	return d.children, nil
}

// PopulateChildren recursively populates the children of the directory entry.
func (d *DirectoryEntry) PopulateChildren(visited map[uint32]bool, parentPath string) error {
	// Ensure that the DirectoryEntry is actually a directory
	if !d.IsDir() {
		return fmt.Errorf("cannot populate children for a file")
	}

	// Prevent revisiting the same directory extent
	if visited[d.Record.LocationOfExtent] {
		return nil
	}
	visited[d.Record.LocationOfExtent] = true

	d.logger.V(logging.TRACE).Info("Processing directory extent", "extent", d.Record.LocationOfExtent)

	// Create a slice to hold the child DirectoryEntries
	var children []*DirectoryEntry

	// Prepare to read the directory data
	sectorSize := int64(consts.ISO9660_SECTOR_SIZE)
	buffer := make([]byte, sectorSize)
	location := int64(d.Record.LocationOfExtent)
	length := int64(d.Record.DataLength)

	// Read directory data in sector-sized chunks
	for offset := int64(0); offset < length; offset += sectorSize {
		readOffset := (location * sectorSize) + offset
		n, err := d.IsoReader.ReadAt(buffer, readOffset)
		if err != nil {
			return fmt.Errorf("failed to read directory sector: %w", err)
		}
		d.logger.V(logging.TRACE).Info("Read directory sector", "offset", readOffset, "length", n)

		// Process each directory entry within this buffer
		for entryOffset := 0; entryOffset < len(buffer); {
			entryLength := int(buffer[entryOffset])
			if entryLength == 0 {
				break // End of entries in this sector
			}

			d.logger.V(logging.TRACE).Info("Processing directory entry", "offset", entryOffset, "length", entryLength)

			// Unmarshal directory record
			record := NewRecord(d.logger)
			record.Joliet = d.Record.Joliet
			if err := record.Unmarshal(buffer[entryOffset:entryOffset+entryLength], d.IsoReader); err != nil {
				return fmt.Errorf("failed to parse directory record: %w", err)
			}
			d.logger.V(logging.TRACE).Info("Unmarshalled directory record", "identifier", record.FileIdentifier)

			// Skip special entries (0x00, 0x01)
			if len(record.FileIdentifier) == 1 && (record.FileIdentifier[0] == 0x00 || record.FileIdentifier[0] == 0x01) {
				d.logger.V(logging.TRACE).Info("Skipping special entry", "identifier", record.FileIdentifier)
				entryOffset += entryLength
				continue
			}

			// Build the child entry
			child := &DirectoryEntry{
				Record:     record,
				IsoReader:  d.IsoReader,
				parentPath: parentPath,
			}

			// Recursively populate children if it's a directory
			if child.IsDir() {
				d.logger.V(logging.TRACE).Info("Processing child directory", "name", child.Name())
				if err := child.PopulateChildren(visited, path.Join(child.parentPath, child.Name())); err != nil {
					return fmt.Errorf("failed to populate children for %s: %w", child.Name(), err)
				}
			}

			children = append(children, child)
			entryOffset += entryLength
		}
	}

	// Assign the collected children back to this DirectoryEntry
	d.children = children
	return nil
}
