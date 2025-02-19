package iso

import (
	"errors"
	"github.com/bgrewell/iso-kit/pkg/filesystem"
	"github.com/bgrewell/iso-kit/pkg/iso9660"
	"github.com/bgrewell/iso-kit/pkg/option"
	"github.com/bgrewell/iso-kit/pkg/udf"
	"io"
	"os"
	"time"
)

// ISO represents a generic ISO filesystem with read/write capabilities.
type ISO interface {
	GetVolumeID() string
	GetSystemID() string
	GetVolumeSetID() string
	GetPublisherID() string
	GetDataPreparerID() string
	GetApplicationID() string
	GetCopyrightID() string
	GetAbstractID() string
	GetBibliographicID() string
	GetCreationDateTime() time.Time
	GetModificationDateTime() time.Time
	GetExpirationDateTime() time.Time
	GetEffectiveDateTime() time.Time

	GetVolumeSize() uint32
	RootDirectoryLocation() uint32

	ListBootEntries() ([]*filesystem.FileSystemEntry, error)
	ListFiles() ([]*filesystem.FileSystemEntry, error)
	ListDirectories() ([]*filesystem.FileSystemEntry, error)
	ReadFile(path string) ([]byte, error)
	AddFile(path string, data []byte) error
	RemoveFile(path string) error
	CreateDirectories(path string) error
	Extract(path string) error

	HasJoliet() bool
	HasRockRidge() bool
	HasElTorito() bool

	Save(writer io.Writer) error
	Close() error
}

func Open(filename string, opts ...option.OpenOption) (ISO, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	// Read PVD header at sector 16 (offset 32768)
	var header [6]byte
	if _, err = f.ReadAt(header[:], 16*2048); err != nil {
		f.Close()
		return nil, err
	}

	// Detect ISO9660
	if string(header[1:6]) == "CD001" {
		return iso9660.Open(f, opts...)
	}

	// Read UDF anchor volume descriptor at sector 256 (offset 524288)
	if _, err = f.ReadAt(header[:], 256*2048); err == nil {
		if string(header[1:5]) == "BEA01" {
			return udf.Open(f, opts...)
		}
	}

	return nil, errors.New("unsupported ISO format")
}
