package validation

import (
	"fmt"
	"github.com/bgrewell/iso-kit/pkg/consts"
	"strings"
)

// validateByAllowedChars is a generic helper function that checks if every character in s
// is contained in the allowed set. The setName is used in error messages.
// It also ensures that each rune is within the UCS-2 range.
func validateByAllowedChars(s, allowed, setName string) error {
	for i, r := range s {
		if r > 0xFFFF {
			return fmt.Errorf("invalid %s-character at index %d: code point 0x%X is outside UCS-2 range", setName, i, r)
		}
		if !strings.ContainsRune(allowed, r) {
			return fmt.Errorf("invalid %s-character at index %d: %q is not allowed", setName, i, r)
		}
	}
	return nil
}

// ValidateACharacters checks that every character in the input string is one of the allowed A_CHARACTERS.
// If allowSeparators is true, it also permits the ISO9660 separator characters.
func ValidateACharacters(s string, allowSeparators bool) error {
	allowedChars := consts.A_CHARACTERS
	if allowSeparators {
		allowedChars += consts.ISO9660_SEPARATOR_1 + consts.ISO9660_SEPARATOR_2
	}
	return validateByAllowedChars(s, allowedChars, "A")
}

// ValidateDCharacters checks that every character in the input string is one of the allowed D_CHARACTERS.
// If allowSeparators is true, it also permits the ISO9660 separator characters.
func ValidateDCharacters(s string, allowSeparators bool) error {
	allowedChars := consts.D_CHARACTERS
	if allowSeparators {
		allowedChars += consts.ISO9660_SEPARATOR_1 + consts.ISO9660_SEPARATOR_2
	}
	return validateByAllowedChars(s, allowedChars, "D")
}

// isValidCCharacter returns true if the rune is valid according to the C-character set rules.
// Disallowed are control characters (0x0000 - 0x001F) and specific punctuation.
// It also ensures the rune is within the UCS-2 range.
func isValidCCharacter(r rune) bool {
	if r > 0xFFFF { // Outside UCS-2 range.
		return false
	}
	// Disallow control characters (0x0000 - 0x001F)
	if r <= 0x1F {
		return false
	}
	// Disallow specific code points: '*' (0x2A), '/' (0x2F), ':' (0x3A),
	// ';' (0x3B), '?' (0x3F), '\' (0x5C)
	switch r {
	case 0x2A, 0x2F, 0x3A, 0x3B, 0x3F, 0x5C:
		return false
	}
	return true
}

// ValidateCCharacters checks that every character in the input string is allowed in the C-character set,
// including an extra check that the character is within the UCS-2 range.
func ValidateCCharacters(s string) error {
	for i, r := range s {
		if r > 0xFFFF {
			return fmt.Errorf("invalid C-character at index %d: code point 0x%X is outside UCS-2 range", i, r)
		}
		if !isValidCCharacter(r) {
			return fmt.Errorf("invalid C-character at index %d: disallowed code point 0x%04X", i, r)
		}
	}
	return nil
}

// ValidateA1Characters validates the input string for the A1-character set,
// which is currently identical to the C-character set.
func ValidateA1Characters(s string) error {
	return ValidateCCharacters(s)
}
