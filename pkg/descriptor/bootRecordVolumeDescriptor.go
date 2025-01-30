package descriptor

import (
	"errors"
	"github.com/bgrewell/iso-kit/pkg/consts"
	"github.com/bgrewell/iso-kit/pkg/logging"
	"strings"
)

func ParseBootRecordVolumeDescriptor(vd VolumeDescriptor) (*BootRecordVolumeDescriptor, error) {
	logging.Logger().Trace("Parsing boot record volume descriptor")
	brvd := &BootRecordVolumeDescriptor{}
	if err := brvd.Unmarshal(vd.Data()); err != nil {
		logging.Logger().Error(err, "Failed to unmarshal boot record volume descriptor")
		return nil, err
	}
	logging.Logger().Trace("Successfully parsed boot record volume descriptor")

	logging.Logger().Tracef("Volume descriptor type: %d", brvd.Type)
	if brvd.Type != VolumeDescriptorBootRecord {
		logging.Logger().Warnf("Invalid boot record volume descriptor: %d", brvd.Type)
	}

	logging.Logger().Tracef("Standard identifier: %s", brvd.StandardIdentifier)
	if brvd.StandardIdentifier != consts.ISO9660_STD_IDENTIFIER {
		logging.Logger().Warnf("Invalid standard identifier: %s, expected: %s", brvd.StandardIdentifier, consts.ISO9660_STD_IDENTIFIER)
	}

	logging.Logger().Tracef("Volume descriptor version: %d", brvd.VolumeDescriptorVersion)
	if brvd.VolumeDescriptorVersion != consts.ISO9660_VOLUME_DESC_VERSION {
		logging.Logger().Warnf("Invalid volume descriptor version: %d, expected: %d", brvd.VolumeDescriptorVersion, consts.ISO9660_VOLUME_DESC_VERSION)
	}

	logging.Logger().Tracef("Boot system identifier: %s", brvd.BootSystemIdentifier)
	logging.Logger().Tracef("Boot identifier: %s", brvd.BootIdentifier)

	return brvd, nil
}

type BootRecordVolumeDescriptor struct {
	Type                    VolumeDescriptorType // Numeric value
	StandardIdentifier      string               // Always "CD001"
	VolumeDescriptorVersion int                  // Numeric value
	BootSystemIdentifier    string               // a-characters string
	BootIdentifier          string               // Always "CD001"
	BootSystemUse           [1976]byte           // Boot System Use
}

// Unmarshal parses the given byte slice and populates the PrimaryVolumeDescriptor struct.
func (brvd *BootRecordVolumeDescriptor) Unmarshal(data [consts.ISO9660_SECTOR_SIZE]byte) (err error) {

	logging.Logger().Tracef("Unmarshalling %d bytes of boot record volume descriptor data", len(data))

	if len(data) < consts.ISO9660_SECTOR_SIZE {
		return errors.New("invalid data length")
	}

	brvd.Type = VolumeDescriptorType(data[0])
	brvd.StandardIdentifier = string(data[1:6])
	brvd.VolumeDescriptorVersion = int(data[6])
	brvd.BootSystemIdentifier = strings.TrimSpace(string(data[7:39]))
	brvd.BootIdentifier = string(data[39:71])
	copy(brvd.BootSystemUse[:], data[71:2048])

	return nil
}
