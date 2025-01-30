package descriptor

import (
	"errors"
	"github.com/bgrewell/iso-kit/pkg/consts"
	"github.com/bgrewell/iso-kit/pkg/logging"
	"github.com/bgrewell/iso-kit/pkg/path"
)

// VolumeDescriptorType represents the type of volume descriptor in the ISO 9660 standard.
type VolumeDescriptorType byte

const (
	// VolumeDescriptorBootRecord indicates a Boot Record (type 0).
	VolumeDescriptorBootRecord VolumeDescriptorType = 0x00

	// VolumeDescriptorPrimary indicates a Primary Volume Descriptor (type 1).
	VolumeDescriptorPrimary VolumeDescriptorType = 0x01

	// VolumeDescriptorSupplementary indicates a Supplementary Volume Descriptor (type 2).
	VolumeDescriptorSupplementary VolumeDescriptorType = 0x02

	// VolumeDescriptorPartition indicates a Partition Volume Descriptor (type 3).
	VolumeDescriptorPartition VolumeDescriptorType = 0x03

	// VolumeDescriptorSetTerminator indicates the Volume Descriptor Set Terminator (type 255).
	VolumeDescriptorSetTerminator VolumeDescriptorType = 0xFF
)

func ParseVolumeDescriptor(data []byte) (VolumeDescriptor, error) {
	logging.Logger().Trace("Parsing volume descriptor")
	vd := &volumeDescriptor{}
	if err := vd.Unmarshal(data); err != nil {
		logging.Logger().Errorf("Failed to unmarshal volume descriptor: %v", err)
		return nil, err
	}
	logging.Logger().Trace("Successfully parsed volume descriptor")
	return vd, nil
}

type VolumeDescriptor interface {
	Type() VolumeDescriptorType
	Identifier() string
	Version() int8
	PathTableLocation() uint32
	PathTableSize() int32
	PathTable() *[]*path.PathTableRecord
	Data() [consts.ISO9660_SECTOR_SIZE]byte
}

type volumeDescriptor struct {
	vdType     VolumeDescriptorType
	identifier string
	version    int8
	data       [consts.ISO9660_SECTOR_SIZE]byte
}

func (vd *volumeDescriptor) Type() VolumeDescriptorType {
	return vd.vdType
}

func (vd *volumeDescriptor) Identifier() string {
	return vd.identifier
}

func (vd *volumeDescriptor) Version() int8 {
	return vd.version
}

func (vd *volumeDescriptor) Data() [consts.ISO9660_SECTOR_SIZE]byte {
	return vd.data
}

func (vd *volumeDescriptor) PathTableLocation() uint32 {
	return 0
}

func (vd *volumeDescriptor) PathTableSize() int32 {
	return 0
}

func (vd *volumeDescriptor) PathTable() *[]*path.PathTableRecord {
	return nil
}

func (vd *volumeDescriptor) Unmarshal(data []byte) error {
	if len(data) < 8 {
		return errors.New("invalid data length")
	}

	vd.vdType = VolumeDescriptorType(data[0])
	vd.identifier = string(data[1:5])
	vd.version = int8(data[5])
	copy(vd.data[:], data[:])

	return nil
}
