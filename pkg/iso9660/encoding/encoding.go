package encoding

import (
	"encoding/binary"
	"fmt"
	"time"
)

// MarshalBothByteOrders32 converts a uint32 value into an 8-byte field that
// encodes the value in both little‑endian and big‑endian orders.
// The resulting byte order is: (yz, wx, uv, st, st, uv, wx, yz),
// where (st uv wx yz) is the hexadecimal representation of the value.
func MarshalBothByteOrders32(val uint32) [8]byte {
	var data [8]byte
	// First four bytes: little-endian representation.
	binary.LittleEndian.PutUint32(data[0:4], val)
	// Last four bytes: big-endian representation.
	binary.BigEndian.PutUint32(data[4:8], val)
	return data
}

// UnmarshalBothByteOrders32 converts an 8-byte field encoded in both little‑
// and big‑endian orders back to a uint32 value. It verifies that both halves
// are equal. If they are not, it returns an error.
func UnmarshalBothByteOrders32(data [8]byte) (uint32, error) {
	// Decode little-endian value from the first four bytes.
	little := binary.LittleEndian.Uint32(data[0:4])
	// Decode big-endian value from the last four bytes.
	big := binary.BigEndian.Uint32(data[4:8])
	if little != big {
		return 0, fmt.Errorf("mismatched both-byte orders: little-endian value %d != big-endian value %d", little, big)
	}
	return little, nil
}

// MarshalBothByteOrders16 converts a uint16 value into a 4-byte field that
// encodes the value in both little‑endian and big‑endian orders.
// The resulting field has the layout: (yz, wx, wx, yz), where (wx, yz) is the
// hexadecimal representation of the value.
// For example, for the value 0x1234, it returns [0x34, 0x12, 0x12, 0x34].
func MarshalBothByteOrders16(val uint16) [4]byte {
	var data [4]byte
	// First two bytes: little-endian representation.
	binary.LittleEndian.PutUint16(data[0:2], val)
	// Next two bytes: big-endian representation.
	binary.BigEndian.PutUint16(data[2:4], val)
	return data
}

// UnmarshalBothByteOrders16 converts a 4-byte field encoded in both little‑
// and big‑endian orders back to a uint16 value. It verifies that both halves
// match; if they do not, it returns an error.
func UnmarshalBothByteOrders16(data [4]byte) (uint16, error) {
	// Read the little-endian value from the first two bytes.
	little := binary.LittleEndian.Uint16(data[0:2])
	// Read the big-endian value from the last two bytes.
	big := binary.BigEndian.Uint16(data[2:4])
	if little != big {
		return 0, fmt.Errorf("mismatched both-byte orders: little-endian value %d != big-endian value %d", little, big)
	}
	return little, nil
}

// MarshalDateTime converts a time.Time into a 17-byte field following ISO9660 8.4.26.1.
// The first 16 bytes contain ASCII digits in the format:
//
//	YYYY MM DD hh mm ss cc
//
// and the 17th byte is the time zone offset (in 15-minute intervals) as a signed integer.
// Note: This format is used in Volume Descriptors
func MarshalDateTime(t time.Time) ([17]byte, error) {
	var b [17]byte

	year, month, day := t.Date()
	hour, minute, second := t.Clock()
	// Calculate hundredths of a second (each hundredth = 10,000,000 ns)
	hundredths := t.Nanosecond() / 10000000

	// Format date and time into a 16-character string.
	// For example: "20230327140509" plus hundredths "45" → "2023032714050945"
	dtStr := fmt.Sprintf("%04d%02d%02d%02d%02d%02d%02d",
		year, int(month), day, hour, minute, second, hundredths)
	if len(dtStr) != 16 {
		return b, fmt.Errorf("formatted date/time length is not 16: got %d", len(dtStr))
	}
	copy(b[:16], dtStr)

	// Determine the time zone offset in seconds.
	_, offsetSec := t.Zone()
	// Convert offset to number of 15-minute intervals.
	offset15 := int8(offsetSec / (15 * 60))
	// Validate offset range: must be between -48 and +52.
	if offset15 < -48 || offset15 > 52 {
		return b, fmt.Errorf("time zone offset %d (in 15-minute intervals: %d) is out of allowed range", offsetSec, offset15)
	}
	// Set the 17th byte to the offset.
	b[16] = byte(offset15)
	return b, nil
}

// UnmarshalDateTime converts a 17-byte ISO9660 date/time field into a time.Time.
// It expects the first 16 bytes to be ASCII digits representing
// YYYY MM DD hh mm ss cc, and the 17th byte as the offset in 15-minute intervals.
// Note: This format is used in Volume Descriptors
func UnmarshalDateTime(data [17]byte) (time.Time, error) {
	// Extract the date/time string from the first 16 bytes.
	dtStr := string(data[:16])
	if len(dtStr) != 16 {
		return time.Time{}, fmt.Errorf("data length for date/time is not 16: got %d", len(dtStr))
	}

	// Parse the fields using fixed-width substrings.
	var (
		year, month, day     int
		hour, minute, second int
		hundredths           int
	)
	// Instead of using Sscanf (which can be tricky with fixed widths),
	// we parse the substrings directly.
	_, err := fmt.Sscanf(dtStr, "%4d%2d%2d%2d%2d%2d%2d",
		&year, &month, &day, &hour, &minute, &second, &hundredths)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse date/time string %q: %w", dtStr, err)
	}

	// Convert hundredths of a second into nanoseconds.
	nsec := hundredths * 10000000

	// The 17th byte is the time zone offset in 15-minute intervals.
	offset15 := int8(data[16])
	offsetSec := int(offset15) * 15 * 60

	// Create a fixed location for the time zone.
	loc := time.FixedZone("ISO9660", offsetSec)
	// Construct the time.Time value.
	t := time.Date(year, time.Month(month), day, hour, minute, second, nsec, loc)
	return t, nil
}

// MarshalRecordingDateTime converts a time.Time into a 7-byte field according
// to Table 9 – Recording Date and Time. It returns an error if the year is out of range.
// Note: All fields are stored as numerical values (not ASCII digits).
// Note: This type format is used in DirectoryRecords
func MarshalRecordingDateTime(t time.Time) ([7]byte, error) {
	var b [7]byte

	year, month, day := t.Date()
	hour, minute, second := t.Clock()

	// The field stores the number of years since 1900, so valid years are 1900–2155.
	if year < 1900 || year > 2155 {
		return b, fmt.Errorf("year %d out of range for Recording Date and Time (must be between 1900 and 2155)", year)
	}
	b[0] = byte(year - 1900)
	b[1] = byte(month) // month is 1-12
	b[2] = byte(day)
	b[3] = byte(hour)
	b[4] = byte(minute)
	b[5] = byte(second)

	// Get the time zone offset in seconds and convert to 15-minute intervals.
	_, offsetSec := t.Zone()
	offset15 := int8(offsetSec / (15 * 60))
	if offset15 < -48 || offset15 > 52 {
		return b, fmt.Errorf("time zone offset %d (in 15-minute intervals: %d) is out of allowed range", offsetSec, offset15)
	}
	// Store offset as a signed numerical value.
	b[6] = byte(offset15)
	return b, nil
}

// UnmarshalRecordingDateTime converts a 7-byte Recording Date and Time field into a time.Time.
// The fields are interpreted as follows:
//
//	Byte 1: years since 1900,
//	Byte 2: month (1-12),
//	Byte 3: day,
//	Byte 4: hour,
//	Byte 5: minute,
//	Byte 6: second,
//	Byte 7: offset from GMT in 15-minute intervals (as a signed value).
//
// If all seven bytes are zero, it indicates that the date/time are not specified.
// Note: This type format is used in DirectoryRecords
func UnmarshalRecordingDateTime(b [7]byte) (time.Time, error) {
	// If all fields are zero, return the zero time.
	allZero := true
	for _, v := range b {
		if v != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		return time.Time{}, nil
	}

	year := int(b[0]) + 1900
	month := time.Month(b[1])
	day := int(b[2])
	hour := int(b[3])
	minute := int(b[4])
	second := int(b[5])
	// b[6] is stored as a byte but represents a signed 8-bit integer.
	offset15 := int8(b[6])
	offsetSec := int(offset15) * 15 * 60

	loc := time.FixedZone("ISO9660", offsetSec)
	return time.Date(year, month, day, hour, minute, second, 0, loc), nil
}
