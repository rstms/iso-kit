package path

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/bgrewell/iso-kit/pkg/logging"
	"github.com/go-logr/logr"
)

func NewExtendedAttributeRecord(logger logr.Logger) *ExtendedAttributeRecord {
	return &ExtendedAttributeRecord{
		logger: logger,
	}
}

type ExtendedAttributeRecord struct {
	OwnerIdentifier                uint16
	GroupIdentifier                uint16
	Permissions                    uint16
	FileCreationDate               [8]byte
	FileModificationDate           [8]byte
	FileExpirationDate             [8]byte
	FileEffectiveDate              [8]byte
	RecordFormat                   uint8
	RecordAttributes               uint8
	RecordLength                   uint32
	SystemUseIdentifier            [32]byte
	SystemUse                      [64]byte
	ExtendedAttributeRecordVersion uint8
	LengthOfEscapeSequences        uint8
	Unused1                        [64]byte
	LengthOfApplicationUse         uint32
	ApplicationUse                 []byte
	EscapeSequences                []byte
	logger                         logr.Logger
}

// Unmarshal parses the given data into the ExtendedAttributeRecord struct
func (ear *ExtendedAttributeRecord) Unmarshal(data []byte) error {
	if len(data) < 256 {
		return errors.New("invalid data length")
	}

	// Parse fields
	ear.OwnerIdentifier = binary.LittleEndian.Uint16(data[0:4])
	ear.GroupIdentifier = binary.LittleEndian.Uint16(data[4:8])
	ear.Permissions = binary.LittleEndian.Uint16(data[8:10])

	copy(ear.FileCreationDate[:], data[10:27])
	copy(ear.FileModificationDate[:], data[27:44])
	copy(ear.FileExpirationDate[:], data[44:61])
	copy(ear.FileEffectiveDate[:], data[61:78])

	ear.RecordFormat = data[78]
	ear.RecordAttributes = data[79]
	ear.RecordLength = binary.LittleEndian.Uint32(data[80:84])

	copy(ear.SystemUseIdentifier[:], data[84:116])
	copy(ear.SystemUse[:], data[116:180])

	ear.ExtendedAttributeRecordVersion = data[180]
	ear.LengthOfEscapeSequences = data[181]

	copy(ear.Unused1[:], data[182:246])

	ear.LengthOfApplicationUse = binary.LittleEndian.Uint32(data[246:250])

	// Make sure slices won't go out of range
	appUseEnd := 250 + ear.LengthOfApplicationUse
	if appUseEnd > uint32(len(data)) {
		return fmt.Errorf("applicationUse slice out of range: end=%d, len(data)=%d", appUseEnd, len(data))
	}
	// Copy into a new slice, avoiding reference to the original 'data'
	ear.ApplicationUse = append([]byte(nil), data[250:appUseEnd]...)

	escSeqEnd := appUseEnd + uint32(ear.LengthOfEscapeSequences)
	if escSeqEnd > uint32(len(data)) {
		return fmt.Errorf("escapeSequences slice out of range: end=%d, len(data)=%d", escSeqEnd, len(data))
	}
	// Copy into a new slice, too
	ear.EscapeSequences = append([]byte(nil), data[appUseEnd:escSeqEnd]...)

	// Single grouped logging call
	ear.logger.V(logging.TRACE).Info("Extended Attribute Record fields",
		"ownerIdentifier", ear.OwnerIdentifier,
		"groupIdentifier", ear.GroupIdentifier,
		"permissions", ear.Permissions,
		"fileCreationDate", string(ear.FileCreationDate[:]),
		"fileModificationDate", string(ear.FileModificationDate[:]),
		"fileExpirationDate", string(ear.FileExpirationDate[:]),
		"fileEffectiveDate", string(ear.FileEffectiveDate[:]),
		"recordFormat", ear.RecordFormat,
		"recordAttributes", ear.RecordAttributes,
		"recordLength", ear.RecordLength,
		"systemUseIdentifier", string(ear.SystemUseIdentifier[:]),
		"systemUse", string(ear.SystemUse[:]),
		"extendedAttributeRecordVersion", ear.ExtendedAttributeRecordVersion,
		"lengthOfEscapeSequences", ear.LengthOfEscapeSequences,
		"lengthOfApplicationUse", ear.LengthOfApplicationUse,
		"applicationUse", string(ear.ApplicationUse),
		"escapeSequences", string(ear.EscapeSequences),
	)

	return nil
}
