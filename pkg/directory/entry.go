package directory

import (
	"fmt"
	"github.com/bgrewell/iso-kit/pkg/consts"
	"github.com/bgrewell/iso-kit/pkg/encoding"
	"github.com/bgrewell/iso-kit/pkg/logging"
	"io"
	"io/fs"
	"os"
	"path"
	"time"
)

// Ensure that DirectoryEntry implements the os.FileInfo interface.
var _ fs.FileInfo = DirectoryEntry{}

// DirectoryEntry is an os.FileInfo compatible wrapper around a DirectoryRecord.
type DirectoryEntry struct {
	Record           *DirectoryRecord  // Reference to the underlying DirectoryRecord
	IsoReader        io.ReaderAt       // Reference to the underlying ISO image reader
	children         []*DirectoryEntry // Lazily populated children
	parentPath       string            // Parent path of the directory entry
	StripVersionInfo bool              // Strip version info from filenames (e.g., ";1")
}

// Name returns the name of the directory entry. If the entry has Rock Ridge extensions, the Rock Ridge name is
// returned. Otherwise, the FileIdentifier is returned.
func (d DirectoryEntry) Name() string {
	if d.HasRockRidge() && d.Record.rockRidgeName != nil {
		logging.Logger().Tracef("Using Rock Ridge name: %s", *d.Record.rockRidgeName)
		return *d.Record.rockRidgeName
	}

	switch d.Record.FileIdentifier {
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
		logging.Logger().Tracef("Using Rock Ridge permissions: %v", d.Record.rockRidgePermissions)
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
			logging.Logger().Tracef("Using Rock Ridge permissions: %v", perms)
			return perms.Mode.IsDir()
		}
	}
	return d.Record.FileFlags.Directory
}

// Sys returns the underlying system-specific data.
func (d DirectoryEntry) Sys() any {
	//TODO look into what would be appropriate to return here
	return nil
}

// FullPath returns the full path of the directory entry.
func (d DirectoryEntry) FullPath() string {
	return path.Join(d.parentPath, d.Name())
}

// HasRockRidge returns true if the directory entry has Rock Ridge extensions.
func (d DirectoryEntry) HasRockRidge() bool {
	hasRR := d.Record.HasRockRidge()
	logging.Logger().Tracef("DirectoryEntry has Rock Ridge: %t", hasRR)
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

	logging.Logger().Tracef("=== Processing directory extent: %x", d.Record.LocationOfExtent)

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
		logging.Logger().Tracef("Read %d bytes from directory sector at offset %d", n, readOffset)

		// Process each directory entry within this buffer
		for entryOffset := 0; entryOffset < len(buffer); {
			entryLength := int(buffer[entryOffset])
			if entryLength == 0 {
				break // End of entries in this sector
			}

			logging.Logger().Tracef("Processing directory entry at offset %d with length %d", entryOffset, entryLength)

			// Unmarshal directory record
			record := &DirectoryRecord{Joliet: d.Record.Joliet}
			if err := record.Unmarshal(buffer[entryOffset:entryOffset+entryLength], d.IsoReader); err != nil {
				return fmt.Errorf("failed to parse directory record: %w", err)
			}
			logging.Logger().Tracef("Unmarshalled directory record: %v", record.FileIdentifier)

			// Skip special entries (0x00, 0x01)
			if len(record.FileIdentifier) == 1 && (record.FileIdentifier[0] == 0x00 || record.FileIdentifier[0] == 0x01) {
				logging.Logger().Tracef("Skipping special entry: %x", record.FileIdentifier[0])
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
				logging.Logger().Tracef("Processing child directory: %s", child.Name())
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
