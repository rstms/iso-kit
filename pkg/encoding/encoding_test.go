// encoding_test.go
package encoding

import (
	"bytes"
	"encoding/binary"
	"io"
	"strings"
	"testing"
	"time"
)

// TestMarshalString verifies that MarshalString properly truncates or pads a string.
func TestMarshalString(t *testing.T) {
	// Case 1: input shorter than pad length → pads with spaces.
	s := "hello"
	result := MarshalString(s, 10)
	expected := "hello     "
	if got := string(result); got != expected {
		t.Errorf("MarshalString(%q, 10) = %q; want %q", s, got, expected)
	}

	// Case 2: input exactly the pad length → no padding.
	s = "12345"
	result = MarshalString(s, 5)
	expected = "12345"
	if got := string(result); got != expected {
		t.Errorf("MarshalString(%q, 5) = %q; want %q", s, got, expected)
	}

	// Case 3: input longer than pad length → truncates.
	s = "Hello, World!"
	result = MarshalString(s, 5)
	expected = "Hello"
	if got := string(result); got != expected {
		t.Errorf("MarshalString(%q, 5) = %q; want %q", s, got, expected)
	}

	// Edge: pad length zero returns an empty byte slice.
	s = "anything"
	result = MarshalString(s, 0)
	if len(result) != 0 {
		t.Errorf("MarshalString(%q, 0) returned non-empty result: %q", s, string(result))
	}
}

// --- UnmarshalInt32LSBMSB & UnmarshalUint32LSBMSB Tests ---

// TestUnmarshalInt32LSBMSB_Positive tests a valid 32-bit integer decoding.
func TestUnmarshalInt32LSBMSB_Positive(t *testing.T) {
	var buf [8]byte
	value := int32(12345678)
	// Create 8 bytes where both representations encode the same value.
	binary.LittleEndian.PutUint32(buf[0:4], uint32(value))
	binary.BigEndian.PutUint32(buf[4:8], uint32(value))

	result, err := UnmarshalInt32LSBMSB(buf[:])
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != value {
		t.Errorf("Expected %d, got %d", value, result)
	}
}

// TestUnmarshalInt32LSBMSB_Negative tests error conditions for UnmarshalInt32LSBMSB.
func TestUnmarshalInt32LSBMSB_Negative(t *testing.T) {
	// Test with insufficient data.
	data := []byte{0, 1, 2, 3, 4, 5, 6} // Only 7 bytes.
	_, err := UnmarshalInt32LSBMSB(data)
	if err != io.ErrUnexpectedEOF {
		t.Errorf("Expected error %v for insufficient data, got %v", io.ErrUnexpectedEOF, err)
	}

	// Test with mismatched little- and big-endian representations.
	var buf [8]byte
	binary.LittleEndian.PutUint32(buf[0:4], uint32(100))
	binary.BigEndian.PutUint32(buf[4:8], uint32(101))
	_, err = UnmarshalInt32LSBMSB(buf[:])
	if err == nil {
		t.Errorf("Expected error for mismatched values, got nil")
	}
}

// TestUnmarshalUint32LSBMSB_Positive tests the unsigned version.
func TestUnmarshalUint32LSBMSB_Positive(t *testing.T) {
	var buf [8]byte
	value := uint32(98765432)
	binary.LittleEndian.PutUint32(buf[0:4], value)
	binary.BigEndian.PutUint32(buf[4:8], value)

	result, err := UnmarshalUint32LSBMSB(buf[:])
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != value {
		t.Errorf("Expected %d, got %d", value, result)
	}
}

// TestUnmarshalUint32LSBMSB_Negative verifies error conditions.
func TestUnmarshalUint32LSBMSB_Negative(t *testing.T) {
	// Insufficient data.
	data := []byte{0, 1, 2, 3, 4, 5, 6}
	_, err := UnmarshalUint32LSBMSB(data)
	if err != io.ErrUnexpectedEOF {
		t.Errorf("Expected error %v for insufficient data, got %v", io.ErrUnexpectedEOF, err)
	}

	// Mismatched values.
	var buf [8]byte
	binary.LittleEndian.PutUint32(buf[0:4], uint32(200))
	binary.BigEndian.PutUint32(buf[4:8], uint32(201))
	_, err = UnmarshalUint32LSBMSB(buf[:])
	if err == nil {
		t.Errorf("Expected error for mismatched values, got nil")
	}
}

// --- UnmarshalInt16LSBMSB Tests ---

// TestUnmarshalInt16LSBMSB_Positive tests a valid 16-bit integer decoding.
func TestUnmarshalInt16LSBMSB_Positive(t *testing.T) {
	var buf [4]byte
	value := int16(12345)
	binary.LittleEndian.PutUint16(buf[0:2], uint16(value))
	binary.BigEndian.PutUint16(buf[2:4], uint16(value))

	result, err := UnmarshalInt16LSBMSB(buf[:])
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != value {
		t.Errorf("Expected %d, got %d", value, result)
	}
}

// TestUnmarshalInt16LSBMSB_Negative tests error conditions for 16-bit decoding.
func TestUnmarshalInt16LSBMSB_Negative(t *testing.T) {
	// Test with insufficient data.
	data := []byte{0, 1, 2} // Only 3 bytes.
	_, err := UnmarshalInt16LSBMSB(data)
	if err != io.ErrUnexpectedEOF {
		t.Errorf("Expected error %v for insufficient data, got %v", io.ErrUnexpectedEOF, err)
	}

	// Test with mismatched little- and big-endian representations.
	var buf [4]byte
	binary.LittleEndian.PutUint16(buf[0:2], uint16(300))
	binary.BigEndian.PutUint16(buf[2:4], uint16(301))
	_, err = UnmarshalInt16LSBMSB(buf[:])
	if err == nil {
		t.Errorf("Expected error for mismatched values, got nil")
	}
}

// --- WriteInt32LSBMSB Tests ---

// TestWriteInt32LSBMSB_Positive verifies that WriteInt32LSBMSB writes correctly.
func TestWriteInt32LSBMSB_Positive(t *testing.T) {
	buf := make([]byte, 8)
	value := int32(54321)
	WriteInt32LSBMSB(buf, value)

	// Now decode the written bytes.
	result, err := UnmarshalInt32LSBMSB(buf)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != value {
		t.Errorf("Expected %d, got %d", value, result)
	}
}

// TestWriteInt32LSBMSB_Negative verifies that WriteInt32LSBMSB panics when given a too-short slice.
func TestWriteInt32LSBMSB_Negative(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic due to insufficient slice length")
		}
	}()
	buf := make([]byte, 7) // Too short for 8 bytes.
	WriteInt32LSBMSB(buf, 123)
}

// --- WriteInt16LSBMSB Tests ---

// TestWriteInt16LSBMSB_Positive verifies correct writing for a 16-bit integer.
func TestWriteInt16LSBMSB_Positive(t *testing.T) {
	buf := make([]byte, 4)
	value := int16(1234)
	WriteInt16LSBMSB(buf, value)

	result, err := UnmarshalInt16LSBMSB(buf)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != value {
		t.Errorf("Expected %d, got %d", value, result)
	}
}

// TestWriteInt16LSBMSB_Negative verifies that WriteInt16LSBMSB panics when the destination slice is too short.
func TestWriteInt16LSBMSB_Negative(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic due to insufficient slice length")
		}
	}()
	buf := make([]byte, 3) // Too short for 4 bytes.
	WriteInt16LSBMSB(buf, 1234)
}

// --- DecodeDirectoryTime Tests ---

// TestDecodeDirectoryTime_Positive tests decoding of a valid directory time.
func TestDecodeDirectoryTime_Positive(t *testing.T) {
	// Create valid data:
	// Year: 2020 → 2020-1900 = 120; Month: 5; Day: 15; Hour: 12; Minute: 34; Second: 56; Offset: 0.
	data := []byte{120, 5, 15, 12, 34, 56, 0}
	result, err := DecodeDirectoryTime(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Validate components.
	if result.Year() != 2020 || result.Month() != 5 || result.Day() != 15 ||
		result.Hour() != 12 || result.Minute() != 34 || result.Second() != 56 {
		t.Errorf("Decoded time mismatch: got %v", result)
	}
	// Check the time zone offset is 0.
	_, offsetSeconds := result.Zone()
	if offsetSeconds != 0 {
		t.Errorf("Expected GMT offset 0 seconds, got %d seconds", offsetSeconds)
	}
}

// TestDecodeDirectoryTime_Negative tests various invalid inputs.
func TestDecodeDirectoryTime_Negative(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		errMsg string
	}{
		{"Insufficient length", []byte{120, 5, 15, 12, 34, 56}, "invalid data length"},
		{"Invalid month", []byte{120, 0, 15, 12, 34, 56, 0}, "invalid month"},
		{"Invalid day", []byte{120, 5, 0, 12, 34, 56, 0}, "invalid day"},
		{"Invalid hour", []byte{120, 5, 15, 24, 34, 56, 0}, "invalid hour"},
		{"Invalid minute", []byte{120, 5, 15, 12, 60, 56, 0}, "invalid minute"},
		{"Invalid second", []byte{120, 5, 15, 12, 34, 60, 0}, "invalid second"},
		// For offset: we want an int8 value out of the acceptable range (-48 to 52).
		// To produce -49, we can store 207 (since 207-256 = -49).
		{"Invalid GMT offset", []byte{120, 5, 15, 12, 34, 56, 207}, "invalid GMT offset"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeDirectoryTime(tt.data)
			if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("For %s, expected error containing %q; got %v", tt.name, tt.errMsg, err)
			}
		})
	}
}

// --- EncodeDirectoryTime Tests ---

// TestEncodeDirectoryTime_Positive tests encoding a valid time.
func TestEncodeDirectoryTime_Positive(t *testing.T) {
	// Use a valid time: 2020-05-15 12:34:56 in a UTC zone (offset 0).
	loc := time.FixedZone("UTC", 0)
	tme := time.Date(2020, 5, 15, 12, 34, 56, 0, loc)
	data, err := EncodeDirectoryTime(tme)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []byte{
		byte(2020 - 1900), // Year
		byte(5),           // Month
		byte(15),          // Day
		byte(12),          // Hour
		byte(34),          // Minute
		byte(56),          // Second
		0,                 // Offset (0/15=0)
	}
	if !bytes.Equal(data, expected) {
		t.Errorf("Expected %v, got %v", expected, data)
	}
}

// TestEncodeDirectoryTime_Negative tests error conditions for encoding.
func TestEncodeDirectoryTime_Negative(t *testing.T) {
	// Case 1: Year too low (year < 1900).
	loc := time.FixedZone("UTC", 0)
	tme := time.Date(1800, 1, 1, 0, 0, 0, 0, loc)
	_, err := EncodeDirectoryTime(tme)
	if err == nil || !strings.Contains(err.Error(), "year out of range") {
		t.Errorf("Expected error for year out of range, got %v", err)
	}

	// Case 2: Year too high (year - 1900 > 255).
	tme = time.Date(2200, 1, 1, 0, 0, 0, 0, loc)
	_, err = EncodeDirectoryTime(tme)
	if err == nil || !strings.Contains(err.Error(), "year out of range") {
		t.Errorf("Expected error for year out of range, got %v", err)
	}

	// Case 3: GMT offset out of range.
	// Create a time with a huge offset. For example, 1440 minutes offset gives offset = 96 (1440/15),
	// which is above the maximum allowed 52.
	hugeOffsetSeconds := 1440 * 60
	loc = time.FixedZone("HugeOffset", hugeOffsetSeconds)
	tme = time.Date(2020, 1, 1, 0, 0, 0, 0, loc)
	_, err = EncodeDirectoryTime(tme)
	if err == nil || !strings.Contains(err.Error(), "GMT offset out of range") {
		t.Errorf("Expected error for GMT offset out of range, got %v", err)
	}
}
