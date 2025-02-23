package xattr

import (
	"encoding/binary"
	"fmt"
	"github.com/bgrewell/iso-kit/pkg/helpers"
	"github.com/bgrewell/iso-kit/pkg/iso9660/encoding"
	"github.com/bgrewell/iso-kit/pkg/iso9660/info"
	"time"
)

const (
	// Record System Use Size is 64 bytes
	RECORD_SYSTEM_USE_SIZE = 64
)

type ExtendedAttributeRecord struct {
	// Owner Identification specifies the identification of the file owner who is a member of the group identified by
	// the Group Identification field of the Extended Attribute Record specified as a 16-bit number
	//  | Encoding: BothByteOrder
	OwnerIdentification uint16 `json:"owner_identification"`
	// Group Identification specifies the identification of the group which owns the file.
	//  | Encoding: BothByteOrder
	GroupIdentification uint16 `json:"group_identification"`
	// Permissions is a 16-bit flag field
	//  0 - If set to 0, shall mean that an owner who is a member of a group of the System class of user
	//      may read the file. If set to 1, shall mean that an owner who is a member of a group of the System class of
	//      user may not read the file
	//  1 - Shall be set to 1
	//  2 - If set to 0, shall mean that an owner who is a member of a group of the System class of user may execute the
	//      file. If set to 1, shall mean that an owner who is a member of a group of the System class of user may not
	//      execute the file.
	//  3 - Shall be set to 1
	//  4 - If set to 0, shall mean that the owner may read the file. If set to 0, shall mean that the owner may not
	//      read the file.
	//  5 - Shall be set to 1
	//  6 - If set to 0, shall mean that the owner may execute the file. If set to 1, shall mean that the owner may not
	//      execute the file.
	//  7 - Shall be set to 1
	//  8 - If set to 0, shall mean that any user who is a member of the group specified by the Group Identification
	//      field may read the file. If set to 2, shall mean that of the users who are members of the group specified by
	//      the Group Identification field, only the owner may read the file.
	//  9 - Shall be set to 1
	// 10 - If set to 0, shall mean that any user who is a member of the group specified by the Group Identification
	//      field may execute the file. If set to 1, shall mean that of the users who are members of the group specified
	//      by the Group Identification field, only the owner may execute the file.
	// 11 - Shall be set to 1
	// 12 - If set to 0, shall mean that any user may read the file. If set to 1, shall mean that a user not a member of
	//      the group specified by the Group Identification field may not read the file.
	// 13 - Shall be set to 1
	// 14 - If set to 0, shall mean that any user may execute the file. If set to 1, shall mean that a user not a member
	//      of the group specified by the Group Identification field may not execute the file.
	// 15 - Shall be set to 1
	Permissions ExtendedAttrPermissions `json:"permissions"`
	// File Creation Date and Time specifies the date and time of the day at which the information in the file was
	// created.
	//  | Encoding: 17-byte date/time
	FileCreationDateAndTime time.Time `json:"file_creation_date_and_time"`
	// File Modification Date and Time specifies the date and time of the day which the file was last modified.
	//  | Encoding: 17-byte date/time
	FileModificationDateAndTime time.Time `json:"file_modification_date_and_time"`
	// File Expiration Date and Time specifies the date and time of the day at which the file is to be considered
	// obsolete.
	//  | Encoding: 17-byte date/time
	FileExpirationDateAndTime time.Time `json:"file_expiration_date_and_time"`
	// File Effective Date and Time specifies the date and time of the day at which the file is to be considered
	// effective.
	//  | Encoding: 17-byte date/time
	FileEffectiveDateAndTime time.Time `json:"file_effective_date_and_time"`
	// Record Format specifies the format of the information in the file. 0 shall mean that the information is recorded
	// in an unspecified format. 1 shall mean that the information in the file is a sequence of fixed-length records
	// (see EMCA-119 6.10.3). 2 shall mean that the information in the file is a sequence of variable-length records
	// (see ECMA-119 6.10.4) in which the RCW is recorded according to 7.2.1. 3 shall mean that the information in the
	// file is a sequence of variable-length records in which the RCW is recorded according to 7.2.2. Numbers 4-127 are
	// reserved for future standardization. Numbers 128-255 are reserved for system use.
	RecordFormat uint8 `json:"record_format"`
	// Record Attributes contains an 8-bit number specifying certain processing of the records in a file when they are
	// displayed on a character-imaging device.
	//  0 - Means that each record shall be preceded by a LINE FEED character and followed by a CARRIAGE RETURN.
	//  1 - Means that the first byte of a record shall be interpreted as specified in ISO 1539 for vertical spacing.
	//  2 - Means that the record contains the necessary control information.
	// 3-255 - Reserved for future standardization.
	RecordAttributes uint8 `json:"record_attributes"`
	// Record Length specifies different things based on the value of Record Format. If Record Format is 0 then the
	// Record Length field shall be set to 0. If Record Format is 1 then the Record Length field shall be set to the
	// length in bytes of each record in the file.
	// If Record Format is 2 or 3 then the Record Length field shall specify the maximum length in bytes of a record in
	// the file.
	//  | Encoding: BothByteOrder
	RecordLength uint16 `json:"record_length"`
	// System Identifier specifies an identification of a system which can recognize and act upon the content of the
	// System Use fields in the Extended Attribute Record and associated Directory Record.
	// (a-characters or a1-characters)
	SystemIdentifier string `json:"system_identifier"`
	// System Use specifies an area that is reserved for system use. The contents of this field are not interpreted by
	// the standard.
	SystemUse [RECORD_SYSTEM_USE_SIZE]byte `json:"system_use"`
	// Extended Attribute Record Version specifies the version of the Extended Attribute Record.
	// Should always be 1.
	ExtendedAttributeRecordVersion uint8 `json:"extended_attribute_record_version"`
	// Length Of Escape Sequences specifies the length in bytes of the Escape Sequences field in the Extended Attribute
	// Record.
	LengthOfEscapeSequences uint8 `json:"length_of_escape_sequences"`
	// Reserved For Future Use BP 183 - 246 all bytes should be 0x00
	ReservedForFutureUse [64]byte `json:"reserved_for_future_use"`
	// Length of Application Use specifies the length in bytes of the Application Use field in the Extended Attribute
	// Record.
	//  | Encoding: BothByteOrder
	LengthOfApplicationUse uint16 `json:"length_of_application_use"`
	// Application Use specifies an area that is reserved for application use. The contents of this field are not
	// interpreted by the standard. BP 251 to (250 - LEN-AU)
	// Note: Be careful to make a copy of the bytes when unmarshalling to avoid a slice pointing to the original buffer.
	//       which could be modified, especially if the buffer is reused such as when looping over a slice of records.
	ApplicationUse []byte `json:"application_use"`
	// Escape Sequences shall be optional. If present, this field shall contain escape sequences that designate the
	// coded character sets to be used to interpret the contents of the file. These escape sequences shall conform to
	// ISO 2022, except that the ESCAPE character shall be omitted from each escape sequence. The first or only escape
	// sequence shall begin at the first byte of the field. Each successive escape sequence shall begin at the byte in
	// the field immediately following the last byte of the preceding escape sequence. Any unused positions following
	// the last escape sequence shall be set to (00). BP (251 + LEN_AU) to (250 + LEN_ESC + LEN_AU
	EscapeSequences []byte `json:"escape_sequences"`
	// --- Fields that are not part of the ISO9660 object ---
	// Object Location (in bytes)
	ObjectLocation int64 `json:"object_location"`
	// Object Size (in bytes)
	ObjectSize uint32 `json:"object_size"`
}

func (ear *ExtendedAttributeRecord) Type() string {
	return "Extended Attribute Record"
}

func (ear *ExtendedAttributeRecord) Name() string {
	return "Extended Attribute Record"
}

func (ear *ExtendedAttributeRecord) Description() string {
	return ""
}

func (ear *ExtendedAttributeRecord) Properties() map[string]interface{} {
	return map[string]interface{}{}
}

func (ear *ExtendedAttributeRecord) Offset() int64 {
	return ear.ObjectLocation
}

func (ear *ExtendedAttributeRecord) Size() int {
	return int(ear.RecordLength)
}

func (ear *ExtendedAttributeRecord) GetObjects() []info.ImageObject {
	return []info.ImageObject{ear}
}

// Marshal converts the ExtendedAttributeRecord into its on‑disk byte representation.
func (ear *ExtendedAttributeRecord) Marshal() ([]byte, error) {
	var buf []byte

	// 1. OwnerIdentification (Both-byte orders, 4 bytes)
	ownerBytes := encoding.MarshalBothByteOrders16(ear.OwnerIdentification)
	buf = append(buf, ownerBytes[:]...)

	// 2. GroupIdentification (Both-byte orders, 4 bytes)
	groupBytes := encoding.MarshalBothByteOrders16(ear.GroupIdentification)
	buf = append(buf, groupBytes[:]...)

	// 3. Permissions (2 bytes, encoded here in big-endian)
	permVal := ear.Permissions.Marshal() // uint16
	permBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(permBytes, permVal)
	buf = append(buf, permBytes...)

	// 4. File Creation Date and Time (17 bytes)
	fcdBytes, err := encoding.MarshalDateTime(ear.FileCreationDateAndTime)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal FileCreationDateAndTime: %w", err)
	}
	buf = append(buf, fcdBytes[:]...)

	// 5. File Modification Date and Time (17 bytes)
	fmdBytes, err := encoding.MarshalDateTime(ear.FileModificationDateAndTime)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal FileModificationDateAndTime: %w", err)
	}
	buf = append(buf, fmdBytes[:]...)

	// 6. File Expiration Date and Time (17 bytes)
	fedBytes, err := encoding.MarshalDateTime(ear.FileExpirationDateAndTime)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal FileExpirationDateAndTime: %w", err)
	}
	buf = append(buf, fedBytes[:]...)

	// 7. File Effective Date and Time (17 bytes)
	fefBytes, err := encoding.MarshalDateTime(ear.FileEffectiveDateAndTime)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal FileEffectiveDateAndTime: %w", err)
	}
	buf = append(buf, fefBytes[:]...)

	// 8. RecordFormat (1 byte)
	buf = append(buf, ear.RecordFormat)

	// 9. RecordAttributes (1 byte)
	buf = append(buf, ear.RecordAttributes)

	// 10. RecordLength (Both-byte orders, 4 bytes)
	recLenBytes := encoding.MarshalBothByteOrders16(ear.RecordLength)
	buf = append(buf, recLenBytes[:]...)

	// 11. systemIdentifier: fixed 32 bytes (pad/truncate)
	sysIDBytes := helpers.PadString(ear.SystemIdentifier, 32)
	buf = append(buf, sysIDBytes...)

	// 12. SystemUse: fixed-length field
	buf = append(buf, ear.SystemUse[:]...)

	// 13. ExtendedAttributeRecordVersion (1 byte)
	buf = append(buf, ear.ExtendedAttributeRecordVersion)

	// 14. LengthOfEscapeSequences (1 byte)
	buf = append(buf, ear.LengthOfEscapeSequences)

	// 15. ReservedForFutureUse (64 bytes)
	buf = append(buf, ear.ReservedForFutureUse[:]...)

	// 16. LengthOfApplicationUse (Both-byte orders, 4 bytes)
	appUseLenBytes := encoding.MarshalBothByteOrders16(ear.LengthOfApplicationUse)
	buf = append(buf, appUseLenBytes[:]...)

	// 17. applicationUse (variable length, must match LengthOfApplicationUse)
	if len(ear.ApplicationUse) != int(ear.LengthOfApplicationUse) {
		return nil, fmt.Errorf("application use length mismatch: expected %d, got %d", ear.LengthOfApplicationUse, len(ear.ApplicationUse))
	}
	buf = append(buf, ear.ApplicationUse...)

	// 18. EscapeSequences (variable length, must match LengthOfEscapeSequences)
	if len(ear.EscapeSequences) != int(ear.LengthOfEscapeSequences) {
		return nil, fmt.Errorf("escape sequences length mismatch: expected %d, got %d", ear.LengthOfEscapeSequences, len(ear.EscapeSequences))
	}
	buf = append(buf, ear.EscapeSequences...)

	return buf, nil
}

// Unmarshal decodes an ExtendedAttributeRecord from the provided byte slice.
// It expects the data to be at least as long as the fixed portion of the record.
// The variable-length fields (ApplicationUse and EscapeSequences) are determined by their respective length fields.
func (ear *ExtendedAttributeRecord) Unmarshal(data []byte) error {
	offset := 0

	// 1. OwnerIdentification (4 bytes)
	if offset+4 > len(data) {
		return fmt.Errorf("insufficient data for OwnerIdentification")
	}
	var ownerBytes [4]byte
	copy(ownerBytes[:], data[offset:offset+4])
	ownerID, err := encoding.UnmarshalUint16LSBMSB(ownerBytes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal OwnerIdentification: %w", err)
	}
	ear.OwnerIdentification = ownerID
	offset += 4

	// 2. GroupIdentification (4 bytes)
	if offset+4 > len(data) {
		return fmt.Errorf("insufficient data for GroupIdentification")
	}
	var groupBytes [4]byte
	copy(groupBytes[:], data[offset:offset+4])
	groupID, err := encoding.UnmarshalUint16LSBMSB(groupBytes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal GroupIdentification: %w", err)
	}
	ear.GroupIdentification = groupID
	offset += 4

	// 3. Permissions (2 bytes, big-endian)
	if offset+2 > len(data) {
		return fmt.Errorf("insufficient data for Permissions")
	}
	permVal := binary.BigEndian.Uint16(data[offset : offset+2])
	perms, err := UnmarshalExtendedAttrPermissions(permVal)
	if err != nil {
		return fmt.Errorf("failed to unmarshal Permissions: %w", err)
	}
	ear.Permissions = perms
	offset += 2

	// 4. FileCreationDateAndTime (17 bytes)
	if offset+17 > len(data) {
		return fmt.Errorf("insufficient data for FileCreationDateAndTime")
	}
	var fcdBytes [17]byte
	copy(fcdBytes[:], data[offset:offset+17])
	fcd, err := encoding.UnmarshalDateTime(fcdBytes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal FileCreationDateAndTime: %w", err)
	}
	ear.FileCreationDateAndTime = fcd
	offset += 17

	// 5. FileModificationDateAndTime (17 bytes)
	if offset+17 > len(data) {
		return fmt.Errorf("insufficient data for FileModificationDateAndTime")
	}
	var fmdBytes [17]byte
	copy(fmdBytes[:], data[offset:offset+17])
	fmd, err := encoding.UnmarshalDateTime(fmdBytes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal FileModificationDateAndTime: %w", err)
	}
	ear.FileModificationDateAndTime = fmd
	offset += 17

	// 6. FileExpirationDateAndTime (17 bytes)
	if offset+17 > len(data) {
		return fmt.Errorf("insufficient data for FileExpirationDateAndTime")
	}
	var fedBytes [17]byte
	copy(fedBytes[:], data[offset:offset+17])
	fed, err := encoding.UnmarshalDateTime(fedBytes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal FileExpirationDateAndTime: %w", err)
	}
	ear.FileExpirationDateAndTime = fed
	offset += 17

	// 7. FileEffectiveDateAndTime (17 bytes)
	if offset+17 > len(data) {
		return fmt.Errorf("insufficient data for FileEffectiveDateAndTime")
	}
	var fefBytes [17]byte
	copy(fefBytes[:], data[offset:offset+17])
	fef, err := encoding.UnmarshalDateTime(fefBytes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal FileEffectiveDateAndTime: %w", err)
	}
	ear.FileEffectiveDateAndTime = fef
	offset += 17

	// 8. RecordFormat (1 byte)
	if offset+1 > len(data) {
		return fmt.Errorf("insufficient data for RecordFormat")
	}
	ear.RecordFormat = data[offset]
	offset++

	// 9. RecordAttributes (1 byte)
	if offset+1 > len(data) {
		return fmt.Errorf("insufficient data for RecordAttributes")
	}
	ear.RecordAttributes = data[offset]
	offset++

	// 10. RecordLength (4 bytes, Both-byte orders for uint16)
	if offset+4 > len(data) {
		return fmt.Errorf("insufficient data for RecordLength")
	}
	var recLenBytes [4]byte
	copy(recLenBytes[:], data[offset:offset+4])
	recLen, err := encoding.UnmarshalUint16LSBMSB(recLenBytes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal RecordLength: %w", err)
	}
	ear.RecordLength = recLen
	offset += 4

	// 11. systemIdentifier: fixed 32 bytes.
	if offset+32 > len(data) {
		return fmt.Errorf("insufficient data for systemIdentifier")
	}
	ear.SystemIdentifier = string(data[offset : offset+32])
	offset += 32

	// 12. SystemUse: RECORD_SYSTEM_USE_SIZE bytes.
	if offset+RECORD_SYSTEM_USE_SIZE > len(data) {
		return fmt.Errorf("insufficient data for SystemUse")
	}
	copy(ear.SystemUse[:], data[offset:offset+RECORD_SYSTEM_USE_SIZE])
	offset += RECORD_SYSTEM_USE_SIZE

	// 13. ExtendedAttributeRecordVersion (1 byte)
	if offset+1 > len(data) {
		return fmt.Errorf("insufficient data for ExtendedAttributeRecordVersion")
	}
	ear.ExtendedAttributeRecordVersion = data[offset]
	offset++

	// 14. LengthOfEscapeSequences (1 byte)
	if offset+1 > len(data) {
		return fmt.Errorf("insufficient data for LengthOfEscapeSequences")
	}
	ear.LengthOfEscapeSequences = data[offset]
	offset++

	// 15. ReservedForFutureUse (64 bytes)
	if offset+64 > len(data) {
		return fmt.Errorf("insufficient data for ReservedForFutureUse")
	}
	copy(ear.ReservedForFutureUse[:], data[offset:offset+64])
	offset += 64

	// 16. LengthOfApplicationUse (4 bytes, Both-byte orders for uint16)
	if offset+4 > len(data) {
		return fmt.Errorf("insufficient data for LengthOfApplicationUse")
	}
	var appUseLenBytes [4]byte
	copy(appUseLenBytes[:], data[offset:offset+4])
	appUseLen, err := encoding.UnmarshalUint16LSBMSB(appUseLenBytes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal LengthOfApplicationUse: %w", err)
	}
	ear.LengthOfApplicationUse = appUseLen
	offset += 4

	// 17. applicationUse (variable length)
	if offset+int(ear.LengthOfApplicationUse) > len(data) {
		return fmt.Errorf("insufficient data for applicationUse: need %d, have %d", ear.LengthOfApplicationUse, len(data)-offset)
	}
	ear.ApplicationUse = make([]byte, ear.LengthOfApplicationUse)
	copy(ear.ApplicationUse, data[offset:offset+int(ear.LengthOfApplicationUse)])
	offset += int(ear.LengthOfApplicationUse)

	// 18. EscapeSequences (variable length, length given by LengthOfEscapeSequences)
	if offset+int(ear.LengthOfEscapeSequences) > len(data) {
		return fmt.Errorf("insufficient data for EscapeSequences: need %d, have %d", ear.LengthOfEscapeSequences, len(data)-offset)
	}
	ear.EscapeSequences = make([]byte, ear.LengthOfEscapeSequences)
	copy(ear.EscapeSequences, data[offset:offset+int(ear.LengthOfEscapeSequences)])
	offset += int(ear.LengthOfEscapeSequences)

	return nil
}
