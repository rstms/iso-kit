package susp

import (
	"fmt"
	"github.com/bgrewell/iso-kit/pkg/consts"
	"github.com/bgrewell/iso-kit/pkg/logging"
	"github.com/go-logr/logr"
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
func GetSystemUseEntries(data []byte, isoReader io.ReaderAt, logger logr.Logger) (*SystemUseEntries, error) {
	visited := make(map[uint32]bool)
	entries, err := ParseSystemUseEntries(data, visited, isoReader, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SystemUseEntries: %w", err)
	}
	return NewSystemUseEntries(entries, logger), nil
}

// ParseSystemUseEntries parses the System Use Entries from the SUSP area recursively as needed
// and returns a slice of SystemUseEntry elements. Generally speaking, GetSystemUseEntries should be
// used instead of this function.
func ParseSystemUseEntries(
	data []byte,
	visited map[uint32]bool,
	isoReader io.ReaderAt,
	logger logr.Logger,
) ([]*SystemUseEntry, error) {

	var entries []*SystemUseEntry

	// Top-level info about incoming data
	logger.V(logging.TRACE).Info("Parsing SystemUseEntries", "dataLength", len(data))

	for offset := 0; offset < len(data); {
		// Check if the remaining data is padding
		if data[offset] == 0x00 {
			break
		}

		// Must have at least 4 bytes for a valid entry
		remaining := len(data[offset:])
		if remaining < 4 {
			logger.Error(nil, "WARNING: Invalid entry length",
				"bytesRemaining", remaining,
				"offset", offset,
			)
			break
		}

		// Determine the entry length from BP3
		entryLen := int(data[offset+2])
		// Combine these into a single structured log
		logger.V(logging.TRACE).Info("Parsing system use entry",
			"offset", offset,
			"entryLen", entryLen,
			"bytesRemaining", remaining,
			// You might want to omit or truncate raw data if it’s huge:
			"entryData", data[offset:offset+entryLen],
		)

		if entryLen < 4 {
			return nil, fmt.Errorf("invalid entry length %d", entryLen)
		}
		if entryLen > remaining {
			return nil, fmt.Errorf("entry length %d exceeds remaining data length %d", entryLen, remaining)
		}

		// Create a new SystemUseEntry
		entry := NewSystemUseEntry(
			SystemUseEntryType(data[offset:offset+2]),
			data[offset+2],
			data[offset+4:offset+entryLen],
			logger,
		)

		logger.V(logging.TRACE).Info("Created SystemUseEntry", "entry", entry)

		// Handle the entry types via switch
		switch entry.Type() {
		case CONTINUATION_AREA:
			// 1) Parse/Unmarshal the ContinuationArea entry
			record, err := UnmarshalContinuationEntry(entry)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal ContinuationArea entry: %w", err)
			}

			// 2) Check for circular references
			if _, exists := visited[record.blockLocation]; exists {
				return nil, fmt.Errorf("circular reference detected in ContinuationArea")
			}
			visited[record.blockLocation] = true

			// 3) Read the continuation data
			buffer := make([]byte, record.lengthOfArea)
			ceOffset := (record.blockLocation * consts.ISO9660_SECTOR_SIZE) + record.offset
			if _, err = isoReader.ReadAt(buffer, int64(ceOffset)); err != nil {
				return nil, fmt.Errorf("failed to read continuation data at offset %d: %w", ceOffset, err)
			}

			// 4) Recursively parse
			continuedEntries, err := ParseSystemUseEntries(buffer, visited, isoReader, logger)
			if err != nil {
				return nil, fmt.Errorf("failed to parse continuation data: %w", err)
			}

			// 5) Append
			entries = append(entries, continuedEntries...)

		case SHARING_PROTOCOL_INDICATOR:
			logger.V(logging.TRACE).Info("SharingProtocolIndicator entry", "entry", entry)
			entries = append(entries, entry)

		case EXTENSION_REFERENCE:
			logger.V(logging.TRACE).Info("ExtensionReference entry", "entry", entry)
			entries = append(entries, entry)

		case EXTENSION_SELECTOR:
			logger.V(logging.TRACE).Info("ExtensionSelector entry", "entry", entry)
			entries = append(entries, entry)

		case AREA_TERMINATOR:
			logger.V(logging.TRACE).Info("AreaTerminator entry", "entry", entry)
			entries = append(entries, entry)

		case PADDING_FIELD:
			logger.V(logging.TRACE).Info("PaddingField entry", "entry", entry)
			entries = append(entries, entry)

		default:
			entries = append(entries, entry)
		}

		offset += entryLen
	}

	logger.V(logging.TRACE).Info("Finished parsing SystemUseEntries", "entriesCount", len(entries))
	return entries, nil
}

func NewSystemUseEntry(entryType SystemUseEntryType, length uint8, data []byte, logger logr.Logger) *SystemUseEntry {
	return &SystemUseEntry{
		entryType: entryType,
		length:    length,
		data:      data,
		logger:    logger,
	}
}

// SystemUseEntry represents a System Use Entry in the SUSP area
type SystemUseEntry struct {
	entryType SystemUseEntryType
	length    uint8
	data      []byte
	logger    logr.Logger
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
