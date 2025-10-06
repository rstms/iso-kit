package descriptor

import (
	"bytes"
	"encoding/binary"
	"github.com/rstms/iso-kit/pkg/iso9660/directory"
	"github.com/rstms/iso-kit/pkg/iso9660/encoding"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestPrimaryVolumeDescriptorBody_MarshalUnmarshal(t *testing.T) {
	t.Run("happy path with real DirectoryRecord", func(t *testing.T) {
		// 1) Create a minimal valid DirectoryRecord that produces exactly 34 bytes when Marshaled.
		dr := &directory.DirectoryRecord{
			FileIdentifier: "\x00", // Minimal special identifier
			FileFlags: directory.FileFlags{
				Directory: true,
			},
			RecordingDateAndTime: time.Now(),
		}
		// Confirm it's 34 bytes:
		drBytes, err := dr.Marshal()
		require.NoError(t, err)
		require.Equal(t, 34, len(drBytes), "DirectoryRecord must marshal to 34 bytes")

		// 2) Create a PrimaryVolumeDescriptorBody with that DirectoryRecord.
		pvdb := &PrimaryVolumeDescriptorBody{
			UnusedField1:         0x01,
			SystemIdentifier:     "SYS_ID",
			VolumeIdentifier:     "VOL_ID",
			VolumeSpaceSize:      12345,
			VolumeSetSize:        7,
			VolumeSequenceNumber: 1,
			LogicalBlockSize:     2048,
			PathTableSize:        4096,
			RootDirectoryRecord:  dr, // <-- Our actual DirectoryRecord
			// Demonstrate setting a date/time field:
			VolumeCreationDateAndTime: time.Date(2025, time.January, 2, 3, 4, 5, 0, time.UTC),
			FileStructureVersion:      1,
		}

		// 3) Marshal the PVD
		data, err := pvdb.Marshal()
		require.NoError(t, err, "Marshal should succeed for fully populated PVD")
		require.Equal(t, PRIMARY_VOLUME_DESCRIPTOR_BODY_SIZE, len(data),
			"Marshalled PVD must be 2041 bytes")

		// 4) Unmarshal into a new PVD struct
		var pvdb2 PrimaryVolumeDescriptorBody
		// We must provide a real DirectoryRecord here as well, so it can Unmarshal into it.
		pvdb2.RootDirectoryRecord = &directory.DirectoryRecord{}

		err = pvdb2.Unmarshal(data[:])
		require.NoError(t, err, "Unmarshal should succeed for valid data")

		// 5) Check that the second PVD has matching fields
		require.Equal(t, pvdb.UnusedField1, pvdb2.UnusedField1)
		require.Equal(t, "SYS_ID", pvdb2.SystemIdentifier)
		require.Equal(t, "VOL_ID", pvdb2.VolumeIdentifier)
		require.Equal(t, uint32(12345), pvdb2.VolumeSpaceSize)
		require.Equal(t, uint16(7), pvdb2.VolumeSetSize)
		require.Equal(t, uint16(1), pvdb2.VolumeSequenceNumber)
		require.Equal(t, uint16(2048), pvdb2.LogicalBlockSize)
		require.Equal(t, uint32(4096), pvdb2.PathTableSize)
		require.Equal(t, pvdb.VolumeCreationDateAndTime, pvdb2.VolumeCreationDateAndTime)
		require.Equal(t, pvdb.FileStructureVersion, pvdb2.FileStructureVersion)
	})

	t.Run("marshal fails when RootDirectoryRecord is nil", func(t *testing.T) {
		pvdb := &PrimaryVolumeDescriptorBody{
			RootDirectoryRecord: nil,
		}
		_, err := pvdb.Marshal()
		require.Error(t, err)
		require.Contains(t, err.Error(), "rootDirectoryRecord is nil")
	})

	t.Run("unmarshal fails on short data", func(t *testing.T) {
		shortData := make([]byte, PRIMARY_VOLUME_DESCRIPTOR_BODY_SIZE-1) // 1 byte too short
		var pvdb PrimaryVolumeDescriptorBody
		// Provide a real DR so we don't panic on nil
		pvdb.RootDirectoryRecord = &directory.DirectoryRecord{}

		err := pvdb.Unmarshal(shortData)
		require.Error(t, err)
		require.Contains(t, err.Error(), "data too short")
	})
}

func TestPrimaryVolumeDescriptorBody_RoundTripRawBytes(t *testing.T) {
	// 1) Create a valid PVD with a real DirectoryRecord
	dr := &directory.DirectoryRecord{
		FileIdentifier:       "\x00",
		RecordingDateAndTime: time.Now(),
	}
	drBytes, err := dr.Marshal()
	require.NoError(t, err)
	require.Equal(t, 34, len(drBytes))

	pvdb := &PrimaryVolumeDescriptorBody{
		RootDirectoryRecord: dr,
		SystemIdentifier:    "XYZ",
	}

	// 2) Marshal it
	data, err := pvdb.Marshal()
	require.NoError(t, err)

	// 3) Unmarshal into a second struct
	var pvdb2 PrimaryVolumeDescriptorBody
	pvdb2.RootDirectoryRecord = &directory.DirectoryRecord{}

	err = pvdb2.Unmarshal(data[:])
	require.NoError(t, err)

	// 4) Marshal the second struct
	data2, err := pvdb2.Marshal()
	require.NoError(t, err)

	// 5) Compare data vs data2
	require.True(t, bytes.Equal(data[:], data2[:]),
		"Marshalled bytes should match on re-serialization")
}

func TestPrimaryVolumeDescriptorBody_PreservesTrailingSpaces(t *testing.T) {
	// Build a full 2041-byte raw array that includes:
	//  - offset=0: 1 byte for pvdb.UnusedField1
	//  - offset=1..32 (32 bytes) for pvdb.SystemIdentifier (with trailing spaces)
	//  - offset=33..64 (32 bytes) for pvdb.VolumeIdentifier (with trailing spaces)
	//  - offset=... (we fill minimal data for the rest)
	//  - offset=149..182 (34 bytes) for RootDirectoryRecord
	//  - The rest is zeroed or minimal.
	//
	// We'll then Unmarshal -> Marshal and expect the same data if the code
	// truly preserves trailing spaces exactly.

	require.Equal(t, PRIMARY_VOLUME_DESCRIPTOR_BODY_SIZE, 2041)

	// 1) Create the raw 2041-byte slice and fill the relevant parts.
	original := make([]byte, PRIMARY_VOLUME_DESCRIPTOR_BODY_SIZE)

	offset := 0

	// (1) UnusedField1 = 0x01
	original[offset] = 0x01
	offset++

	// (2) SystemIdentifier: 32 bytes.
	//    We'll do "SYS_ID" + 10 spaces, then the rest up to 32 total with spaces.
	//    So the first 6 chars are 'S','Y','S','_','I','D', then 10 spaces => 16 bytes total,
	//    we'll fill all 32 with spaces anyway, but the first 6 are the visible text.
	sysID := []byte("SYS_ID          ") // 6 chars + 10 spaces = 16 bytes
	for len(sysID) < 32 {
		sysID = append(sysID, ' ')
	}
	copy(original[offset:offset+32], sysID)
	offset += 32

	// (3) VolumeIdentifier: 32 bytes.
	//    "VOL_ID" + 5 spaces => 11 bytes, fill to 32 with spaces.
	volID := []byte("VOL_ID     ") // 6 chars + 5 spaces = 11
	for len(volID) < 32 {
		volID = append(volID, ' ')
	}
	copy(original[offset:offset+32], volID)
	offset += 32

	// (4) UnusedField2: 8 bytes => fill with 0 for minimal
	offset += 8

	// (5) volumeSpaceSize: 8 bytes => we can store a valid BothByteOrders32 => e.g. 12345
	a := encoding.MarshalBothByteOrders32(12345)
	copy(original[offset:offset+8], a[:])
	offset += 8

	// (6) unusedField3: 32 bytes => zero
	offset += 32

	// (7) volumeSetSize: 4 bytes => store e.g. 7
	b := encoding.MarshalBothByteOrders16(7)
	copy(original[offset:offset+4], b[:])
	offset += 4

	// (8) volumeSequenceNumber: 4 bytes => store e.g. 1
	c := encoding.MarshalBothByteOrders16(1)
	copy(original[offset:offset+4], c[:])
	offset += 4

	// (9) logicalBlockSize: 4 bytes => store e.g. 2048
	d := encoding.MarshalBothByteOrders16(2048)
	copy(original[offset:offset+4], d[:])
	offset += 4

	// (10) pathTableSize: 8 bytes => store e.g. 4096
	e := encoding.MarshalBothByteOrders32(4096)
	copy(original[offset:offset+8], e[:])
	offset += 8

	// (11) locationOfTypeLPathTable: 4 bytes, little-endian => e.g. 0x11223344
	binary.LittleEndian.PutUint32(original[offset:offset+4], 0x11223344)
	offset += 4

	// (12) locationOfOptionalTypeLPathTable: 4 bytes => e.g. 0x55667788
	binary.LittleEndian.PutUint32(original[offset:offset+4], 0x55667788)
	offset += 4

	// (13) locationOfTypeMPathTable: 4 bytes, big-endian => e.g. 0xAABBCCDD
	binary.BigEndian.PutUint32(original[offset:offset+4], 0xAABBCCDD)
	offset += 4

	// (14) locationOfOptionalTypeMPathTable: 4 bytes => e.g. 0xEEFF0011
	binary.BigEndian.PutUint32(original[offset:offset+4], 0xEEFF0011)
	offset += 4

	// (15) rootDirectoryRecord: 34 bytes
	// We'll create a minimal 34-byte DirectoryRecord. For example:
	//  - Byte 0 => record length = 34
	//  - Byte 1 => extended attribute length = 0
	//  - Next 8 => locationOfExtent (both byte orders)
	//  - Next 8 => dataLength
	//  - Next 7 => date/time
	//  - Next 1 => fileFlags
	//  - Next 1 => fileUnitSize
	//  - Next 1 => interleaveGapSize
	//  - Next 4 => volumeSequenceNumber
	//  - Next 1 => lengthOfFileIdentifier
	//  - Next N => fileIdentifier
	//  - Possibly a padding byte
	//
	// For simplicity, let's cheat by marshaling a known minimal DirectoryRecord.
	// We only need *something* valid. We'll do "FileIdentifier = '\x00'".
	//
	dirRec := &directory.DirectoryRecord{
		FileIdentifier:       "\x00",
		FileFlags:            directory.FileFlags{},
		RecordingDateAndTime: time.Now(),
	}
	dirData, err := dirRec.Marshal()
	require.NoError(t, err)
	require.Equal(t, 34, len(dirData), "Must be exactly 34 bytes")
	copy(original[offset:offset+34], dirData)
	offset += 34

	// We'll skip filling the rest of the fields in detail, but do ensure we fill up the entire buffer
	// with zeros so the offset finishes at 2041.
	// The next fields in sequence:
	// 16. volumeSetIdentifier (128 bytes)
	// 17. publisherIdentifier (128 bytes)
	// 18. dataPreparerIdentifier (128 bytes)
	// 19. applicationIdentifier (128 bytes)
	// 20. copyrightFileIdentifier (37 bytes)
	// 21. abstractFileIdentifier (37 bytes)
	// 22. bibliographicFileIdentifier (37 bytes)
	// 23-26. date/time fields (4 * 17 bytes = 68)
	// 27. fileStructureVersion (1 byte)
	// 28. reservedField1 (1 byte)
	// 29. applicationUse (512 bytes)
	// 30. reservedField2 (653 bytes)
	// We can just do offset += those sizes, leaving zeroed memory.

	offset += 128 // volumeSetIdentifier
	offset += 128 // publisherIdentifier
	offset += 128 // dataPreparerIdentifier
	offset += 128 // applicationIdentifier
	offset += 37  // copyrightFileIdentifier
	offset += 37  // abstractFileIdentifier
	offset += 37  // bibliographicFileIdentifier
	offset += 17  // volumeCreationDateAndTime
	offset += 17  // volumeModificationDateAndTime
	offset += 17  // volumeExpirationDateAndTime
	offset += 17  // volumeEffectiveDateAndTime
	offset++      // fileStructureVersion
	offset++      // reservedField1
	offset += 512 // applicationUse
	offset += 653 // reservedField2

	require.Equal(t, PRIMARY_VOLUME_DESCRIPTOR_BODY_SIZE, offset,
		"Should end exactly at 2041 bytes")

	// 2) Unmarshal into a fresh PVD
	var pvd PrimaryVolumeDescriptorBody
	pvd.RootDirectoryRecord = &directory.DirectoryRecord{
		RecordingDateAndTime: time.Date(2025, time.January, 2, 3, 4, 5, 0, time.UTC),
	}
	err = pvd.Unmarshal(original)
	require.NoError(t, err, "Unmarshal should succeed for a well-formed PVD")

	// 3) Marshal again
	newBytes, err := pvd.Marshal()
	require.NoError(t, err, "Re-marshal should succeed")

	// 4) Compare original vs new. We expect them to match exactly.
	//    If the code has changed trailing spaces or did not preserve them,
	//    this will fail.
	require.Equal(t, original, newBytes,
		"Round-trip data should match, preserving trailing spaces exactly")

	// This test will FAIL if your code discards the original trailing spaces
	// or re-injects a different pattern of spaces when re-marshal happens.
}
