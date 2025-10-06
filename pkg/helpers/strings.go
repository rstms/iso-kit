package helpers

import "github.com/rstms/iso-kit/pkg/consts"

func PadString(s string, length int) []byte {
	b := make([]byte, length)
	copy(b, s)
	for i := len(s); i < length; i++ {
		b[i] = consts.ISO9660_FILLER
	}
	return b
}
