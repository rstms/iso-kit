package susp

import (
	"fmt"
	"github.com/bgrewell/iso-kit/pkg/encoding"
)

type ExtensionRecord struct {
	Version    int
	Identifier string
	Descriptor string
	Source     string
}

// SUSP-112 5.1
type ContinuationEntry struct {
	blockLocation uint32
	offset        uint32
	lengthOfArea  uint32
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
