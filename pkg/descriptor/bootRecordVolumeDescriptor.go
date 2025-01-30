package descriptor

import (
	"github.com/bgrewell/iso-kit/pkg/consts"
	"github.com/bgrewell/iso-kit/pkg/logging"
	"github.com/go-logr/logr"
	"strings"
)

func ParseBootRecordVolumeDescriptor(vd VolumeDescriptor, logger logr.Logger) (*BootRecordVolumeDescriptor, error) {
	logger.V(logging.TRACE).Info("Parsing boot record volume descriptor")

	brvd := &BootRecordVolumeDescriptor{}
	if err := brvd.Unmarshal(vd.Data()); err != nil {
		logger.Error(err, "Failed to unmarshal boot record volume descriptor")
		return nil, err
	}
	logger.V(logging.TRACE).Info("Successfully parsed boot record volume descriptor")

	// Check type
	logger.V(logging.TRACE).Info("Volume descriptor type", "type", brvd.Type)
	if brvd.Type != VolumeDescriptorBootRecord {
		logger.Error(nil, "WARNING: Invalid boot record volume descriptor", "actualType", brvd.Type,
			"expectedType", VolumeDescriptorBootRecord)
	}

	// Check standard identifier
	logger.V(logging.TRACE).Info("Standard identifier", "identifier", brvd.StandardIdentifier)
	if brvd.StandardIdentifier != consts.ISO9660_STD_IDENTIFIER {
		logger.Error(nil, "WARNING: Invalid standard identifier",
			"actualIdentifier", brvd.StandardIdentifier,
			"expectedIdentifier", consts.ISO9660_STD_IDENTIFIER)
	}

	// Check volume descriptor version
	logger.V(logging.TRACE).Info("Volume descriptor version", "version", brvd.VolumeDescriptorVersion)
	if brvd.VolumeDescriptorVersion != consts.ISO9660_VOLUME_DESC_VERSION {
		logger.Error(nil, "WARNING: Invalid volume descriptor version",
			"actualVersion", brvd.VolumeDescriptorVersion,
			"expectedVersion", consts.ISO9660_VOLUME_DESC_VERSION)
	}

	logger.V(logging.TRACE).Info("Boot system identifier", "bootSystemIdentifier", brvd.BootSystemIdentifier)
	logger.V(logging.TRACE).Info("Boot identifier", "bootIdentifier", brvd.BootIdentifier)

	return brvd, nil
}

type BootRecordVolumeDescriptor struct {
	Type                    VolumeDescriptorType // Numeric value
	StandardIdentifier      string               // Always "CD001"
	VolumeDescriptorVersion int                  // Numeric value
	BootSystemIdentifier    string               // a-characters string
	BootIdentifier          string               // Always "CD001"
	BootSystemUse           [1976]byte           // Boot System Use
	logger                  logr.Logger          // Logger
}

// Unmarshal parses the given byte slice and populates the PrimaryVolumeDescriptor struct.
func (brvd *BootRecordVolumeDescriptor) Unmarshal(data [consts.ISO9660_SECTOR_SIZE]byte) (err error) {

	brvd.logger.V(logging.TRACE).Info("Unmarshalling boot volume descriptor", "len", len(data))

	brvd.Type = VolumeDescriptorType(data[0])
	brvd.StandardIdentifier = string(data[1:6])
	brvd.VolumeDescriptorVersion = int(data[6])
	brvd.BootSystemIdentifier = strings.TrimSpace(string(data[7:39]))
	brvd.BootIdentifier = string(data[39:71])
	copy(brvd.BootSystemUse[:], data[71:2048])

	return nil
}
