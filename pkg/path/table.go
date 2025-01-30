package path

import (
	"encoding/binary"
	"errors"
	"github.com/bgrewell/iso-kit/pkg/logging"
)

type PathTableRecord struct {
	DirectoryIdentifierLength     byte   // Directory identifier length
	ExtendedAttributeRecordLength byte   // Extended attribute record length
	LocationOfExtent              uint32 // Location of extent
	ParentDirectoryNumber         uint16 // Parent directory number
	DirectoryIdentifier           string // Directory identifier
	Padding                       []byte // Padding to align record if identifier length is odd
}

func (ptr *PathTableRecord) Unmarshal(data []byte) error {
	if len(data) < 8 {
		return errors.New("invalid data length")
	}

	ptr.DirectoryIdentifierLength = data[0]
	logging.Logger().Tracef("Directory identifier length: %d", ptr.DirectoryIdentifierLength)
	ptr.ExtendedAttributeRecordLength = data[1]
	logging.Logger().Tracef("Extended attribute record length: %d", ptr.ExtendedAttributeRecordLength)
	ptr.LocationOfExtent = binary.LittleEndian.Uint32(data[2:6])
	logging.Logger().Tracef("Location of extent: %d", ptr.LocationOfExtent)
	ptr.ParentDirectoryNumber = binary.LittleEndian.Uint16(data[6:8])
	logging.Logger().Tracef("Parent directory number: %d", ptr.ParentDirectoryNumber)
	ptr.DirectoryIdentifier = string(data[8 : 8+ptr.DirectoryIdentifierLength])
	logging.Logger().Tracef("Directory identifier: %s", ptr.DirectoryIdentifier)
	if ptr.DirectoryIdentifierLength%2 != 0 {
		ptr.Padding = data[8+ptr.DirectoryIdentifierLength : 8+ptr.DirectoryIdentifierLength+1]
	}
	logging.Logger().Tracef("Padding len: %d, value: %x", len(ptr.Padding), ptr.Padding)

	return nil
}
