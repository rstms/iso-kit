package susp

import (
	"fmt"
	"github.com/bgrewell/iso-kit/pkg/consts"
	"github.com/bgrewell/iso-kit/pkg/encoding"
	"github.com/bgrewell/iso-kit/pkg/logging"
	"github.com/bgrewell/iso-kit/pkg/rockridge"
	"io"
)

type SystemUseEntryType string

const (
	CONTINUATION_AREA          SystemUseEntryType = "CE"
	PADDING_FIELD              SystemUseEntryType = "PD"
	SHARING_PROTOCOL_INDICATOR SystemUseEntryType = "SP"
	AREA_TERMINATOR            SystemUseEntryType = "ST"
	EXTENSION_REFERENCE        SystemUseEntryType = "ER"
	EXTENSION_SELECTOR         SystemUseEntryType = "ES"
)

// GetSystemUseEntries returns a slice of SystemUseEntry elements from the SUSP area
func GetSystemUseEntries(data []byte, isoReader io.ReaderAt) (SystemUseEntries, error) {
	visited := make(map[uint32]bool)
	entries, err := ParseSystemUseEntries(data, visited, isoReader)
	if err != nil {
		return nil, err
	}
	return entries, nil
}

// ParseSystemUseEntries parses the System Use Entries from the SUSP area recursively as needed
// and returns a slice of SystemUseEntry elements. Generally speaking, GetSystemUseEntries should be
// used instead of this function.
func ParseSystemUseEntries(data []byte, visited map[uint32]bool, isoReader io.ReaderAt) ([]*SystemUseEntry, error) {
	var entries []*SystemUseEntry

	logging.Logger().Tracef("=== Parsing SystemUseEntries from %d bytes", len(data))
	for offset := 0; offset < len(data); {

		// Check if the remaining data is padding
		if data[offset] == 0x00 {
			break
		}

		// Valid entry must have at least 4 bytes
		if len(data[offset:]) < 4 {
			diff := len(data[offset:])
			logging.Logger().Warnf("Invalid entry length %d, remaining data: %x", diff, data[offset:])
			break
		}
		logging.Logger().Tracef("Parsing entry at offset %d", offset)

		// Get the entry length from BP3
		entryLen := int(data[offset+2])
		logging.Logger().Tracef("Entry length: %d", entryLen)
		if entryLen < 4 {
			return nil, fmt.Errorf("invalid entry length %d", entryLen)
		} else if entryLen > len(data) {
			return nil, fmt.Errorf("invalid entry length %d, length exceeds data length %d", entryLen, len(data[offset:]))
		}

		// Create a new SystemUseEntry
		entry := &SystemUseEntry{
			entryType: SystemUseEntryType(data[offset : offset+2]),
			length:    data[offset+2],
			data:      data[offset+4 : offset+entryLen],
		}
		logging.Logger().Tracef("Entry type: %v", entry.Type())

		// Handle the entry types
		switch entry.Type() {
		case CONTINUATION_AREA:

			logging.Logger().Errorf("ContinuationArea entry: %v", entry)
			// 1: Parse/Unmarshal the ContinuationArea entry
			record, err := UnmarshalContinuationEntry(entry)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal ContinuationArea entry: %w", err)
			}
			// 2: Add the block location to the visited map
			if _, exists := visited[record.blockLocation]; exists {
				return nil, fmt.Errorf("circular reference detected in ContinuationArea")
			}
			visited[record.blockLocation] = true

			// 3: Read the continuation data from the block location
			buffer := make([]byte, record.lengthOfArea)
			ceOffset := (record.blockLocation * consts.ISO9660_SECTOR_SIZE) + record.offset
			if _, err = isoReader.ReadAt(buffer, int64(ceOffset)); err != nil {
				return nil, fmt.Errorf("failed to read continuation data at offset %d: %w", ceOffset, err)
			}
			// 4: Recursively call ParseSystemUseEntries with the continuation data and the visited map
			continuedEntries, err := ParseSystemUseEntries(buffer, visited, isoReader)
			if err != nil {
				return nil, fmt.Errorf("failed to parse continuation data: %w", err)
			}
			// 5: Append the returned entries to the entries slice
			entries = append(entries, continuedEntries...)
		case SHARING_PROTOCOL_INDICATOR:
			logging.Logger().Tracef("SharingProtocolIndicator entry: %v", entry)
			entries = append(entries, entry)
		case EXTENSION_REFERENCE:
			logging.Logger().Tracef("ExtensionReference entry: %v", entry)
			entries = append(entries, entry)
		case EXTENSION_SELECTOR:
			logging.Logger().Tracef("ExtensionSelector entry: %v", entry)
			entries = append(entries, entry)
		case AREA_TERMINATOR:
			logging.Logger().Tracef("AreaTerminator entry: %v", entry)
			entries = append(entries, entry)
		case PADDING_FIELD:
			logging.Logger().Tracef("PaddingField entry: %v", entry)
			entries = append(entries, entry)
		default:
			entries = append(entries, entry)
		}

		offset += entryLen
	}

	logging.Logger().Tracef("=== Parsed %d SystemUseEntries", len(entries))
	return entries, nil
}

// SystemUseEntry represents a System Use Entry in the SUSP area
type SystemUseEntry struct {
	entryType SystemUseEntryType
	length    uint8
	data      []byte
}

// Type returns the SystemUseEntryType of the SystemUseEntry
func (e SystemUseEntry) Type() SystemUseEntryType {
	return e.entryType
}

// Length returns the length of the SystemUseEntry
func (e SystemUseEntry) Length() uint8 {
	return e.length
}

// Data returns the raw data of the SystemUseEntry
func (e SystemUseEntry) Data() []byte {
	return e.data
}

// Unmarshal unmarshals the SystemUseEntry data into a struct
func (e *SystemUseEntry) Unmarshal(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("invalid SystemUseEntry data length %d", len(data))
	}
	e.entryType = SystemUseEntryType(data[0:2])
	e.length = data[2]
	e.data = data[4:]
	return nil
}

// SystemUseEntries is a slice of SystemUseEntry elements with some additional helper methods
type SystemUseEntries []*SystemUseEntry

// Len returns the number of SystemUseEntry elements
func (e SystemUseEntries) Len() int {
	return len(e)
}

func (e SystemUseEntries) GetExtensionRecords() (records []*ExtensionRecord, err error) {
	for _, entry := range e {
		if entry.Type() == EXTENSION_REFERENCE {
			er, err := UnmarshalExtensionRecord(entry)
			if err != nil {
				return nil, err
			}
			records = append(records, er)
		}
	}
	return records, nil
}

// HasRockRidge returns true if the SystemUseEntries contains Rock Ridge extensions
func (e SystemUseEntries) HasRockRidge() bool {
	records, err := e.GetExtensionRecords()
	if err != nil {
		logging.Logger().Errorf("Failed to get ExtensionRecords: %v", err)
		return false
	}

	for i, record := range records {
		logging.Logger().Tracef("ExtensionRecord %d: %v", i, record)
		if record.Identifier == rockridge.ROCK_RIDGE_IDENTIFIER && record.Version == rockridge.ROCK_RIDGE_VERSION {
			logging.Logger().Tracef("Found Rock Ridge extension")
			return true
		}
	}

	// TODO: This is temporary until I figure out why the extension records aren't appearing for the actual items that
	//  have Rock Ridge extensions
	for _, entry := range e {
		if entry.Type() == SystemUseEntryType(rockridge.POSIX_FILE_PERMS) ||
			entry.Type() == SystemUseEntryType(rockridge.ALTERNATE_NAME) ||
			entry.Type() == SystemUseEntryType(rockridge.TIME_STAMPS) {
			logging.Logger().Tracef("Found Rock Ridge extension")
			return true
		}
	}

	return false
}

// RockRidgeName returns the Rock Ridge name if present otherwise it returns nil
func (e SystemUseEntries) RockRidgeName() *string {
	for _, record := range e {
		if record.Type() == SystemUseEntryType(rockridge.ALTERNATE_NAME) {
			var name string
			logging.Logger().Tracef("Found Rock Ridge alternate name")
			entry := rockridge.UnmarshalRockRidgeNameEntry(record.Length(), record.Data())
			if entry == nil {
				logging.Logger().Errorf("Failed to unmarshal Rock Ridge name entry: %w", record.Data())
				return nil
			}
			if entry.Current {
				name = "."
				return &name
			}
			if entry.Parent {
				name = ".."
				return &name
			}

			name = entry.Name
			return &name
		}
	}

	logging.Logger().Tracef("Rock Ridge name not found")
	return nil
}

// RockRidgePermissions returns the Rock Ridge permissions if present otherwise it returns nil
func (e SystemUseEntries) RockRidgePermissions() *rockridge.RockRidgePosixEntry {
	for _, record := range e {
		if record.Type() == SystemUseEntryType(rockridge.POSIX_FILE_PERMS) {
			logging.Logger().Tracef("Found Rock Ridge permissions")
			return rockridge.UnmarshalRockRidgePosixEntry(record.Data())
		}
	}

	logging.Logger().Tracef("Rock Ridge permissions not found")
	return nil
}

// RockRidgeTimestamps returns the Rock Ridge timestamps if present otherwise it returns nil
func (e SystemUseEntries) RockRidgeTimestamps() *rockridge.RockRidgeTimestamps {
	return nil
}

// UnmarshalExtensionRecord unmarshals the SystemUseEntry data into an ExtensionRecord struct
func UnmarshalExtensionRecord(e *SystemUseEntry) (*ExtensionRecord, error) {
	if e.Length() < 8 {
		return nil, fmt.Errorf("invalid ExtensionRecord length %d", e.Length())
	}

	if e.Type() != EXTENSION_REFERENCE {
		return nil, fmt.Errorf("wrong type of record, expected ER")
	}

	identifierLength := e.data[0]
	if e.Length() < 8+identifierLength {
		return nil, fmt.Errorf("invalid identifier data length %d, expected at least %d", e.Length(), 8+identifierLength)
	}

	descriptorLength := e.data[1]
	if e.Length() < 8+identifierLength+descriptorLength {
		return nil, fmt.Errorf("invalid descriptor data length %d, expected at least %d", e.Length(), 8+identifierLength+descriptorLength)
	}

	sourceLength := e.data[2]
	if e.Length() < 8+identifierLength+descriptorLength+sourceLength {
		return nil, fmt.Errorf("invalid source data length %d, expected at least %d", e.Length(), 8+identifierLength+descriptorLength+sourceLength)
	}

	return &ExtensionRecord{
		Version:    int(e.data[3]),
		Identifier: string(e.data[4 : 4+identifierLength]),
		Descriptor: string(e.data[4+identifierLength : 4+identifierLength+descriptorLength]),
		Source:     string(e.data[4+identifierLength+descriptorLength : 4+identifierLength+descriptorLength+sourceLength]),
	}, nil
}

// UnmarshalContinuationEntry unmarshals the SystemUseEntry data into a ContinuationEntry struct
func UnmarshalContinuationEntry(e *SystemUseEntry) (*ContinuationEntry, error) {
	if e.Length() != 28 {
		return nil, fmt.Errorf("invalid ContinuationEntry length %d, espected 28", e.Length())
	}

	location, err := encoding.UnmarshalUint32LSBMSB(e.data[0:8])
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling location: %w", err)
	}
	offset, err := encoding.UnmarshalUint32LSBMSB(e.data[8:16])
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling offset: %w", err)
	}
	length, err := encoding.UnmarshalUint32LSBMSB(e.data[16:24])
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling length: %w", err)
	}

	return &ContinuationEntry{
		blockLocation: location,
		offset:        offset,
		lengthOfArea:  length,
	}, nil
}
