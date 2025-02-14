package descriptor

import (
	"github.com/bgrewell/iso-kit/pkg/consts"
	"io"
)

type SupplementaryVolumeDescriptor struct {
	VolumeDescriptorHeader
	SupplementaryVolumeDescriptorBody
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
	// Application Use field is reserved for application use.
	ApplicationUse [consts.ISO9660_APPLICATION_USE_SIZE]byte `json:"application_use"`
}

func (d *SupplementaryVolumeDescriptor) Marshal() ([consts.ISO9660_SECTOR_SIZE]byte, error) {
	return [consts.ISO9660_SECTOR_SIZE]byte{}, nil
}

func (d *SupplementaryVolumeDescriptor) Unmarshal(data [consts.ISO9660_SECTOR_SIZE]byte, isoFile io.ReaderAt) error {
	return nil
}
