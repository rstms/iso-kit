package path

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/bgrewell/iso-kit/pkg/logging"
	"github.com/go-logr/logr"
)

// NewPathTableRecord creates a new PathTableRecord with the provided logger.
// Parameters:
// - logger: logr.Logger - the logger to be used by the PathTableRecord.
// Returns:
// - *PathTableRecord - a pointer to the newly created PathTableRecord.
func NewPathTableRecord(logger logr.Logger) *PathTableRecord {
	return &PathTableRecord{logger: logger}
}

// PathTableRecord represents a record in the path table.
type PathTableRecord struct {
	DirectoryIdentifierLength     byte        // Directory identifier length
	ExtendedAttributeRecordLength byte        // Extended attribute record length
	LocationOfExtent              uint32      // Location of extent
	ParentDirectoryNumber         uint16      // Parent directory number
	DirectoryIdentifier           string      // Directory identifier
	Padding                       []byte      // Padding to align record if identifier length is odd
	logger                        logr.Logger // Logger
}

// Unmarshal parses the Path Table Record from the given data slice.
// Parameters:
// - data: []byte - the byte slice containing the path table record data.
// Returns:
// - error - an error if the data is invalid or parsing fails.
func (ptr *PathTableRecord) Unmarshal(data []byte) error {
	if len(data) < 9 {
		return errors.New("invalid data length")
	}

	// Parse fields
	ptr.DirectoryIdentifierLength = data[0]
	ptr.ExtendedAttributeRecordLength = data[1]
	ptr.LocationOfExtent = binary.LittleEndian.Uint32(data[2:6])
	ptr.ParentDirectoryNumber = binary.LittleEndian.Uint16(data[6:8])

	// Check bounds for DirectoryIdentifier
	dirIDEnd := 8 + int(ptr.DirectoryIdentifierLength)
	if dirIDEnd > len(data) {
		return fmt.Errorf("directory identifier out of range: end=%d, data len=%d", dirIDEnd, len(data))
	}
	ptr.DirectoryIdentifier = string(data[8:dirIDEnd])

	// Handle padding
	ptr.Padding = nil
	if ptr.DirectoryIdentifierLength%2 != 0 {
		ptr.Padding = []byte{0}
	}

	// Single grouped logging call (TRACE level)
	ptr.logger.V(logging.TRACE).Info("PathTableRecord fields",
		"directoryIdentifierLength", ptr.DirectoryIdentifierLength,
		"extendedAttributeRecordLength", ptr.ExtendedAttributeRecordLength,
		"locationOfExtent", ptr.LocationOfExtent,
		"parentDirectoryNumber", ptr.ParentDirectoryNumber,
		"directoryIdentifier", ptr.DirectoryIdentifier,
		"paddingLength", len(ptr.Padding),
		"paddingHex", fmt.Sprintf("%x", ptr.Padding),
	)

	return nil
}
