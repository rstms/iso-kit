package descriptor

import (
	"github.com/bgrewell/iso-kit/pkg/consts"
	"io"
)

const (
	// Partition System Use Size is the size of a sector minus 88 bytes
	PARTITION_SYSTEM_USE_SIZE = consts.ISO9660_SECTOR_SIZE - 88
)

type VolumePartitionDescriptor struct {
	VolumeDescriptorHeader
	VolumePartitionDescriptorBody
}

type VolumePartitionDescriptorBody struct {
	// Unused field should always be 0x00
	UnusedField1 byte `json:"unusedField1"`
	// System Identifier specifies a system which can recognize and act upon the content of the Logical Sectors within
	// logical Sector Numbers 0 to 15 of the volume.
	//  | (a-characters)
	SystemIdentifier string `json:"system_identifier"`
	// Volume Partition Identifier specifies an identification of the Volume Partition.
	//  | (d-characters)
	VolumePartitionIdentifier string `json:"volume_partition_identifier"`
	// Volume Partition Location specifies the number of Logical Block Number of the first Logical Block allocated to
	// the Volume Partition
	//  | Encoding: BothByteOrder
	VolumePartitionLocation uint32 `json:"volume_partition_location"`
	// Volume Partition Size specifies the number of Logical Blocks in which the Volume Partition is recorded.
	//  | Encoding: BothByteOrder
	VolumePartitionSize uint32 `json:"volume_partition_size"`
	// System Use Area
	SystemUse [PARTITION_SYSTEM_USE_SIZE]byte `json:"system_use"`
}

func (d *VolumePartitionDescriptor) Marshal() ([consts.ISO9660_SECTOR_SIZE]byte, error) {
	return [consts.ISO9660_SECTOR_SIZE]byte{}, nil
}

func (d *VolumePartitionDescriptor) Unmarshal(data [consts.ISO9660_SECTOR_SIZE]byte, isoFile io.ReaderAt) error {
	return nil
}
