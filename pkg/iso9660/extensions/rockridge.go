package extensions

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/bgrewell/iso-kit/pkg/iso9660/encoding"
	"io/fs"
	"os"
	"time"
)

const (
	ROCK_RIDGE_IDENTIFIER = "RRIP_1991A"
	ROCK_RIDGE_VERSION    = 1
)

type RockRidgeEntryType string

const (
	//POSIX file permissions (owner, group, other)
	POSIX_FILE_PERMS RockRidgeEntryType = "PX"
	//Device numbers for block/character device nodes (major/minor)
	POSIX_DEVICE_NUM RockRidgeEntryType = "PN"
	//Symbolic link data (path components,flags)
	SYMBOLIC_LINK RockRidgeEntryType = "SL"
	//AlternateName (used for long filenames, case preservation, etc.)
	ALTERNATE_NAME RockRidgeEntryType = "NM"
	//ChildLink (used for directory relocation chains)
	CHILD_LINK RockRidgeEntryType = "CL"
	//ParentLink (links a relocated directory back to its parent)
	PARENT_LINK RockRidgeEntryType = "PL"
	//Marks a directory that has been relocated
	RELOCATED_DIR RockRidgeEntryType = "RE"
	//Time stamp information (creation, modification, access,etc)
	TIME_STAMPS RockRidgeEntryType = "TF"
	//Sparse file information (less commonly used)
	SPARSE_FILE RockRidgeEntryType = "SF"
	//An older “Rock Ridge” extension signature (now typically replaced by ER).
	ROCK_RIDGE RockRidgeEntryType = "RR"
)

type NameEntryFlags struct {
	Continue  bool // Bit 0: Alternate Name continues in the next "NM" entry
	Current   bool // Bit 1: Alternate Name refers to the current directory ("." in POSIX)
	Parent    bool // Bit 2: Alternate Name refers to the parent directory (".." in POSIX)
	Reserved1 bool // Bit 3: Reserved, should be ZERO
	Reserved2 bool // Bit 4: Reserved, should be ZERO
	Reserved3 bool // Bit 5: Historically contains the network node name
	Reserved4 bool // Bit 6: Unused, reserved for future use
	Reserved5 bool // Bit 7: Unused, reserved for future use
}

type RockRidgeExtensions struct {
	// PX - POSIX file permissions (UID, GID, Mode)
	UID         *uint32      // User ID
	GID         *uint32      // Group ID
	Permissions *fs.FileMode // File permissions

	// PN - Device number (if block/char device)
	Major *uint32
	Minor *uint32

	// SL - Symbolic link target path
	SymlinkTarget *string
	SymlinkFlags  *byte // Stores flags for symlink interpretation

	// NM - Alternate name (long filename or case-sensitive filename)
	AlternateNameFlags *NameEntryFlags
	AlternateName      *string

	// CL - Child link LBA (for relocated directories)
	ChildLinkLBA *uint32

	// PL - Parent link LBA (if directory was relocated)
	ParentLinkLBA *uint32

	// RE - Relocated directory flag
	IsRelocated *bool

	// TF - Time stamps (creation, modification, access)
	CreationTime     *time.Time
	ModificationTime *time.Time
	AccessTime       *time.Time

	// SF - Sparse file info (if applicable)
	IsSparse *bool
}

// HasRockRidge determines if any Rock Ridge extensions were set.
func (r *RockRidgeExtensions) HasRockRidge() bool {
	return r.UID != nil || r.GID != nil || r.Permissions != nil ||
		r.Major != nil || r.Minor != nil || r.SymlinkTarget != nil ||
		r.AlternateName != nil || r.ChildLinkLBA != nil || r.ParentLinkLBA != nil ||
		r.IsRelocated != nil || r.CreationTime != nil || r.ModificationTime != nil ||
		r.AccessTime != nil || r.IsSparse != nil
}

func UnmarshalRockRidge(data []byte) (*RockRidgeExtensions, error) {
	if len(data) < 2 {
		return nil, errors.New("invalid Rock Ridge data")
	}

	rr := &RockRidgeExtensions{}
	reader := bytes.NewReader(data)

	for reader.Len() > 4 {
		// Read signature (2-byte identifier)
		var sig [2]byte
		if err := binary.Read(reader, binary.LittleEndian, &sig); err != nil {
			return nil, err
		}
		entryType := string(sig[:])

		// Read length (1 byte)
		var length byte
		if err := binary.Read(reader, binary.LittleEndian, &length); err != nil {
			return nil, err
		}

		// Read version (1 byte)
		var version byte
		if err := binary.Read(reader, binary.LittleEndian, &version); err != nil {
			return nil, err
		}

		// Read payload
		payloadLen := int(length) - 4
		payload := make([]byte, payloadLen)
		if _, err := reader.Read(payload); err != nil {
			return nil, err
		}

		switch RockRidgeEntryType(entryType) {
		case POSIX_FILE_PERMS: // PX (POSIX permissions)
			if len(payload) >= 32 {
				// Payload is the bytes from offset 4 to 36 (32 bytes) ... technically there are another 8 bytes for the
				// file serial number but that generally hasn't been present so it is ignored here.
				// Decode 8-byte File Mode (Permissions)
				mode, err := encoding.UnmarshalUint32LSBMSB([8]byte(payload[0:8]))
				if err == nil {
					permissions := parseFileMode(mode)
					rr.Permissions = &permissions
				}

				// Decode 8-byte Number of Links
				_, err = encoding.UnmarshalUint32LSBMSB([8]byte(payload[8:16]))
				if err != nil {
					return nil, errors.New("failed to parse PX link count")
				}

				// Decode 8-byte UID
				uid, err := encoding.UnmarshalUint32LSBMSB([8]byte(payload[16:24]))
				if err == nil {
					rr.UID = &uid
				}

				// Decode 8-byte GID
				gid, err := encoding.UnmarshalUint32LSBMSB([8]byte(payload[24:32]))
				if err == nil {
					rr.GID = &gid
				}
			}
		case TIME_STAMPS: // TF (Timestamps)
			// Rock Ridge TF entry uses a **variable-length encoding**
			offset := 0
			for offset < len(payload) {
				flag := payload[offset]
				offset++

				if flag&0x01 != 0 && offset+7 <= len(payload) { // Creation time
					seconds := int64(binary.LittleEndian.Uint32(payload[offset : offset+4]))
					rr.CreationTime = new(time.Time)
					*rr.CreationTime = time.Unix(seconds, 0)
					offset += 7
				}

				if flag&0x02 != 0 && offset+7 <= len(payload) { // Modification time
					seconds := int64(binary.LittleEndian.Uint32(payload[offset : offset+4]))
					rr.ModificationTime = new(time.Time)
					*rr.ModificationTime = time.Unix(seconds, 0)
					offset += 7
				}

				if flag&0x04 != 0 && offset+7 <= len(payload) { // Access time
					seconds := int64(binary.LittleEndian.Uint32(payload[offset : offset+4]))
					rr.AccessTime = new(time.Time)
					*rr.AccessTime = time.Unix(seconds, 0)
					offset += 7
				}
			}

		case ALTERNATE_NAME: // NM (Alternate name)
			// Flags (NM Flags) - 8-bit number. The following bits are defined:
			// 	 Bit 0: Continuation - If set to 1, the Name Content record is continued in the next "NM" entry.
			//   Bit 1: Current - If set to 1, the Name Content record refers to the current directory.
			//   Bit 2: Parent - If set to 1, the Name Content record refers to the parent directory.
			//   Bit 3: Reserved - Should be set to 0.
			//   Bit 4: Reserved - Should be set to 0.
			//   Bit 5: Historical - Historically contains the network node name.
			//   Bit 6: Reserved - Should be set to 0.
			//   Bit 7: Reserved - Should be set to 0.
			flags := payload[0]
			rr.AlternateNameFlags = &NameEntryFlags{
				Continue:  flags&0x01 > 0,
				Current:   flags&0x02 > 0,
				Parent:    flags&0x04 > 0,
				Reserved1: flags&0x08 > 0,
				Reserved2: flags&0x10 > 0,
				Reserved3: flags&0x20 > 0,
				Reserved4: flags&0x40 > 0,
				Reserved5: flags&0x80 > 0,
			}
			rr.AlternateName = new(string)
			*rr.AlternateName = string(payload[1:])

		case SYMBOLIC_LINK: // SL (Symbolic link)
			rr.SymlinkTarget = new(string)
			*rr.SymlinkTarget = string(payload[1:]) // Skip flags byte
		}
	}

	return rr, nil
}

// MarshalRockRidge serializes Rock Ridge extension fields into ISO format.
func MarshalRockRidge(rr *RockRidgeExtensions) ([]byte, error) {
	var buf bytes.Buffer

	//TODO: Fix this whole function, there were a lot of errors with sizes and offsets
	if rr.UID != nil && rr.GID != nil && rr.Permissions != nil {
		buf.Write([]byte("PX"))           // Signature
		buf.WriteByte(10 + 4)             // Length (10 bytes data + header)
		buf.WriteByte(ROCK_RIDGE_VERSION) // Version
		binary.Write(&buf, binary.LittleEndian, *rr.UID)
		binary.Write(&buf, binary.LittleEndian, *rr.GID)
		binary.Write(&buf, binary.LittleEndian, *rr.Permissions)
	}

	if rr.Major != nil && rr.Minor != nil {
		buf.Write([]byte("PN"))           // Signature
		buf.WriteByte(8 + 4)              // Length (8 bytes data + header)
		buf.WriteByte(ROCK_RIDGE_VERSION) // Version
		binary.Write(&buf, binary.LittleEndian, *rr.Major)
		binary.Write(&buf, binary.LittleEndian, *rr.Minor)
	}

	if rr.SymlinkTarget != nil {
		buf.Write([]byte("SL")) // Signature
		buf.WriteByte(byte(len(*rr.SymlinkTarget) + 4))
		buf.WriteByte(ROCK_RIDGE_VERSION)
		buf.WriteString(*rr.SymlinkTarget)
	}

	if rr.AlternateName != nil {
		buf.Write([]byte("NM")) // Signature
		buf.WriteByte(byte(len(*rr.AlternateName) + 4))
		buf.WriteByte(ROCK_RIDGE_VERSION)
		buf.WriteString(*rr.AlternateName)
	}

	if rr.ChildLinkLBA != nil {
		buf.Write([]byte("CL")) // Signature
		buf.WriteByte(4 + 4)
		buf.WriteByte(ROCK_RIDGE_VERSION)
		binary.Write(&buf, binary.LittleEndian, *rr.ChildLinkLBA)
	}

	if rr.ParentLinkLBA != nil {
		buf.Write([]byte("PL")) // Signature
		buf.WriteByte(4 + 4)
		buf.WriteByte(ROCK_RIDGE_VERSION)
		binary.Write(&buf, binary.LittleEndian, *rr.ParentLinkLBA)
	}

	if rr.IsRelocated != nil && *rr.IsRelocated {
		buf.Write([]byte("RE")) // Signature
		buf.WriteByte(4)
		buf.WriteByte(ROCK_RIDGE_VERSION)
	}

	if rr.CreationTime != nil {
		buf.Write([]byte("TF")) // Signature
		buf.WriteByte(7 + 4)
		buf.WriteByte(ROCK_RIDGE_VERSION)
		binary.Write(&buf, binary.LittleEndian, uint32(rr.CreationTime.Unix()))
	}

	if rr.IsSparse != nil && *rr.IsSparse {
		buf.Write([]byte("SF")) // Signature
		buf.WriteByte(4)
		buf.WriteByte(ROCK_RIDGE_VERSION)
	}

	return buf.Bytes(), nil
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
