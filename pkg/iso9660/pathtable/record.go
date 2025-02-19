package pathtable

import (
	"encoding/binary"
	"fmt"
)

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

// Marshal converts the PathTableRecord into its on‑disk byte representation.
func (ptr *PathTableRecord) Marshal() ([]byte, error) {
	// Convert the DirectoryIdentifier to bytes.
	dirIDBytes := []byte(ptr.DirectoryIdentifier)
	// Set the LengthOfDirectoryIdentifier field (should equal len(dirIDBytes)).
	ptr.LengthOfDirectoryIdentifier = uint8(len(dirIDBytes))

	// Compute the total record length:
	// Fixed fields: 1 + 1 + 4 + 2 = 8 bytes,
	// plus DirectoryIdentifier bytes,
	// plus 1 padding byte if the identifier length is odd.
	recordLen := 8 + len(dirIDBytes)
	if len(dirIDBytes)%2 != 0 {
		recordLen++
	}

	buf := make([]byte, 0, recordLen)

	// 1. LengthOfDirectoryIdentifier (1 byte)
	buf = append(buf, ptr.LengthOfDirectoryIdentifier)

	// 2. ExtendedAttributeRecordLength (1 byte)
	buf = append(buf, ptr.ExtendedAttributeRecordLength)

	// 3. LocationOfExtent (4 bytes, little‑endian)
	locBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(locBytes, ptr.LocationOfExtent)
	buf = append(buf, locBytes...)

	// 4. ParentDirectoryNumber (2 bytes, little‑endian)
	parentBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(parentBytes, ptr.ParentDirectoryNumber)
	buf = append(buf, parentBytes...)

	// 5. DirectoryIdentifier (variable length)
	buf = append(buf, dirIDBytes...)

	// 6. Padding Field: if the Directory Identifier length is odd, append one null byte.
	if len(dirIDBytes)%2 != 0 {
		buf = append(buf, 0x00)
	}

	return buf, nil
}

// Unmarshal decodes a PathTableRecord from the given byte slice.
// It expects the data slice to contain exactly one record.
func (ptr *PathTableRecord) Unmarshal(data []byte) error {
	// At minimum, we need the fixed 8 bytes.
	if len(data) < 8 {
		return fmt.Errorf("data too short to contain a PathTableRecord: %d bytes", len(data))
	}
	offset := 0

	// 1. LengthOfDirectoryIdentifier (1 byte)
	ptr.LengthOfDirectoryIdentifier = data[offset]
	offset++

	// 2. ExtendedAttributeRecordLength (1 byte)
	ptr.ExtendedAttributeRecordLength = data[offset]
	offset++

	// 3. LocationOfExtent (4 bytes, little‑endian)
	if len(data) < offset+4 {
		return fmt.Errorf("data too short for LocationOfExtent")
	}
	ptr.LocationOfExtent = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	// 4. ParentDirectoryNumber (2 bytes, little‑endian)
	if len(data) < offset+2 {
		return fmt.Errorf("data too short for ParentDirectoryNumber")
	}
	ptr.ParentDirectoryNumber = binary.LittleEndian.Uint16(data[offset : offset+2])
	offset += 2

	// 5. DirectoryIdentifier (LengthOfDirectoryIdentifier bytes)
	n := int(ptr.LengthOfDirectoryIdentifier)
	if len(data) < offset+n {
		return fmt.Errorf("data too short for DirectoryIdentifier: need %d, got %d", n, len(data)-offset)
	}
	ptr.DirectoryIdentifier = string(data[offset : offset+n])
	offset += n

	// 6. Padding Field: present if the Directory Identifier length is odd.
	if n%2 != 0 {
		if len(data) < offset+1 {
			return fmt.Errorf("data too short for padding byte")
		}
		if data[offset] != 0x00 {
			return fmt.Errorf("expected padding byte 0x00, got 0x%02X", data[offset])
		}
		offset++
	}

	// (Optionally, you might check that offset == len(data) to ensure there’s no trailing garbage.)
	return nil
}
