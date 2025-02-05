package path

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
)

func TestNewPathTableRecord(t *testing.T) {
	logger := logr.Discard()
	ptr := NewPathTableRecord(logger)
	assert.NotNil(t, ptr)
	assert.Equal(t, logger, ptr.logger)
}

func TestPathTableRecord_Unmarshal_ValidData(t *testing.T) {
	logger := logr.Discard()
	ptr := NewPathTableRecord(logger)
	data := []byte{
		5, 0, // DirectoryIdentifierLength, ExtendedAttributeRecordLength
		1, 0, 0, 0, // LocationOfExtent
		2, 0, // ParentDirectoryNumber
		'a', 'b', 'c', 'd', 'e', // DirectoryIdentifier
	}

	err := ptr.Unmarshal(data)
	assert.NoError(t, err)
	assert.Equal(t, byte(5), ptr.DirectoryIdentifierLength)
	assert.Equal(t, byte(0), ptr.ExtendedAttributeRecordLength)
	assert.Equal(t, uint32(1), ptr.LocationOfExtent)
	assert.Equal(t, uint16(2), ptr.ParentDirectoryNumber)
	assert.Equal(t, "abcde", ptr.DirectoryIdentifier)
	assert.Equal(t, []byte{0x00}, ptr.Padding)
}

func TestPathTableRecord_Unmarshal_InvalidDataLength(t *testing.T) {
	logger := logr.Discard()
	ptr := NewPathTableRecord(logger)
	data := []byte{1, 2, 3}

	err := ptr.Unmarshal(data)
	assert.Error(t, err)
	assert.Equal(t, "invalid data length", err.Error())
}

func TestPathTableRecord_Unmarshal_DirectoryIdentifierOutOfRange(t *testing.T) {
	logger := logr.Discard()
	ptr := NewPathTableRecord(logger)
	data := []byte{
		10, 0, // DirectoryIdentifierLength, ExtendedAttributeRecordLength
		1, 0, 0, 0, // LocationOfExtent
		2, 0, // ParentDirectoryNumber
		'a', 'b', 'c', 'd', 'e', // DirectoryIdentifier (incomplete)
	}

	err := ptr.Unmarshal(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "directory identifier out of range")
}
