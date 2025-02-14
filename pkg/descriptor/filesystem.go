package descriptor

import "github.com/bgrewell/iso-kit/pkg/consts"

// FileSystemVolumeDescriptor is an interface that represents the methods common to both the PrimaryVolumeDescriptor and
// the SupplementaryVolumeDescriptor. This is not a type found in ISO9660/ECMA119 spec.
type FileSystemVolumeDescriptor interface {
	Marshal() ([consts.ISO9660_SECTOR_SIZE]byte, error)
	Unmarshal([consts.ISO9660_SECTOR_SIZE]byte) error
}
