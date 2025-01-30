package encoding

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"time"
)

// MarshalString encodes the given string as a byte array padded to the given length
func MarshalString(s string, padToLength int) []byte {
	if len(s) > padToLength {
		s = s[:padToLength]
	}
	missingPadding := padToLength - len(s)
	s = s + strings.Repeat(" ", missingPadding)
	return []byte(s)
}

// UnmarshalInt32LSBMSB decodes a 32-bit integer in both byte orders, as defined in ECMA-119 7.3.3
func UnmarshalInt32LSBMSB(data []byte) (int32, error) {
	if len(data) < 8 {
		return 0, io.ErrUnexpectedEOF
	}

	lsb := int32(binary.LittleEndian.Uint32(data[0:4]))
	msb := int32(binary.BigEndian.Uint32(data[4:8]))

	if lsb != msb {
		return 0, fmt.Errorf("little-endian and big-endian value mismatch: %d != %d", lsb, msb)
	}

	return lsb, nil
}

// UnmarshalUint32LSBMSB is the same as UnmarshalInt32LSBMSB but returns an unsigned integer
func UnmarshalUint32LSBMSB(data []byte) (uint32, error) {
	n, err := UnmarshalInt32LSBMSB(data)
	return uint32(n), err
}

// UnmarshalInt16LSBMSB decodes a 16-bit integer in both byte orders, as defined in ECMA-119 7.3.3
func UnmarshalInt16LSBMSB(data []byte) (int16, error) {
	if len(data) < 4 {
		return 0, io.ErrUnexpectedEOF
	}

	lsb := int16(binary.LittleEndian.Uint16(data[0:2]))
	msb := int16(binary.BigEndian.Uint16(data[2:4]))

	if lsb != msb {
		return 0, fmt.Errorf("little-endian and big-endian value mismatch: %d != %d", lsb, msb)
	}

	return lsb, nil
}

// WriteInt32LSBMSB writes a 32-bit integer in both byte orders, as defined in ECMA-119 7.3.3
func WriteInt32LSBMSB(dst []byte, value int32) {
	_ = dst[7] // early bounds check to guarantee safety of writes below
	binary.LittleEndian.PutUint32(dst[0:4], uint32(value))
	binary.BigEndian.PutUint32(dst[4:8], uint32(value))
}

// WriteInt16LSBMSB writes a 16-bit integer in both byte orders, as defined in ECMA-119 7.2.3
func WriteInt16LSBMSB(dst []byte, value int16) {
	_ = dst[3] // early bounds check to guarantee safety of writes below
	binary.LittleEndian.PutUint16(dst[0:2], uint16(value))
	binary.BigEndian.PutUint16(dst[2:4], uint16(value))
}

// DecodeDirectoryTime converts a byte slice in the directory entry custom format into a Go time.Time struct.
func DecodeDirectoryTime(data []byte) (time.Time, error) {
	if len(data) != 7 {
		return time.Time{}, fmt.Errorf("invalid data length: expected 7 bytes, got %d", len(data))
	}

	// Extract components
	year := int(data[0]) + 1900
	month := time.Month(data[1])
	day := int(data[2])
	hour := int(data[3])
	minute := int(data[4])
	second := int(data[5])
	offset := int8(data[6]) // signed value for offset

	// Validate components
	if month < 1 || month > 12 {
		return time.Time{}, fmt.Errorf("invalid month: %d", month)
	}
	if day < 1 || day > 31 {
		return time.Time{}, fmt.Errorf("invalid day: %d", day)
	}
	if hour < 0 || hour > 23 {
		return time.Time{}, fmt.Errorf("invalid hour: %d", hour)
	}
	if minute < 0 || minute > 59 {
		return time.Time{}, fmt.Errorf("invalid minute: %d", minute)
	}
	if second < 0 || second > 59 {
		return time.Time{}, fmt.Errorf("invalid second: %d", second)
	}
	if offset < -48 || offset > 52 {
		return time.Time{}, fmt.Errorf("invalid GMT offset: %d", offset)
	}

	// Convert GMT offset to hours and minutes
	offsetMinutes := int(offset) * 15

	// Build Go time.Time struct
	location := time.FixedZone("CustomTimeZone", offsetMinutes*60)
	return time.Date(year, month, day, hour, minute, second, 0, location), nil
}

// EncodeDirectoryTime converts a Go time.Time struct into the custom directory entry time format.
func EncodeDirectoryTime(t time.Time) ([]byte, error) {
	year := t.Year() - 1900
	if year < 0 || year > 255 {
		return nil, fmt.Errorf("year out of range: %d", t.Year())
	}

	month := t.Month()
	day := t.Day()
	hour := t.Hour()
	minute := t.Minute()
	second := t.Second()

	// Calculate GMT offset in 15-minute intervals
	_, offsetSeconds := t.Zone()
	offsetMinutes := offsetSeconds / 60
	offset := offsetMinutes / 15
	if offset < -48 || offset > 52 {
		return nil, fmt.Errorf("GMT offset out of range: %d", offset)
	}

	// Build the byte slice
	return []byte{
		byte(year),
		byte(month),
		byte(day),
		byte(hour),
		byte(minute),
		byte(second),
		byte(offset),
	}, nil
}
