package extensions

import (
	"bytes"
	"encoding/binary"
	"errors"
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

type RockRidgeExtensions struct {
	// PX - POSIX file permissions (UID, GID, Mode)
	UID         *uint32 // User ID
	GID         *uint32 // Group ID
	Permissions *uint16 // File permissions

	// PN - Device number (if block/char device)
	Major *uint32
	Minor *uint32

	// SL - Symbolic link target path
	SymlinkTarget *string
	SymlinkFlags  *byte // Stores flags for symlink interpretation

	// NM - Alternate name (long filename or case-sensitive filename)
	AlternateName *string

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

// UnmarshalRockRidge extracts Rock Ridge extensions from raw directory record data.
func UnmarshalRockRidge(data []byte) (*RockRidgeExtensions, error) {
	if len(data) < 2 {
		return nil, errors.New("invalid Rock Ridge data")
	}

	rr := &RockRidgeExtensions{}
	reader := bytes.NewReader(data)

	for reader.Len() > 2 {
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

		// Read payload based on the entry type
		payload := make([]byte, length-4)
		if _, err := reader.Read(payload); err != nil {
			return nil, err
		}

		switch RockRidgeEntryType(entryType) {
		case POSIX_FILE_PERMS: // PX (POSIX permissions)
			if len(payload) >= 10 {
				rr.UID = new(uint32)
				rr.GID = new(uint32)
				rr.Permissions = new(uint16)
				*rr.UID = binary.LittleEndian.Uint32(payload[0:4])
				*rr.GID = binary.LittleEndian.Uint32(payload[4:8])
				*rr.Permissions = binary.LittleEndian.Uint16(payload[8:10])
			}
		case POSIX_DEVICE_NUM: // PN (Device major/minor numbers)
			if len(payload) >= 8 {
				rr.Major = new(uint32)
				rr.Minor = new(uint32)
				*rr.Major = binary.LittleEndian.Uint32(payload[0:4])
				*rr.Minor = binary.LittleEndian.Uint32(payload[4:8])
			}
		case SYMBOLIC_LINK: // SL (Symbolic link)
			rr.SymlinkTarget = new(string)
			*rr.SymlinkTarget = string(payload)
		case ALTERNATE_NAME: // NM (Alternate name)
			rr.AlternateName = new(string)
			*rr.AlternateName = string(payload)
		case CHILD_LINK: // CL (Child link)
			if len(payload) >= 4 {
				rr.ChildLinkLBA = new(uint32)
				*rr.ChildLinkLBA = binary.LittleEndian.Uint32(payload[0:4])
			}
		case PARENT_LINK: // PL (Parent link)
			if len(payload) >= 4 {
				rr.ParentLinkLBA = new(uint32)
				*rr.ParentLinkLBA = binary.LittleEndian.Uint32(payload[0:4])
			}
		case RELOCATED_DIR: // RE (Relocated directory flag)
			rr.IsRelocated = new(bool)
			*rr.IsRelocated = true
		case TIME_STAMPS: // TF (Timestamps)
			if len(payload) >= 7 {
				tt := time.Unix(int64(binary.LittleEndian.Uint32(payload[0:4])), 0)
				rr.CreationTime = &tt
			}
		case SPARSE_FILE: // SF (Sparse file flag)
			rr.IsSparse = new(bool)
			*rr.IsSparse = true
		}
	}

	return rr, nil
}

// MarshalRockRidge serializes Rock Ridge extension fields into ISO format.
func MarshalRockRidge(rr *RockRidgeExtensions) ([]byte, error) {
	var buf bytes.Buffer

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
