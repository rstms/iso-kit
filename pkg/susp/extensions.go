package susp

type ExtensionRecord struct {
	Version    int
	Identifier string
	Descriptor string
	Source     string
}

// SUSP-112 5.1
type ContinuationEntry struct {
	blockLocation uint32
	offset        uint32
	lengthOfArea  uint32
}
