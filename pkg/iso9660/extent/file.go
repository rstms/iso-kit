package extent

import (
	"fmt"
	"github.com/bgrewell/iso-kit/pkg/consts"
	"github.com/bgrewell/iso-kit/pkg/iso9660/info"
	"io"
)

type FileExtent struct {
	// --- This struct is just a concept and not defined in the ISO9660 standard ---
	FileIdentifier string `json:"file_identifier"`
	Joliet         bool   `json:"joliet"`
	LocationOfFile uint32 `json:"location_of_file"`
	SizeOfFile     uint32 `json:"size_of_file"`
	Reader         io.ReaderAt
}

func (f FileExtent) Type() string {
	return "File Extent"
}

func (f FileExtent) Name() string {
	return f.FileIdentifier
}

func (f FileExtent) Description() string {
	return ""
}

func (f FileExtent) Properties() map[string]interface{} {
	return map[string]interface{}{
		"LocationOfFile": f.LocationOfFile,
		"SizeOfFile":     f.SizeOfFile,
	}
}

func (f FileExtent) Offset() int64 {
	return int64(f.LocationOfFile * consts.ISO9660_SECTOR_SIZE)
}

func (f FileExtent) Size() int {
	return int(f.SizeOfFile)
}

func (f FileExtent) GetObjects() []info.ImageObject {
	return []info.ImageObject{f}
}

func (f FileExtent) Marshal() ([]byte, error) {
	// Allocate a buffer of the file's size
	buf := make([]byte, f.SizeOfFile)

	// Read from the Reader at the specified offset
	n, err := f.Reader.ReadAt(buf, f.Offset())
	if err != nil {
		return nil, fmt.Errorf("failed to read file extent %s: %w", f.FileIdentifier, err)
	}

	// Ensure we read the expected number of bytes
	if uint32(n) != f.SizeOfFile {
		return nil, fmt.Errorf("unexpected read size for %s: got %d, expected %d", f.FileIdentifier, n, f.SizeOfFile)
	}

	return buf, nil
}
