package descriptor

import (
	"github.com/bgrewell/iso-kit/pkg/consts"
	"io"
)

const (
	// Terminator resrved size is the size of a sector minus 7 bytes
	TERMINATOR_RESERVED_SIZE = consts.ISO9660_SECTOR_SIZE - 7
)

type VolumeDescriptorSetTerminator struct {
	VolumeDescriptorHeader
	VolumeDescriptorSetTerminatorBody
}

type VolumeDescriptorSetTerminatorBody struct {
	// Reserved for future standardization
	Reserved [TERMINATOR_RESERVED_SIZE]byte `json:"reserved"`
}

func (d *VolumeDescriptorSetTerminator) Marshal() ([consts.ISO9660_SECTOR_SIZE]byte, error) {
	return [consts.ISO9660_SECTOR_SIZE]byte{}, nil
}

func (d *VolumeDescriptorSetTerminator) Unmarshal(data [consts.ISO9660_SECTOR_SIZE]byte, isoFile io.ReaderAt) error {
	return nil
}
