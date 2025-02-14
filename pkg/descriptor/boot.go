package descriptor

import (
	"github.com/bgrewell/iso-kit/pkg/consts"
	"io"
)

const (
	// Boot System Use Size is the size of a sector minus 71 bytes
	BOOT_SYSTEM_USE_SIZE = consts.ISO9660_SECTOR_SIZE - 71
)

type BootRecordDescriptor struct {
	VolumeDescriptorHeader
	BootRecordBody
}

type BootRecordBody struct {
	// Boot System Identifier specifies and identification of a system which can recognize and act upon the contents of
	// the Boot Identifier and Boot System Use fields in the Boot Record. (a-characters)
	BootSystemIdentifier string `json:"boot_system_identifier"`
	// Boot Identifier shall specify an identification of the boot system specified in the Boot System Use field of the
	// Boot Record. (a-characters)
	BootIdentifier string `json:"boot_identifier"`
	// Boot System Use is a byte field that is used by the boot system specified by the identifier.
	BootSystemUse [BOOT_SYSTEM_USE_SIZE]byte `json:"boot_system_use"`
}

func (d *BootRecordDescriptor) Marshal() ([consts.ISO9660_SECTOR_SIZE]byte, error) {
	return [consts.ISO9660_SECTOR_SIZE]byte{}, nil
}

func (d *BootRecordDescriptor) Unmarshal(data [consts.ISO9660_SECTOR_SIZE]byte, isoFile io.ReaderAt) error {
	return nil
}
