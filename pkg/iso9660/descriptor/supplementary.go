package descriptor

import (
	"encoding/binary"
	"fmt"
	"github.com/bgrewell/iso-kit/pkg/helpers"
	"github.com/bgrewell/iso-kit/pkg/iso9660/consts"
	"github.com/bgrewell/iso-kit/pkg/iso9660/directory"
	"github.com/bgrewell/iso-kit/pkg/iso9660/encoding"
	"strings"
	"time"
)

// SUPPLEMENTARY_VOLUME_DESCRIPTOR_BODY_SIZE is the total size (in bytes) of the SVD body according to ISO9660.
const SUPPLEMENTARY_VOLUME_DESCRIPTOR_BODY_SIZE = consts.ISO9660_SECTOR_SIZE - consts.ISO9660_VOLUME_DESC_HEADER_SIZE

type SupplementaryVolumeDescriptor struct {
	VolumeDescriptorHeader
	SupplementaryVolumeDescriptorBody
}

// Marshal converts the entire SupplementaryVolumeDescriptor into a 2048-byte on-disk sector.
func (d *SupplementaryVolumeDescriptor) Marshal() ([consts.ISO9660_SECTOR_SIZE]byte, error) {
	var sector [consts.ISO9660_SECTOR_SIZE]byte
	offset := 0

	// Marshal the header (first 7 bytes).
	headerBytes, err := d.VolumeDescriptorHeader.Marshal()
	if err != nil {
		return sector, fmt.Errorf("failed to marshal header: %w", err)
	}
	copy(sector[0:7], headerBytes[:])
	offset += 7

	// Marshal the body.
	bodyBytes, err := d.SupplementaryVolumeDescriptorBody.Marshal()
	if err != nil {
		return sector, fmt.Errorf("failed to marshal body: %w", err)
	}
	copy(sector[offset:offset+SUPPLEMENTARY_VOLUME_DESCRIPTOR_BODY_SIZE], bodyBytes[:])
	offset += SUPPLEMENTARY_VOLUME_DESCRIPTOR_BODY_SIZE

	// Pad any remaining bytes with zeros (should be none if the total is 2048).
	for i := offset; i < consts.ISO9660_SECTOR_SIZE; i++ {
		sector[i] = 0
	}

	return sector, nil
}

// Unmarshal parses a 2048-byte sector into the SupplementaryVolumeDescriptor.
func (d *SupplementaryVolumeDescriptor) Unmarshal(data [consts.ISO9660_SECTOR_SIZE]byte) error {
	offset := 0

	// Unmarshal the header (first 7 bytes).
	var headerBytes [7]byte
	copy(headerBytes[:], data[0:7])
	if err := d.VolumeDescriptorHeader.Unmarshal(headerBytes); err != nil {
		return fmt.Errorf("failed to unmarshal header: %w", err)
	}
	offset += 7

	// Unmarshal the body.
	bodyData := data[offset : offset+SUPPLEMENTARY_VOLUME_DESCRIPTOR_BODY_SIZE]
	if err := d.SupplementaryVolumeDescriptorBody.Unmarshal(bodyData); err != nil {
		return fmt.Errorf("failed to unmarshal body: %w", err)
	}

	// Remaining bytes (if any) are reserved; ignore.
	return nil
}

type SupplementaryVolumeDescriptorBody struct {
	// Volume Flags only has 1 currently used bit field, bit 0, which if set to 0 means that the Escape Sequences field
	// specifies only escape sequences registered according to ISO 2375. If set to 1 it means that the Escape Sequences
	// field specifies at least one escape sequence not registered according to ISO 2375
	VolumeFlags byte `json:"volume_flags"`
	// System Identifier specifies a system which can recognize and act upon the content of the Logical Sectors within
	// logical Sector Numbers 0 to 15 of the volume.
	//  | (a1-characters)
	SystemIdentifier string `json:"system_identifier"`
	// Volume Identifier specifies an identification of the volume
	//  | (d1-characters)
	VolumeIdentifier string `json:"volume_identifier"`
	// Unused Field is a block of unused bytes from BP73-80 and should contain only 0x00 bytes
	UnusedField1 [8]byte `json:"unused_field_1"`
	// Volume Space Size is a 8 byte field that the spec doesn't seem to address how it's used. (Table 6 of ECMA-199
	// incorrectly lists this as 32 bytes)
	VolumeSpaceSize [8]byte `json:"volume_space_size"`
	// Escape Sequences specifies one or more escape sequences according to ISO 2022 that designate the G0 graphic
	// character set and, optionally, the G1 graphic character set to be used in an 8-bit environment according to
	// ISO 2022 to interpret descriptor fields related to the Directory Hierarchy identified by this Volume Descriptor
	// (see 7.4.4). If the G1 set is designated, it is implicitly invoked into columns 10 to 15 of the code table.
	// These escape sequences shall conform to ECMA-35, except that the ESCAPE character shall be omitted
	// from each designating escape sequence when recorded in this field. The first or only escape sequence shall
	// begin at the first byte of the field. Each successive escape sequence shall begin at the byte in the field
	// 32 immediately following the last byte of the preceding escape sequence. Any unused byte positions following
	// the last sequence shall be set to (00). If all the bytes of this field are set to (00), it shall mean that the
	// set of a1-characters is identical with the set of a-characters and that the set of d1-characters is identical
	// with the set of d-characters. In this case both sets are coded according to ECMA-6.
	EscapeSequences [32]byte `json:"escape_sequences"`
	// Volume Set Size is a numerical value that is 4 bytes in size and is not addressed with regard to usage in the
	// spec. Note: look into more, probably a uint16 stored as both little and big endian.
	VolumeSetSize [4]byte `json:"volume_set_size"`
	// Volume Sequence Number is a numerical value that is 4 bytes in size and is not addressed with regard to usage in
	// the spec. Note: look into more, probably a uint16 stored as both little and big endian.
	VolumeSequenceNumber [4]byte `json:"volume_sequence_number"`
	// Logical Block Size is a numerical value that is 4 bytes in size and is not addressed with regard to usage in the
	// spec. Note: look into more, probably a uint16 stored as both little and big endian.
	LogicalBlockSize [4]byte `json:"logical_block_size"`
	// Path Table Size specifies the length in bytes of a recorded occurrence of the Path Table identified by this
	// Volume Descriptor.
	//  | Encoding: BothByteOrder
	PathTableSize uint32 `json:"path_table_size"`
	// This field shall specify as a 32-bit number the Logical Block Number of the first Logical Block allocated to the
	// Extent which contains an occurrence of the Path Table.
	//  | Encoding: LittleEndian
	LocationOfTypeLPathTable uint32 `json:"location_type_of_l_path_table"`
	// This field shall specify as a 32-bit number the Logical Block Number of the first Logical Block allocated to the
	// Extent which contains an optional occurrence of the Path Table. If the value is 0, it shall mean that the Extent
	// shall not be expected to have been recorded.
	//  | Encoding: LittleEndian
	LocationOfOptionalTypeLPathTable uint32 `json:"location_of_optional_type_l_path_table"`
	// This field shall specify as a 32-bit number the Logical Block Number of the first Logical Block allocated to the
	// Extent which contains an occurrence of the Path Table.
	//  | Encoding: BigEndian
	LocationOfTypeMPathTable uint32 `json:"location_of_m_path_table"`
	// This field shall specify as a 32-bit number the Logical Block Number of the first Logical Block allocated to the
	// Extent which contains an optional occurrence of the Path Table. If the value is 0, it shall mean that the Extent
	// shall not be expected to have been recorded.
	//  | Encoding: BigEndian
	LocationOfOptionalTypeMPathTable uint32 `json:"location_of_optional_type_m_path_table"`
	// Root Director Record contains an occurrence of the Directory Record for the Root Directory.
	RootDirectoryRecord *directory.DirectoryRecord `json:"root_directory_record"`
	// Volume Set Identifier specifies an identification of the Volume Set of which the volume is a member.
	//  | (d1-characters)
	VolumeSetIdentifier string `json:"volume_set_identifier"`
	// Publisher Identifier specifies an identification of the user who specified what shall be recorded on the
	// Volume Group of which the volume is a member. If the first byte is set to (5F), the remaining bytes of this field
	// shall specify an identifier for a file containing the identification of the user. This file shall be described in
	// the Root Directory. If all bytes of this field are set to (FILLER), it shall mean that no such user is
	// identified. Within a Supplementary Volume Descriptor, the characters in this field shall be a1-characters. Within
	// an Enhanced Volume Descriptor, the characters in this field shall be subject to agreement between the originator
	// and recipient of the volume
	PublisherIdentifier string `json:"publisher_identifier"`
	// Data Preparer Identifier specifies an identification of the person or other entity which controls the
	// preparation of the data to be recorded on the Volume Group of which the volume is a member. If the first byte is
	// set to (5F), the remaining bytes of this field shall specify an identifier for a file containing the
	// identification of the data preparer. This file shall be described in the Root Directory. If all bytes of this
	// field are set to (FILLER), it shall mean that no such data preparer is identified. Within a Supplementary Volume
	// Descriptor, the characters in this field shall be a1-characters. Within an Enhanced Volume Descriptor, the
	// characters in this field shall be subject to agreement between the originator and recipient of the volume.
	DataPreparerIdentifier string `json:"data_preparer_identifier"`
	// Application Identifier specifies an identification of the specification of how the data are recorded on the
	// Volume Group of which the volume is a member. If the first byte is set to (5F), the remaining bytes of this field
	// shall specify an identifier for a file containing the identification of the application. This file shall be
	// described in the Root Directory. If all bytes of this field are set to (FILLER), it shall mean that no such
	// application is identified. Within a Supplementary Volume Descriptor, the characters in this field shall be
	// a1-characters. Within an Enhanced Volume Descriptor, the characters in this field shall be subject to agreement
	// between the originator and recipient of the volume.
	ApplicationIdentifier string `json:"application_identifier"`
	// Copyright File Identifier specifies an identification for a file described by the Root Directory and containing a copyright
	// statement for those volumes of the Volume Set the sequence numbers of which are less than, or equal to, the
	// assigned Volume Set size of the volume. If all bytes of this field are set to (FILLER), it shall mean that no
	// such file is identified. Within a Supplementary Volume Descriptor, the characters in this field shall be
	// d1-characters, SEPARATOR 1 and SEPARATOR 2. Within an Enhanced Volume Descriptor, the characters in this field
	// shall be subject to agreement between the originator and recipient of the volume.
	CopyrightFileIdentifier string `json:"copyright_file_identifier"`
	// Abstract File Identifier specifies an identification for a file described by the Root Directory and containing an
	// abstract statement for those volumes of the Volume Set the sequence numbers of which are less than, or equal to,
	// the assigned Volume Set size of the volume. If all bytes of this field are set to (FILLER), it shall mean that no
	// such file is identified. Within a Supplementary Volume Descriptor, the characters in this field shall be
	// d1-characters, SEPARATOR 1 and SEPARATOR 2. Within an Enhanced Volume Descriptor, the characters in this field
	// shall be subject to agreement between the originator and recipient of the volume.
	AbstractFileIdentifier string `json:"abstract_file_identifier"`
	// Bibliographic File Identifier specifies an identification for a file described by the Root Directory and
	// containing bibliographic records interpreted according to standards that are the subject of an agreement between
	// the originator and the recipient of the volume. If all bytes of this field are set to (FILLER), it shall mean
	// that no such file is identified. Within a Supplementary Volume Descriptor, the characters in this field shall be
	// d1-characters, SEPARATOR 1 and SEPARATOR 2. Within an Enhanced Volume Descriptor, the characters in this field
	// shall be subject to agreement between the originator and recipient of the volume.
	BibliographicFileIdentifier string `json:"bibliographic_file_identifier"`
	// Volume Creation Date and Time specifies the date and time of the creation of the volume.
	//  | Encoding: 17-byte time and date format
	VolumeCreationDateAndTime time.Time `json:"volume_creation_date_and_time"`
	// Volume Modification Date and Time specifies the date and time of the most recent modification of the volume.
	//  | Encoding: 17-byte time and date format
	VolumeModificationDateAndTime time.Time `json:"volume_modification_date_and_time"`
	// Volume Expiration Date and Time specifies the date and time after which the volume is considered to be
	// obsolete.
	//  | Encoding: 17-byte time and date format
	VolumeExpirationDateAndTime time.Time `json:"volume_expiration_date_and_time"`
	// Volume Effective Date and Time specifies the date and time after which the volume may be used.
	//  | Encoding: 17-byte time and date format
	VolumeEffectiveDateAndTime time.Time `json:"volume_effective_date_and_time"`
	// File Structure Version specifies the version of the Directory Records specified in this Volume Descriptor.
	FileStructureVersion uint8 `json:"file_structure_version"`
	// Reserved Field 1. Value should be 0x00
	ReservedField1 byte `json:"reserved_field_1"`
	// Application Use field is reserved for application use.
	ApplicationUse [consts.ISO9660_APPLICATION_USE_SIZE]byte `json:"application_use"`
	// Reserved Field 2. Values should all be 0x00
	ReservedField2 [653]byte `json:"reserved_field_2"`
}

// Marshal converts the SupplementaryVolumeDescriptorBody into its fixed‑size on‑disk representation.
func (svdb *SupplementaryVolumeDescriptorBody) Marshal() ([SUPPLEMENTARY_VOLUME_DESCRIPTOR_BODY_SIZE]byte, error) {
	var data [SUPPLEMENTARY_VOLUME_DESCRIPTOR_BODY_SIZE]byte
	offset := 0

	// 1. VolumeFlags: 1 byte.
	data[offset] = svdb.VolumeFlags
	offset++

	// 2. SystemIdentifier: 32 bytes.
	sysID := helpers.PadString(svdb.SystemIdentifier, 32)
	copy(data[offset:offset+32], sysID)
	offset += 32

	// 3. VolumeIdentifier: 32 bytes.
	volID := helpers.PadString(svdb.VolumeIdentifier, 32)
	copy(data[offset:offset+32], volID)
	offset += 32

	// 4. UnusedField1: 8 bytes.
	copy(data[offset:offset+8], svdb.UnusedField1[:])
	offset += 8

	// 5. VolumeSpaceSize: 8 bytes.
	copy(data[offset:offset+8], svdb.VolumeSpaceSize[:])
	offset += 8

	// 6. EscapeSequences: 32 bytes.
	copy(data[offset:offset+32], svdb.EscapeSequences[:])
	offset += 32

	// 7. VolumeSetSize: 4 bytes
	copy(data[offset:offset+4], svdb.VolumeSetSize[:])
	offset += 4

	// 8. VolumeSequenceNumber: 4 bytes
	copy(data[offset:offset+4], svdb.VolumeSequenceNumber[:])
	offset += 4

	// 9. LogicalBlockSize: 4 bytes
	copy(data[offset:offset+4], svdb.LogicalBlockSize[:])
	offset += 4

	// 10. PathTableSize: 8 bytes (both-byte orders for uint32).
	ptsBytes := encoding.MarshalBothByteOrders32(svdb.PathTableSize)
	copy(data[offset:offset+8], ptsBytes[:])
	offset += 8

	// 11. LocationOfTypeLPathTable: 4 bytes, little-endian.
	binary.LittleEndian.PutUint32(data[offset:offset+4], svdb.LocationOfTypeLPathTable)
	offset += 4

	// 12. LocationOfOptionalTypeLPathTable: 4 bytes, little-endian.
	binary.LittleEndian.PutUint32(data[offset:offset+4], svdb.LocationOfOptionalTypeLPathTable)
	offset += 4

	// 13. LocationOfTypeMPathTable: 4 bytes, big-endian.
	binary.BigEndian.PutUint32(data[offset:offset+4], svdb.LocationOfTypeMPathTable)
	offset += 4

	// 14. LocationOfOptionalTypeMPathTable: 4 bytes, big-endian.
	binary.BigEndian.PutUint32(data[offset:offset+4], svdb.LocationOfOptionalTypeMPathTable)
	offset += 4

	// 15. RootDirectoryRecord: 34 bytes.
	if svdb.RootDirectoryRecord == nil {
		return data, fmt.Errorf("RootDirectoryRecord is nil")
	}
	rdBytes, err := svdb.RootDirectoryRecord.Marshal()
	if err != nil {
		return data, fmt.Errorf("failed to marshal RootDirectoryRecord: %w", err)
	}
	if len(rdBytes) != 34 {
		return data, fmt.Errorf("expected 34 bytes for RootDirectoryRecord, got %d", len(rdBytes))
	}
	copy(data[offset:offset+34], rdBytes)
	offset += 34

	// 16. VolumeSetIdentifier: 128 bytes.
	vsi := helpers.PadString(svdb.VolumeSetIdentifier, 128)
	copy(data[offset:offset+128], vsi)
	offset += 128

	// 17. PublisherIdentifier: 128 bytes.
	pubID := helpers.PadString(svdb.PublisherIdentifier, 128)
	copy(data[offset:offset+128], pubID)
	offset += 128

	// 18. DataPreparerIdentifier: 128 bytes.
	dpID := helpers.PadString(svdb.DataPreparerIdentifier, 128)
	copy(data[offset:offset+128], dpID)
	offset += 128

	// 19. ApplicationIdentifier: 128 bytes.
	appID := helpers.PadString(svdb.ApplicationIdentifier, 128)
	copy(data[offset:offset+128], appID)
	offset += 128

	// 20. CopyrightFileIdentifier: 37 bytes.
	cfID := helpers.PadString(svdb.CopyrightFileIdentifier, 37)
	copy(data[offset:offset+37], cfID)
	offset += 37

	// 21. AbstractFileIdentifier: 37 bytes.
	afID := helpers.PadString(svdb.AbstractFileIdentifier, 37)
	copy(data[offset:offset+37], afID)
	offset += 37

	// 22. BibliographicFileIdentifier: 37 bytes.
	bfID := helpers.PadString(svdb.BibliographicFileIdentifier, 37)
	copy(data[offset:offset+37], bfID)
	offset += 37

	// 23. VolumeCreationDateAndTime: 17 bytes.
	vcdBytes, err := encoding.MarshalDateTime(svdb.VolumeCreationDateAndTime)
	if err != nil {
		return data, fmt.Errorf("failed to marshal VolumeCreationDateAndTime: %w", err)
	}
	copy(data[offset:offset+17], vcdBytes[:])
	offset += 17

	// 24. VolumeModificationDateAndTime: 17 bytes.
	vmdBytes, err := encoding.MarshalDateTime(svdb.VolumeModificationDateAndTime)
	if err != nil {
		return data, fmt.Errorf("failed to marshal VolumeModificationDateAndTime: %w", err)
	}
	copy(data[offset:offset+17], vmdBytes[:])
	offset += 17

	// 25. VolumeExpirationDateAndTime: 17 bytes.
	vedBytes, err := encoding.MarshalDateTime(svdb.VolumeExpirationDateAndTime)
	if err != nil {
		return data, fmt.Errorf("failed to marshal VolumeExpirationDateAndTime: %w", err)
	}
	copy(data[offset:offset+17], vedBytes[:])
	offset += 17

	// 26. VolumeEffectiveDateAndTime: 17 bytes.
	vefBytes, err := encoding.MarshalDateTime(svdb.VolumeEffectiveDateAndTime)
	if err != nil {
		return data, fmt.Errorf("failed to marshal VolumeEffectiveDateAndTime: %w", err)
	}
	copy(data[offset:offset+17], vefBytes[:])
	offset += 17

	// 27. FileStructureVersion: 1 byte.
	data[offset] = svdb.FileStructureVersion
	offset++

	// 28. ReservedField1: 1 byte.
	data[offset] = svdb.ReservedField1
	offset++

	// 29. ApplicationUse: fixed size.
	copy(data[offset:offset+len(svdb.ApplicationUse)], svdb.ApplicationUse[:])
	offset += len(svdb.ApplicationUse)

	// 30. ReservedField2: 653 bytes.
	copy(data[offset:offset+len(svdb.ReservedField2)], svdb.ReservedField2[:])
	offset += len(svdb.ReservedField2)

	if offset != SUPPLEMENTARY_VOLUME_DESCRIPTOR_BODY_SIZE {
		return data, fmt.Errorf("marshal error: expected offset %d, got %d", SUPPLEMENTARY_VOLUME_DESCRIPTOR_BODY_SIZE, offset)
	}

	return data, nil
}

// Unmarshal parses a SUPPLEMENTARY_VOLUME_DESCRIPTOR_BODY_SIZE‑byte slice into the SupplementaryVolumeDescriptorBody.
// Fixed‑width string fields have their trailing spaces trimmed.
func (svdb *SupplementaryVolumeDescriptorBody) Unmarshal(data []byte) error {
	if len(data) < SUPPLEMENTARY_VOLUME_DESCRIPTOR_BODY_SIZE {
		return fmt.Errorf("data too short: expected %d bytes, got %d", SUPPLEMENTARY_VOLUME_DESCRIPTOR_BODY_SIZE, len(data))
	}
	offset := 0

	// 1. VolumeFlags.
	svdb.VolumeFlags = data[offset]
	offset++

	// 2. SystemIdentifier: 32 bytes.
	svdb.SystemIdentifier = strings.TrimRight(string(data[offset:offset+32]), " ")
	offset += 32

	// 3. VolumeIdentifier: 32 bytes.
	svdb.VolumeIdentifier = strings.TrimRight(string(data[offset:offset+32]), " ")
	offset += 32

	// 4. UnusedField1: 8 bytes.
	copy(svdb.UnusedField1[:], data[offset:offset+8])
	offset += 8

	// 5. VolumeSpaceSize: 8 bytes.
	copy(svdb.VolumeSpaceSize[:], data[offset:offset+8])
	offset += 8

	// 6. EscapeSequences: 32 bytes.
	copy(svdb.EscapeSequences[:], data[offset:offset+32])
	offset += 32

	// 7. VolumeSetSize: 4 bytes.
	copy(svdb.VolumeSetSize[:], data[offset:offset+4])
	offset += 4

	// 8. VolumeSequenceNumber: 4 bytes.
	copy(svdb.VolumeSequenceNumber[:], data[offset:offset+4])
	offset += 4

	// 9. LogicalBlockSize: 4 bytes.
	copy(svdb.LogicalBlockSize[:], data[offset:offset+4])
	offset += 4

	// 10. PathTableSize: 8 bytes (both-byte orders for uint32; use little-endian value).
	var ptsBytes [8]byte
	copy(ptsBytes[:], data[offset:offset+8])
	pathTableSize, err := encoding.UnmarshalBothByteOrders32(ptsBytes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal PathTableSize: %w", err)
	}
	svdb.PathTableSize = pathTableSize
	offset += 8

	// 11. LocationOfTypeLPathTable: 4 bytes, little-endian.
	svdb.LocationOfTypeLPathTable = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	// 12. LocationOfOptionalTypeLPathTable: 4 bytes, little-endian.
	svdb.LocationOfOptionalTypeLPathTable = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	// 13. LocationOfTypeMPathTable: 4 bytes, big-endian.
	svdb.LocationOfTypeMPathTable = binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	// 14. LocationOfOptionalTypeMPathTable: 4 bytes, big-endian.
	svdb.LocationOfOptionalTypeMPathTable = binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	// 15. RootDirectoryRecord: 34 bytes.
	if svdb.RootDirectoryRecord == nil {
		svdb.RootDirectoryRecord = new(directory.DirectoryRecord)
	}
	if err := svdb.RootDirectoryRecord.Unmarshal(data[offset : offset+34]); err != nil {
		return fmt.Errorf("failed to unmarshal RootDirectoryRecord: %w", err)
	}
	offset += 34

	// 16. VolumeSetIdentifier: 128 bytes.
	svdb.VolumeSetIdentifier = strings.TrimRight(string(data[offset:offset+128]), " ")
	offset += 128

	// 17. PublisherIdentifier: 128 bytes.
	svdb.PublisherIdentifier = strings.TrimRight(string(data[offset:offset+128]), " ")
	offset += 128

	// 18. DataPreparerIdentifier: 128 bytes.
	svdb.DataPreparerIdentifier = strings.TrimRight(string(data[offset:offset+128]), " ")
	offset += 128

	// 19. ApplicationIdentifier: 128 bytes.
	svdb.ApplicationIdentifier = strings.TrimRight(string(data[offset:offset+128]), " ")
	offset += 128

	// 20. CopyrightFileIdentifier: 37 bytes.
	svdb.CopyrightFileIdentifier = strings.TrimRight(string(data[offset:offset+37]), " ")
	offset += 37

	// 21. AbstractFileIdentifier: 37 bytes.
	svdb.AbstractFileIdentifier = strings.TrimRight(string(data[offset:offset+37]), " ")
	offset += 37

	// 22. BibliographicFileIdentifier: 37 bytes.
	svdb.BibliographicFileIdentifier = strings.TrimRight(string(data[offset:offset+37]), " ")
	offset += 37

	// 23. VolumeCreationDateAndTime: 17 bytes.
	var vcdBytes [17]byte
	copy(vcdBytes[:], data[offset:offset+17])
	volCreation, err := encoding.UnmarshalDateTime(vcdBytes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal VolumeCreationDateAndTime: %w", err)
	}
	svdb.VolumeCreationDateAndTime = volCreation
	offset += 17

	// 24. VolumeModificationDateAndTime: 17 bytes.
	var vmdBytes [17]byte
	copy(vmdBytes[:], data[offset:offset+17])
	volMod, err := encoding.UnmarshalDateTime(vmdBytes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal VolumeModificationDateAndTime: %w", err)
	}
	svdb.VolumeModificationDateAndTime = volMod
	offset += 17

	// 25. VolumeExpirationDateAndTime: 17 bytes.
	var vedBytes [17]byte
	copy(vedBytes[:], data[offset:offset+17])
	volExp, err := encoding.UnmarshalDateTime(vedBytes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal VolumeExpirationDateAndTime: %w", err)
	}
	svdb.VolumeExpirationDateAndTime = volExp
	offset += 17

	// 26. VolumeEffectiveDateAndTime: 17 bytes.
	var vefBytes [17]byte
	copy(vefBytes[:], data[offset:offset+17])
	volEff, err := encoding.UnmarshalDateTime(vefBytes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal VolumeEffectiveDateAndTime: %w", err)
	}
	svdb.VolumeEffectiveDateAndTime = volEff
	offset += 17

	// 27. FileStructureVersion: 1 byte.
	svdb.FileStructureVersion = data[offset]
	offset++

	// 28. ReservedField1: 1 byte.
	svdb.ReservedField1 = data[offset]
	offset++

	// 29. ApplicationUse: fixed size.
	copy(svdb.ApplicationUse[:], data[offset:offset+len(svdb.ApplicationUse)])
	offset += len(svdb.ApplicationUse)

	// 30. ReservedField2: 653 bytes.
	copy(svdb.ReservedField2[:], data[offset:offset+len(svdb.ReservedField2)])
	offset += len(svdb.ReservedField2)

	if offset != SUPPLEMENTARY_VOLUME_DESCRIPTOR_BODY_SIZE {
		return fmt.Errorf("unmarshal error: expected offset %d, got %d", SUPPLEMENTARY_VOLUME_DESCRIPTOR_BODY_SIZE, offset)
	}
	return nil
}
