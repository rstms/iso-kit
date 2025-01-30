package path

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/bgrewell/iso-kit/pkg/logging"
	"github.com/go-logr/logr"
)

func NewPathTableRecord(logger logr.Logger) *PathTableRecord {
	return &PathTableRecord{logger: logger}
}

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
func (ptr *PathTableRecord) Unmarshal(data []byte) error {
	if len(data) < 8 {
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
		padEnd := dirIDEnd + 1
		if padEnd > len(data) {
			return fmt.Errorf("padding out of range: end=%d, data len=%d", padEnd, len(data))
		}
		// Make a copy of the padding slice
		ptr.Padding = append([]byte(nil), data[dirIDEnd:padEnd]...)
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
