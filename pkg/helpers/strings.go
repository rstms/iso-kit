package helpers

func PadString(s string, length int) []byte {
	b := make([]byte, length)
	copy(b, s)
	for i := len(s); i < length; i++ {
		b[i] = ' '
	}
	return b
}
