package directory

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/bgrewell/iso-kit/pkg/logging"
	"github.com/bgrewell/iso-kit/pkg/rockridge"
	"github.com/bgrewell/iso-kit/pkg/susp"
	"io"
	"unicode/utf16"
)

// DirectoryRecord represents a single Record in a directory.
type DirectoryRecord struct {
	LengthOfDirectoryRecord uint8
	ExtendedAttributeRecord uint8
	LocationOfExtent        uint32
	DataLength              uint32
	RecordingDateAndTime    []byte
	FileFlags               *FileFlags
	FileUnitSize            uint8
	InterleaveGapSize       uint8
	VolumeSequenceNumber    uint16
	FileIdentifierLength    uint8
	FileIdentifier          string
	PaddingField            []byte
	SystemUse               []byte
	SystemUseEntries        susp.SystemUseEntries
	ExtensionRecords        []*susp.ExtensionRecord
	Joliet                  bool
	hasRockRidge            bool
	rockRidgeName           *string
	rockRidgePermissions    *rockridge.RockRidgePosixEntry
	rockRidgeTimestamps     *rockridge.RockRidgeTimestamps
}

// Unmarshal decodes a DirectoryRecord from binary form.
func (dr *DirectoryRecord) Unmarshal(data []byte, isoFile io.ReaderAt) error {
	if len(data) < 33 {
		return errors.New("invalid data length")
	}

	logging.Logger().Tracef("==== Start Directory Record ====")

	// Basic fields
	dr.LengthOfDirectoryRecord = data[0]
	dr.ExtendedAttributeRecord = data[1]
	dr.LocationOfExtent = binary.LittleEndian.Uint32(data[2:6])
	dr.DataLength = binary.LittleEndian.Uint32(data[10:14])
	dr.RecordingDateAndTime = data[18:25]
	dr.FileFlags = &FileFlags{}
	dr.FileFlags.Set(data[25])
	dr.FileUnitSize = data[26]
	dr.InterleaveGapSize = data[27]
	dr.VolumeSequenceNumber = binary.LittleEndian.Uint16(data[28:30])
	dr.FileIdentifierLength = data[32]

	// Log the basic fields
	logging.Logger().Tracef("Length of directory Record: %d", dr.LengthOfDirectoryRecord)
	logging.Logger().Tracef("Extended attribute Record: %d", dr.ExtendedAttributeRecord)
	logging.Logger().Tracef("Location of extent: %d", dr.LocationOfExtent)
	logging.Logger().Tracef("Data length: %d", dr.DataLength)
	logging.Logger().Tracef("Recording date and time: %d", dr.RecordingDateAndTime)
	logging.Logger().Tracef("File flags: %s", dr.FileFlags.String())
	logging.Logger().Tracef("File unit size: %d", dr.FileUnitSize)
	logging.Logger().Tracef("Interleave gap size: %d", dr.InterleaveGapSize)
	logging.Logger().Tracef("Volume sequence number: %d", dr.VolumeSequenceNumber)
	logging.Logger().Tracef("File identifier length: %d", dr.FileIdentifierLength)

	// Handle Joliet vs. non-Joliet file identifiers
	rawIdentifier := data[33 : 33+dr.FileIdentifierLength]
	if dr.Joliet && dr.FileIdentifierLength != 1 {
		jolietName, err := DecodeJolietName(rawIdentifier)
		if err != nil {
			return fmt.Errorf("failed to decode Joliet name: %w", err)
		}
		dr.FileIdentifier = jolietName
	} else {
		dr.FileIdentifier = string(rawIdentifier)
	}

	// Special cases: root dir and parent dir
	identifier := dr.FileIdentifier
	switch identifier {
	case "\x00":
		identifier = "<root_dir>"
	case "\x01":
		identifier = "<parent>"
	}
	logging.Logger().Tracef("File identifier: %s", identifier)

	// Compute where system use fields begin
	systemUseStart := 33 + dr.FileIdentifierLength
	if dr.FileIdentifierLength%2 == 0 {
		dr.PaddingField = data[systemUseStart : systemUseStart+1]
		systemUseStart++
		logging.Logger().Tracef("File identifier is even so padding field value set to: %x", dr.PaddingField)
	} else {
		dr.PaddingField = nil
	}
	logging.Logger().Tracef("System use start calculated at: %d", systemUseStart)

	if int(systemUseStart) > len(data) {
		logging.Logger().Errorf("System use start is greater than data length: %d > %d", systemUseStart, len(data))
		return nil // or return an error, depending on desired behavior
	}

	// Parse system use entries (e.g., SUSP, Rock Ridge, etc.)
	systemUse := data[systemUseStart:]
	if len(systemUse) > 0 {
		dr.SystemUse = systemUse
		logging.Logger().Tracef("System use: %x (length = %d)", dr.SystemUse, len(dr.SystemUse))

		entries, err := susp.GetSystemUseEntries(systemUse, isoFile)
		if err != nil {
			return err
		}
		dr.SystemUseEntries = entries

		extensionRecords, err := dr.SystemUseEntries.GetExtensionRecords()
		if err != nil {
			return err
		}
		dr.ExtensionRecords = extensionRecords

		dr.hasRockRidge = dr.SystemUseEntries.HasRockRidge()
		if dr.hasRockRidge {
			dr.rockRidgeName = dr.SystemUseEntries.RockRidgeName()
			if dr.rockRidgeName == nil {
				logging.Logger().Error("Rock Ridge name is nil")
			} else {
				logging.Logger().Tracef("Rock Ridge name: %s", *dr.rockRidgeName)
			}

			dr.rockRidgePermissions = dr.SystemUseEntries.RockRidgePermissions()
			if dr.rockRidgePermissions == nil {
				logging.Logger().Error("Rock Ridge permissions are nil")
			} else {
				logging.Logger().Tracef("Rock Ridge permissions: %v", dr.rockRidgePermissions)
			}

			dr.rockRidgeTimestamps = dr.SystemUseEntries.RockRidgeTimestamps()
		}
	} else {
		logging.Logger().Trace("System use: nil")
	}

	logging.Logger().Tracef("==== End Directory Record ====")
	return nil
}

// HasRockRidge returns true if the directory record has Rock Ridge extensions.
func (dr DirectoryRecord) HasRockRidge() bool {
	return dr.hasRockRidge
}

// RockRidgeName returns the Rock Ridge name of the directory record.
func (dr DirectoryRecord) RockRidgeName() *string {
	return dr.rockRidgeName
}

// RockRidgePermissions returns the Rock Ridge permissions of the directory record.
func (dr DirectoryRecord) RockRidgePermissions() *rockridge.RockRidgePosixEntry {
	return dr.rockRidgePermissions
}

// RockRidgeTimestamps returns the Rock Ridge timestamps of the directory record.
func (dr DirectoryRecord) RockRidgeTimestamps() *rockridge.RockRidgeTimestamps {
	return dr.rockRidgeTimestamps
}

// DecodeJolietName converts a Joliet file identifier (UTF-16BE) into a Go string.
func DecodeJolietName(data []byte) (string, error) {
	if len(data) == 0 {
		return "", nil // Empty name
	}

	// Special case: Root, Current, and Parent Directory Identifiers
	if len(data) == 1 {
		switch data[0] {
		case 0x00:
			return ".", nil // Root or Current Directory
		case 0x01:
			return "..", nil // Parent Directory
		default:
			return "", fmt.Errorf("invalid single-byte directory identifier: 0x%02X", data[0])
		}
	}

	// Ensure data length is even for UTF-16 decoding
	if len(data)%2 != 0 {
		return "", fmt.Errorf("invalid Joliet file identifier: odd byte length")
	}

	// Read as UTF-16 big-endian
	utf16Chars := make([]uint16, len(data)/2)
	err := binary.Read(bytes.NewReader(data), binary.BigEndian, &utf16Chars)
	if err != nil {
		return "", fmt.Errorf("failed to read UTF-16BE: %w", err)
	}

	// Convert UTF-16 to Go string
	name := string(utf16.Decode(utf16Chars))

	// Joliet allows null-padded names, trim null padding
	name = trimNullPadding(name)

	// Validate allowed character set per Joliet spec
	if err := validateJolietCharacters(name); err != nil {
		return "", err
	}

	return name, nil
}

// trimNullPadding removes trailing null characters (U+0000) from the string.
func trimNullPadding(s string) string {
	for len(s) > 0 && s[len(s)-1] == '\x00' {
		s = s[:len(s)-1]
	}
	return s
}

// validateJolietCharacters ensures the decoded name complies with allowed UCS-2 characters.
func validateJolietCharacters(name string) error {
	for _, r := range name {
		if r <= 0x001F || r == 0x002A || r == 0x002F || r == 0x003A ||
			r == 0x003B || r == 0x003F || r == 0x005C {
			return fmt.Errorf("invalid character 0x%04X in Joliet file identifier", r)
		}
	}
	return nil
}
