package descriptor

import (
	"github.com/bgrewell/iso-kit/pkg/consts"
	"io"
	"time"
)

const (
	// Reserved for future use field from BP 1396 to 2048
	PRIMARY_RESERVED_FIELD2_SIZE = 653
)

type PrimaryVolumeDescriptor struct {
	VolumeDescriptorHeader
	PrimaryVolumeDescriptorBody
}

type PrimaryVolumeDescriptorBody struct {
	// Unused byte should be set to 0x00.
	UnusedField1 byte `json:"unused_field_1"`
	// System Identifier specifies a system which can recognize and act upon the content of the Logical Sectors within
	// logical Sector Numbers 0 to 15 of the volume.
	//  | (a-characters)
	SystemIdentifier string `json:"system_identifier"`
	// Volume Identifier specifies an identification of the volume
	//  | (d-characters)
	VolumeIdentifier string `json:"volume_identifier"`
	// Unused all bytes should be set to 0x00
	UnusedField2 [8]byte `json:"unused_field_2"`
	// Volume Space Size is a field that specifies the number of logical blocks in which the Volume Space of the volume
	// is recorded as a 32-bit number.
	//  | Encoding: BothByteOrder
	VolumeSpaceSize uint32 `json:"volume_space_size"`
	// Unused all bytes should be set to 0x00
	UnusedField3 [32]byte `json:"unused_field_3"`
	// Volume Set Size is a field that specifies the assigned Volume Set size of the volume as a 16-bit number.
	//  | Encoding: BothByteOrder
	VolumeSetSize uint16 `json:"volume_set_size"`
	// Volume Sequence Number is a field that represents the ordinal number of the volume in the Volume Set.
	//  | Encoding: BothByteOrder
	VolumeSequenceNumber uint16 `json:"volume_sequence_number"`
	// Logical Block Size specifies the size in bytes of a logical block
	//  | Encoding: BothByteOrder
	LogicalBlockSize uint16 `json:"logical_block_size"`
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
	RootDirectoryRecord *DirectoryRecord `json:"root_directory_record"`
	// Volume Set Identifier specifies an identification of the Volume Set of which the volume is a member.
	//  | (d-characters)
	VolumeSetIdentifier string `json:"volume_set_identifier"`
	// Publisher Identifier specifies an identification of the user who specified what shall be recorded on the
	// Volume Group of which the volume is a member. If the first byte is set to (5F), the remaining bytes of this field
	// shall specify an identifier for a file containing the identification of the user. This file shall be described in
	// the Root Directory. The File Name shall not contain more than eight d-characters and the File Name Extension
	// shall not contain more than three d-characters. All 'ISO9660_FILLER' means there is no identifier.
	PublisherIdentifier string `json:"publisher_identifier"`
	// Data Preparer Identifier specifies an identification of the person or other entity which controls the
	// preparation of the data to be recorded on the Volume Group of which the volume is a member. If the first byte is
	// set to (5F), the remaining bytes of this field shall specify an identifier for a file containing the
	// identification of the data preparer. This file shall be described in the Root Directory. The File Name shall not
	// contain more than eight d-characters and the File Name Extension shall not contain more than three d-characters.
	// All 'ISO9660_FILLER' means there is no identifier.
	DataPreparerIdentifier string `json:"data_preparer_identifier"`
	// Application Identifier specifies an identification of the specification of how the data are recorded on the
	// Volume Group of which the volume is a member. If the first byte is set to (5F), the remaining bytes of this field
	// shall specify an identifier for a file containing the identification of the application. This file shall be
	// described in the Root Directory. The File Name shall not contain more than eight d-characters and the File Name
	// Extension shall not contain more than three d-characters. All 'ISO9660_FILLER' means there is no identifier.
	ApplicationIdentifier string `json:"application_identifier"`
	// Copyright File Identifier specifies an identification for a file described by the Root Directory and containing a copyright
	// statement for those volumes of the Volume Set the sequence numbers of which are less than, or equal to, the
	// assigned Volume Set size of the volume. If all bytes of this field are set to (FILLER), it shall mean that no
	// such file is identified. The File Name of a Copyright File Identifier shall not contain more than eight
	// d-characters. The File Name Extension of a Copyright File Identifier shall not contain more than three
	// d-characters. Allowed characters are d-characters, SEPARATOR 1 and SEPARATOR 2
	CopyrightFileIdentifier string `json:"copyright_file_identifier"`
	// Abstract File Identifier specifies an identification for a file described by the Root Directory and containing an
	// abstract statement for those volumes of the Volume Set the sequence numbers of which are less than, or equal to,
	// the assigned Volume Set size of the volume. If all bytes of this field are set to (FILLER), it shall mean that no
	// such file is identified. The File Name of an Abstract File Identifier shall not contain more than eight
	// d-characters. The File Name Extension of an Abstract File Identifier shall not contain more than three
	// d-characters. The characters in this field shall be d-characters, SEPARATOR 1 and SEPARATOR 2.
	AbstractFileIdentifier string `json:"abstract_file_identifier"`
	// Bibliographic File Identifier specifies an identification for a file described by the Root Directory and
	// containing bibliographic records interpreted according to standards that are the subject of an agreement between
	// the originator and the recipient of the volume. If all bytes of this field are set to (FILLER), it shall mean
	// that no such file is identified. The File Name of a Bibliographic File Identifier shall not contain more than
	// eight d-characters. The File Name Extension of a Bibliographic File Identifier shall not contain more than three
	// d-characters. The characters in this field shall be d-characters, SEPARATOR 1 and SEPARATOR 2.
	BibliographicFileIdentifier string `json:"bibliographic_file_identifier"`
	// Volume Creation Date and Time specifies the date and the time of the day at which the information in the
	// volume was created.
	//  | 8.4.26.1 Date and Time Format
	VolumeCreationDateAndTime time.Time `json:"volume_creation_date_and_time"`
	// Volume Modification Date and Time specifies the date and the time of the day at which the information in the
	// volume was last modified.
	//  | 8.4.26.1 Date and Time Format
	VolumeModificationDateAndTime time.Time `json:"volume_modification_date_and_time"`
	// Volume Expiration Date and Time specifies the date and the time of the day at which the information in the
	// volume may be regarded as obsolete. If the date and time are not specified then the information shall not be
	// regarded as obsolete.
	//  | 8.4.26.1 Date and Time Format
	VolumeExpirationDateAndTime time.Time `json:"volume_expiration_date_and_time"`
	// Volume Effective Date and Time specifies the date and the time of the day at which the information in the volume
	// may be used. If the date and time are not specified then the information may be used at once.
	//  | 8.4.26.1 Date and Time Format
	VolumeEffectiveDateAndTime time.Time `json:"volume_effective_date_and_time"`
	// File Structure Version specifies as an 8-bit number the version of the specification of the records of a
	// directory and of a Path Table. For a Primary Volume Descriptor or for a Supplementary Volume Descriptor, 1 shall
	// indicate the structure of this Standard. For an Enhanced Volume Descriptor, 2 shall indicate the structure of
	// this Standard.
	FileStructureVersion uint8 `json:"file_structure_version"`
	// Reserved Field 1 is unused and should be set to 0x00.
	ReservedField1 byte `json:"reserved_field_1"`
	// Application Use field is reserved for application use.
	ApplicationUse [consts.ISO9660_APPLICATION_USE_SIZE]byte `json:"application_use"`
	// Reserved Field 2 is unused and all bytes should be set to 0x00.
	ReservedField2 [PRIMARY_RESERVED_FIELD2_SIZE]byte `json:"reserved_field_2"`
}

func (d *PrimaryVolumeDescriptor) Marshal() ([consts.ISO9660_SECTOR_SIZE]byte, error) {
	return [consts.ISO9660_SECTOR_SIZE]byte{}, nil
}

func (d *PrimaryVolumeDescriptor) Unmarshal(data [consts.ISO9660_SECTOR_SIZE]byte, isoFile io.ReaderAt) error {
	return nil
}
