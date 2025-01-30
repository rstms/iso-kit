package rockridge

import (
	"fmt"
	"github.com/bgrewell/iso-kit/pkg/encoding"
	"io/fs"
	"os"
)

const (
	ROCK_RIDGE_IDENTIFIER = "RRIP_1991A"
	ROCK_RIDGE_VERSION    = 1
)

type RockRidgeEntryType string

const (
	POSIX_FILE_PERMS RockRidgeEntryType = "PX" //POSIX file permissions (owner, group, other)
	POSIX_DEVICE_NUM RockRidgeEntryType = "PN" //Device numbers for block/character device nodes (major/minor)
	SYMBOLIC_LINK    RockRidgeEntryType = "SL" //Symbolic link data (path components,flags)
	ALTERNATE_NAME   RockRidgeEntryType = "NM" //AlternateName (used for long filenames, case preservation, etc.)
	CHILD_LINK       RockRidgeEntryType = "CL" //ChildLink (used for directory relocation chains)
	PARENT_LINK      RockRidgeEntryType = "PL" //ParentLink (links a relocated directory back to its parent)
	RELOCATED_DIR    RockRidgeEntryType = "RE" //Marks a directory that has been relocated
	TIME_STAMPS      RockRidgeEntryType = "TF" //Time stamp information (creation, modification, access,etc)
	SPARSE_FILE      RockRidgeEntryType = "SF" //Sparse file information (less commonly used)
	ROCK_RIDGE       RockRidgeEntryType = "RR" //An older “Rock Ridge” extension signature (now typically replaced by ER).
)

type RockRidgeTimestamps []byte

type RockRidgeNameEntry struct {
	Continue  bool // Bit 0: Alternate Name continues in the next "NM" entry
	Current   bool // Bit 1: Alternate Name refers to the current directory ("." in POSIX)
	Parent    bool // Bit 2: Alternate Name refers to the parent directory (".." in POSIX)
	Reserved1 bool // Bit 3: Reserved, should be ZERO
	Reserved2 bool // Bit 4: Reserved, should be ZERO
	Reserved3 bool // Bit 5: Historically contains the network node name
	Reserved4 bool // Bit 6: Unused, reserved for future use
	Reserved5 bool // Bit 7: Unused, reserved for future use
	Name      string
}

// NM Entry Details
//
//		  Offset 0-1: Signature Word - "NM"
//		  Offset 2:   Length (LEN_NM) - 8-bit number. The number in this field shall be 5 plus the length of the Name Content record. If bit position 1, 2, or 5 of the "NM" Flags is set to 1, the value of this field shall be 5 and no Name Content shall be recorded.
//		  Offset 3:   System Use Entry Version - 8-bit number. The value of this field shall be 1.
//		  Offset 4:   Flags (NM Flags) - 8-bit number. The following bits are defined:
//				    	 Bit 0: Continuation - If set to 1, the Name Content record is continued in the next "NM" entry.
//		                 Bit 1: Current - If set to 1, the Name Content record refers to the current directory.
//		                 Bit 2: Parent - If set to 1, the Name Content record refers to the parent directory.
//		                 Bit 3: Reserved - Should be set to 0.
//		                 Bit 4: Reserved - Should be set to 0.
//		                 Bit 5: Historical - Historically contains the network node name.
//		                 Bit 6: Reserved - Should be set to 0.
//		                 Bit 7: Reserved - Should be set to 0.
//		  Offset 5:   Name Content - Variable length. The Name Content field shall contain the name of the file or directory. Offset 5-LEN_NM
//	   NOTE: data starts at offset 4 of the System Use Entry record
func UnmarshalRockRidgeNameEntry(length uint8, data []byte) *RockRidgeNameEntry {
	// Data begins at offset 4 of the System Use Entry record
	nameLen := int(length) - 5
	flags := data[0]

	return &RockRidgeNameEntry{
		Continue:  flags&0x01 > 0,
		Current:   flags&0x02 > 0,
		Parent:    flags&0x04 > 0,
		Reserved1: flags&0x08 > 0,
		Reserved2: flags&0x10 > 0,
		Reserved3: flags&0x20 > 0,
		Reserved4: flags&0x40 > 0,
		Reserved5: flags&0x80 > 0,
		Name:      string(data[1 : nameLen+1]),
	}
}

type RockRidgePosixEntry struct {
	Mode     fs.FileMode
	Links    uint32
	UserId   uint32
	GroupId  uint32
	SerialNo uint32
}

// PX Entry Details
//
//	Offset 0-1: Signature Word - "PX"
//	Offset 2:   Length (LEN_PX) - 8-bit number. The number in this field shall be 5.
//	Offset 3:   System Use Entry Version - 8-bit number. The value of this field shall be 1.
//	Offset 4-11: POSIX File Mode - 64-bit number. The value of this field shall be the POSIX file mode.
//	Offset 12-19: POSIX File Links - 64-bit number. The value of this field shall be the number of links to the file.
//	Offset 20-27: POSIX File User ID - 64-bit number. The value of this field shall be the POSIX file user ID.
//	Offset 28-35: POSIX File Group ID - 64-bit number. The value of this field shall be the POSIX file group ID.
//	Offset 36-43: POSIX File Serial Number - 64-bit number. The value of this field shall be the POSIX file serial number.
func UnmarshalRockRidgePosixEntry(data []byte) (entry *RockRidgePosixEntry, err error) {
	// IMPORTANT: data input begins at offset 4 of the System Use Entry record
	modeVal, err := encoding.UnmarshalUint32LSBMSB(data[0:8])
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling POSIX file mode: %s", err)
	}

	mode := parseFileMode(modeVal)

	links, err := encoding.UnmarshalUint32LSBMSB(data[8:16])
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling POSIX file links: %s", err)
	}

	userId, err := encoding.UnmarshalUint32LSBMSB(data[16:24])
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling POSIX file user ID: %s", err)
	}

	groupId, err := encoding.UnmarshalUint32LSBMSB(data[24:32])
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling POSIX file group ID: %s", err)
	}

	serialNo, err := encoding.UnmarshalUint32LSBMSB(data[32:40])
	// TODO: Look more into why some of the serial number unmarshaling has issues
	//if err != nil {
	//	return nil, fmt.Errorf("Error unmarshalling POSIX file serial number: %s", err)
	//}
	return &RockRidgePosixEntry{
		Mode:     mode,
		Links:    links,
		UserId:   userId,
		GroupId:  groupId,
		SerialNo: serialNo,
	}, nil
}

// parseFileMode converts a 32-bit unsigned integer into an fs.FileMode struct
func parseFileMode(mode uint32) fs.FileMode {
	var fileMode fs.FileMode

	// File type bits
	switch mode & 0xF000 {
	case 0xC000:
		fileMode |= fs.ModeSocket
	case 0xA000:
		fileMode |= fs.ModeSymlink
	case 0x8000:
		// Regular file, no specific fs flag needed
	case 0x6000:
		fileMode |= fs.ModeDevice
	case 0x2000:
		fileMode |= fs.ModeCharDevice
	case 0x4000:
		fileMode |= fs.ModeDir
	case 0x1000:
		fileMode |= fs.ModeNamedPipe
	}

	// Permission bits
	if mode&0x0100 != 0 {
		fileMode |= 0400 // S_IRUSR
	}
	if mode&0x0080 != 0 {
		fileMode |= 0200 // S_IWUSR
	}
	if mode&0x0040 != 0 {
		fileMode |= 0100 // S_IXUSR
	}
	if mode&0x0020 != 0 {
		fileMode |= 0040 // S_IRGRP
	}
	if mode&0x0010 != 0 {
		fileMode |= 0020 // S_IWGRP
	}
	if mode&0x0008 != 0 {
		fileMode |= 0010 // S_IXGRP
	}
	if mode&0x0004 != 0 {
		fileMode |= 0004 // S_IROTH
	}
	if mode&0x0002 != 0 {
		fileMode |= 0002 // S_IWOTH
	}
	if mode&0x0001 != 0 {
		fileMode |= 0001 // S_IXOTH
	}

	// Special mode bits
	if mode&0x0800 != 0 {
		fileMode |= os.ModeSetuid
	}
	if mode&0x0400 != 0 {
		fileMode |= os.ModeSetgid
	}
	if mode&0x0200 != 0 {
		fileMode |= os.ModeSticky
	}

	return fileMode
}
