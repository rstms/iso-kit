package validation

import (
	"github.com/bgrewell/iso-kit/pkg/consts"
	"regexp"
	"strings"
)

func ValidISO9660FileIdentifier(identifier string) bool {
	// Special seperator characters 0x2E(.) and 0x3B (;)are allowed
	return validateIdentifierRune(identifier, ".;")
}

func ValidISO9660DirIdentifier(identifier string) bool {
	// Special identifiers 0x00 (root) or 0x01 (parent) are allowed
	if len(identifier) == 1 && (identifier[0] == 0x00 || identifier[0] == 0x01) {
		return true
	}

	return validateIdentifierRune(identifier, "")
}

// validateIdentifierRune checks each rune in the identifier to ensure it is in either allowed constant.
func validateIdentifierRune(identifier string, additionalChars string) bool {
	// Combine both allowed sets.
	allowed := consts.D_CHARACTERS + consts.D1_CHARACTERS + additionalChars
	for _, r := range identifier {
		if !strings.ContainsRune(allowed, r) {
			return false
		}
	}
	return true
}

// Precompile the regular expression using both allowed character sets.
var allowedRegexp = regexp.MustCompile(`^[` + regexp.QuoteMeta(consts.D_CHARACTERS+consts.D1_CHARACTERS) + `]+$`)

// validateIdentifierRegex uses a regular expression to validate the identifier.
func validateIdentifierRegex(id string) bool {
	return allowedRegexp.MatchString(id)
}
