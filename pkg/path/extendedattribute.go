package path

import (
	"encoding/binary"
	"errors"
	"github.com/bgrewell/iso-kit/pkg/logging"
)

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
}

func (ear *ExtendedAttributeRecord) Unmarshal(data []byte) error {
	if len(data) < 256 {
		return errors.New("invalid data length")
	}

	ear.OwnerIdentifier = binary.LittleEndian.Uint16(data[0:4])
	logging.Logger().Tracef("Owner identifier: %d", ear.OwnerIdentifier)
	ear.GroupIdentifier = binary.LittleEndian.Uint16(data[4:8])
	logging.Logger().Tracef("Group identifier: %d", ear.GroupIdentifier)
	ear.Permissions = binary.LittleEndian.Uint16(data[8:10])
	logging.Logger().Tracef("Permissions: %d", ear.Permissions)
	copy(ear.FileCreationDate[:], data[10:27])
	logging.Logger().Tracef("File creation date: %s", ear.FileCreationDate)
	copy(ear.FileModificationDate[:], data[27:44])
	logging.Logger().Tracef("File modification date: %s", ear.FileModificationDate)
	copy(ear.FileExpirationDate[:], data[44:61])
	logging.Logger().Tracef("File expiration date: %s", ear.FileExpirationDate)
	copy(ear.FileEffectiveDate[:], data[61:78])
	logging.Logger().Tracef("File effective date: %s", ear.FileEffectiveDate)
	ear.RecordFormat = data[78]
	logging.Logger().Tracef("Record format: %d", ear.RecordFormat)
	ear.RecordAttributes = data[79]
	logging.Logger().Tracef("Record attributes: %d", ear.RecordAttributes)
	ear.RecordLength = binary.LittleEndian.Uint32(data[80:84])
	logging.Logger().Tracef("Record length: %d", ear.RecordLength)
	copy(ear.SystemUseIdentifier[:], data[84:116])
	logging.Logger().Tracef("System use identifier: %s", ear.SystemUseIdentifier)
	copy(ear.SystemUse[:], data[116:180])
	logging.Logger().Tracef("System use: %s", ear.SystemUse)
	ear.ExtendedAttributeRecordVersion = data[180]
	logging.Logger().Tracef("Extended attribute record version: %d", ear.ExtendedAttributeRecordVersion)
	ear.LengthOfEscapeSequences = data[181]
	logging.Logger().Tracef("Length of escape sequences: %d", ear.LengthOfEscapeSequences)
	copy(ear.Unused1[:], data[182:246])
	ear.LengthOfApplicationUse = binary.LittleEndian.Uint32(data[246:250])
	logging.Logger().Tracef("Length of application use: %d", ear.LengthOfApplicationUse)
	ear.ApplicationUse = data[250 : 250+ear.LengthOfApplicationUse]
	ear.EscapeSequences = data[250+ear.LengthOfApplicationUse : 250+ear.LengthOfApplicationUse+uint32(ear.LengthOfEscapeSequences)]

	return nil
}
