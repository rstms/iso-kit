package validation

import "testing"

// Benchmark for the rune-based validation.
func BenchmarkValidateIdentifierRune(b *testing.B) {
	// Use an example valid identifier.
	id := "HELLO123_-456"
	for i := 0; i < b.N; i++ {
		if !validateIdentifierRune(id) {
			b.Fatal("rune validation failed for valid identifier")
		}
	}
}

// Benchmark for the regex-based validation.
func BenchmarkValidateIdentifierRegex(b *testing.B) {
	// Use the same example valid identifier.
	id := "HELLO123_-456"
	for i := 0; i < b.N; i++ {
		if !validateIdentifierRegex(id) {
			b.Fatal("regex validation failed for valid identifier")
		}
	}
}
