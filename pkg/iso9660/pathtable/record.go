package pathtable

import (
	"encoding/binary"
	"fmt"
	"github.com/bgrewell/iso-kit/pkg/consts"
	"io"
)

func NewPathTable(reader io.ReaderAt, location uint32, size int, littleEndian bool) (*PathTable, error) {
	data := make([]byte, size)
	_, err := reader.ReadAt(data, int64(location)*consts.ISO9660_SECTOR_SIZE)
	if err != nil {
		return nil, fmt.Errorf("failed to read path table: %w", err)
	}

	pt := &PathTable{}
	offset := 0

	for offset < len(data) {
		record := &PathTableRecord{}
		if err := record.Unmarshal(data[offset:], littleEndian); err != nil {
			return nil, err
		}
		pt.Records = append(pt.Records, record)

		// Move to the next record
		recordLen := int(record.LengthOfDirectoryIdentifier) + 8
		if recordLen%2 != 0 {
			recordLen++ // Handle padding byte
		}
		offset += recordLen
	}

	return pt, nil
}

// PathTable represents a full path table, containing multiple records.
type PathTable struct {
	Records []*PathTableRecord
}

// Marshal converts a PathTable into a contiguous byte array.
func (pt *PathTable) Marshal(littleEndian bool) ([]byte, error) {
	var buf []byte

	for _, record := range pt.Records {
		recBytes, err := record.Marshal(littleEndian)
		if err != nil {
			return nil, err
		}
		buf = append(buf, recBytes...)
	}

	return buf, nil
}

type PathTableRecord struct {
	// Length of Directory Identifier specifies the length in bytes of the Directory Identifier field of the Path Table
	// Record.
	LengthOfDirectoryIdentifier uint8 `json:"length_of_directory_identifier"`
	// Extended Attribute Record Length specifies the Extended Attribute Record length if an Extended Attribute Record
	// is recorded. Otherwise, this number will be zero.
	ExtendedAttributeRecordLength uint8 `json:"extended_atrribute_record_length"`
	// Location of Extent specifies the Logical Block Number of the first Logical Block allocated to the Extent in which
	// the directory is recorded.
	LocationOfExtent uint32 `json:"location_of_extent"`
	// Parent Directory Number specifies the record number in the Path Table for the parent directory of the directory.
	ParentDirectoryNumber uint16 `json:"parent_directory_number"`
	// Directory Identifier specifies the identification for the directory. The characters in this field shall be
	// d-characters or d1-characters or only a null byte 0x00.
	DirectoryIdentifier string `json:"directory_identifier"`
	// Padding Field is only present in the Path Table Record if the number in the Length Of Directory Identifier field
	// is an odd number. If present, this field shall be set to a null byte 0x00.
	// Note: The padding field isn't actually represented in this struct since it's presence or absence is simply
	// calculated when marshalling to an array of bytes. When unmarshalling if the LengthOfFileIdentifier field is even
	// then we make sure we skip the padding byte when we continue processing the following fields.
	// Padding *byte `json:"padding" ----------
}

// Marshal converts a single PathTableRecord into a byte slice.
func (ptr *PathTableRecord) Marshal(littleEndian bool) ([]byte, error) {
	dirIDBytes := []byte(ptr.DirectoryIdentifier)
	ptr.LengthOfDirectoryIdentifier = uint8(len(dirIDBytes))

	recordLen := 8 + len(dirIDBytes)
	if len(dirIDBytes)%2 != 0 {
		recordLen++
	}

	buf := make([]byte, recordLen)
	offset := 0

	buf[offset] = ptr.LengthOfDirectoryIdentifier
	offset++
	buf[offset] = ptr.ExtendedAttributeRecordLength
	offset++

	if littleEndian {
		binary.LittleEndian.PutUint32(buf[offset:], ptr.LocationOfExtent)
	} else {
		binary.BigEndian.PutUint32(buf[offset:], ptr.LocationOfExtent)
	}
	offset += 4

	if littleEndian {
		binary.LittleEndian.PutUint16(buf[offset:], ptr.ParentDirectoryNumber)
	} else {
		binary.BigEndian.PutUint16(buf[offset:], ptr.ParentDirectoryNumber)
	}
	offset += 2

	copy(buf[offset:], dirIDBytes)
	offset += len(dirIDBytes)

	if len(dirIDBytes)%2 != 0 {
		buf[offset] = 0x00 // Padding byte
	}

	return buf, nil
}

// Unmarshal decodes a single PathTableRecord from a byte slice.
func (ptr *PathTableRecord) Unmarshal(data []byte, littleEndian bool) error {
	if len(data) < 8 {
		return fmt.Errorf("data too short to contain a PathTableRecord")
	}
	offset := 0

	ptr.LengthOfDirectoryIdentifier = data[offset]
	offset++
	ptr.ExtendedAttributeRecordLength = data[offset]
	offset++

	if littleEndian {
		ptr.LocationOfExtent = binary.LittleEndian.Uint32(data[offset:])
	} else {
		ptr.LocationOfExtent = binary.BigEndian.Uint32(data[offset:])
	}
	offset += 4

	if littleEndian {
		ptr.ParentDirectoryNumber = binary.LittleEndian.Uint16(data[offset:])
	} else {
		ptr.ParentDirectoryNumber = binary.BigEndian.Uint16(data[offset:])
	}
	offset += 2

	n := int(ptr.LengthOfDirectoryIdentifier)
	if len(data) < offset+n {
		return fmt.Errorf("data too short for DirectoryIdentifier")
	}
	ptr.DirectoryIdentifier = string(data[offset : offset+n])
	offset += n

	if n%2 != 0 {
		offset++ // Skip padding byte
	}

	return nil
}
