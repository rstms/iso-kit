package path

import (
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"testing"
)

//func TestUnmarshal_ValidData(t *testing.T) {
//	logger := logr.Discard()
//	ear := NewExtendedAttributeRecord(logger)
//	data := make([]byte, 2048)
//	copy(data[0:4], []byte{1, 0, 0, 0})
//	copy(data[4:8], []byte{2, 0, 0, 0})
//	copy(data[8:10], []byte{3, 0})
//	copy(data[10:27], []byte("20230101000000"))
//	copy(data[27:44], []byte("20230102000000"))
//	copy(data[44:61], []byte("20230103000000"))
//	copy(data[61:78], []byte("20230104000000"))
//	data[78] = 4
//	data[79] = 5
//	copy(data[80:84], []byte{6, 0, 0, 0})
//	copy(data[84:116], []byte("SystemUseIdentifier"))
//	copy(data[116:180], []byte("SystemUse"))
//	data[180] = 7
//	data[181] = 8
//	copy(data[182:246], []byte("Unused1"))
//	copy(data[246:250], []byte{9, 0, 0, 0})
//	copy(data[250:259], []byte("AppUse"))
//	copy(data[259:267], []byte("EscSeq"))
//
//	err := ear.Unmarshal(data)
//	assert.NoError(t, err)
//	assert.Equal(t, uint16(1), ear.OwnerIdentifier)
//	assert.Equal(t, uint16(2), ear.GroupIdentifier)
//	assert.Equal(t, uint16(3), ear.Permissions)
//	assert.Equal(t, "20230101000000", string(ear.FileCreationDate[:]))
//	assert.Equal(t, "20230102000000", string(ear.FileModificationDate[:]))
//	assert.Equal(t, "20230103000000", string(ear.FileExpirationDate[:]))
//	assert.Equal(t, "20230104000000", string(ear.FileEffectiveDate[:]))
//	assert.Equal(t, uint8(4), ear.RecordFormat)
//	assert.Equal(t, uint8(5), ear.RecordAttributes)
//	assert.Equal(t, uint32(6), ear.RecordLength)
//	assert.Equal(t, "SystemUseIdentifier", string(ear.SystemUseIdentifier[:]))
//	assert.Equal(t, "SystemUse", string(ear.SystemUse[:]))
//	assert.Equal(t, uint8(7), ear.ExtendedAttributeRecordVersion)
//	assert.Equal(t, uint8(8), ear.LengthOfEscapeSequences)
//	assert.Equal(t, "Unused1", string(ear.Unused1[:]))
//	assert.Equal(t, uint32(9), ear.LengthOfApplicationUse)
//	assert.Equal(t, "AppUse", string(ear.ApplicationUse))
//	assert.Equal(t, "EscSeq", string(ear.EscapeSequences))
//}

func TestUnmarshal_InvalidDataLength(t *testing.T) {
	logger := logr.Discard()
	ear := NewExtendedAttributeRecord(logger)
	data := make([]byte, 255)

	err := ear.Unmarshal(data)
	assert.Error(t, err)
	assert.Equal(t, "invalid data length", err.Error())
}

func TestUnmarshal_ApplicationUseOutOfRange(t *testing.T) {
	logger := logr.Discard()
	ear := NewExtendedAttributeRecord(logger)
	data := make([]byte, 256)
	copy(data[246:250], []byte{10, 0, 0, 0})

	err := ear.Unmarshal(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "applicationUse slice out of range")
}

func TestUnmarshal_EscapeSequencesOutOfRange(t *testing.T) {
	logger := logr.Discard()
	ear := NewExtendedAttributeRecord(logger)
	data := make([]byte, 256)
	copy(data[246:250], []byte{5, 0, 0, 0})
	copy(data[250:255], []byte("AppUse"))
	data[181] = 10

	err := ear.Unmarshal(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "escapeSequences slice out of range")
}
