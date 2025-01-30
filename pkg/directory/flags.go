package directory

import "fmt"

type FileFlags struct {
	Existence      bool
	Directory      bool
	AssociatedFile bool
	Record         bool
	Protection     bool
	Unused1        bool
	Unused2        bool
	MultiExtent    bool
}

func (ff *FileFlags) Set(flags uint8) {
	ff.Existence = flags&0x01 > 0
	ff.Directory = flags&0x02 > 0
	ff.AssociatedFile = flags&0x04 > 0
	ff.Record = flags&0x08 > 0
	ff.Protection = flags&0x10 > 0
	ff.Unused1 = flags&0x20 > 0
	ff.Unused2 = flags&0x40 > 0
	ff.MultiExtent = flags&0x80 > 0
}

func (ff *FileFlags) String() string {
	// Print out the flags in a human-readable format.
	return fmt.Sprintf("Existence=%t, Directory=%t, Associated File=%t, Record=%t, Protection=%t, Multi-Extent=%t",
		ff.Existence,
		ff.Directory,
		ff.AssociatedFile,
		ff.Record,
		ff.Protection,
		ff.MultiExtent)
}
