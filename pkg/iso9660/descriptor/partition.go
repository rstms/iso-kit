package descriptor

import (
	"fmt"
	"github.com/rstms/iso-kit/pkg/consts"
	"github.com/rstms/iso-kit/pkg/iso9660/directory"
	"github.com/rstms/iso-kit/pkg/iso9660/info"
	"github.com/rstms/iso-kit/pkg/logging"
	"time"
)

const (
	// Partition System Use Size is the size of a sector minus 88 bytes
	PARTITION_SYSTEM_USE_SIZE = consts.ISO9660_SECTOR_SIZE - 88
)

type VolumePartitionDescriptor struct {
	VolumeDescriptorHeader
	VolumePartitionDescriptorBody
}

func (d *VolumePartitionDescriptor) DescriptorType() VolumeDescriptorType {
	return TYPE_PARTITION_DESCRIPTOR
}

func (d *VolumePartitionDescriptor) VolumeIdentifier() string {
	//TODO implement me
	panic("implement me")
}

func (d *VolumePartitionDescriptor) SystemIdentifier() string {
	//TODO implement me
	panic("implement me")
}

func (d *VolumePartitionDescriptor) VolumeSetIdentifier() string {
	//TODO implement me
	panic("implement me")
}

func (d *VolumePartitionDescriptor) PublisherIdentifier() string {
	//TODO implement me
	panic("implement me")
}

func (d *VolumePartitionDescriptor) DataPreparerIdentifier() string {
	//TODO implement me
	panic("implement me")
}

func (d *VolumePartitionDescriptor) ApplicationIdentifier() string {
	//TODO implement me
	panic("implement me")
}

func (d *VolumePartitionDescriptor) CopyrightFileIdentifier() string {
	//TODO implement me
	panic("implement me")
}

func (d *VolumePartitionDescriptor) AbstractFileIdentifier() string {
	//TODO implement me
	panic("implement me")
}

func (d *VolumePartitionDescriptor) BibliographicFileIdentifier() string {
	//TODO implement me
	panic("implement me")
}

func (d *VolumePartitionDescriptor) VolumeCreationDateTime() time.Time {
	//TODO implement me
	panic("implement me")
}

func (d *VolumePartitionDescriptor) VolumeModificationDateTime() time.Time {
	//TODO implement me
	panic("implement me")
}

func (d *VolumePartitionDescriptor) VolumeExpirationDateTime() time.Time {
	//TODO implement me
	panic("implement me")
}

func (d *VolumePartitionDescriptor) VolumeEffectiveDateTime() time.Time {
	//TODO implement me
	panic("implement me")
}

func (d *VolumePartitionDescriptor) HasJoliet() bool {
	//TODO implement me
	panic("implement me")
}

func (d *VolumePartitionDescriptor) HasRockRidge() bool {
	//TODO implement me
	panic("implement me")
}

func (d *VolumePartitionDescriptor) RootDirectory() *directory.DirectoryRecord {
	//TODO implement me
	panic("implement me")
}

func (d *VolumePartitionDescriptor) GetObjects() []info.ImageObject {
	return []info.ImageObject{d}
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
	// --- Fields that are not part of the ISO9660 object ---
	// Object Location (in bytes)
	ObjectLocation int64 `json:"object_location"`
	// Object Size (in bytes)
	ObjectSize uint32 `json:"object_size"`
	// Logger
	Logger *logging.Logger
}

func (v VolumePartitionDescriptorBody) Type() string {
	return "Volume Descriptor"
}

func (v VolumePartitionDescriptorBody) Name() string {
	return "Volume Partition Descriptor"
}

func (v VolumePartitionDescriptorBody) Description() string {
	return fmt.Sprintf("%s: %s", v.SystemIdentifier, v.VolumePartitionIdentifier)
}

func (v VolumePartitionDescriptorBody) Properties() map[string]interface{} {
	return map[string]interface{}{
		"VolumePartitionLocation": v.VolumePartitionLocation,
		"VolumePartitionSize":     v.VolumePartitionSize,
	}
}

func (v VolumePartitionDescriptorBody) Offset() int64 {
	return v.ObjectLocation
}

func (v VolumePartitionDescriptorBody) Size() int {
	return int(v.ObjectSize)
}

func (d *VolumePartitionDescriptor) Marshal() ([]byte, error) {
	return []byte{}, nil
}

func (d *VolumePartitionDescriptor) Unmarshal(data [consts.ISO9660_SECTOR_SIZE]byte) error {
	return nil
}
