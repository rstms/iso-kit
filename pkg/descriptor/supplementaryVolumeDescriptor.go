package descriptor

import (
	"encoding/binary"
	"github.com/bgrewell/iso-kit/pkg/consts"
	"github.com/bgrewell/iso-kit/pkg/directory"
	"github.com/bgrewell/iso-kit/pkg/encoding"
	"github.com/bgrewell/iso-kit/pkg/logging"
	"github.com/bgrewell/iso-kit/pkg/path"
	"github.com/go-logr/logr"
	"io"
	"strings"
)

// ParseSupplementaryVolumeDescriptor parses the given volume descriptor and returns a SupplementaryVolumeDescriptor.
func ParseSupplementaryVolumeDescriptor(vd VolumeDescriptor, isoFile io.ReaderAt, useRR bool, logger logr.Logger) (*SupplementaryVolumeDescriptor, error) {
	logger.V(logging.TRACE).Info("Parsing supplementary volume descriptor")

	svd := &SupplementaryVolumeDescriptor{
		isoFile: isoFile,
		logger:  logger,
	}

	if err := svd.Unmarshal(vd.Data(), isoFile); err != nil {
		logger.Error(err, "Failed to unmarshal supplementary volume descriptor")
		return nil, err
	}
	logger.V(logging.TRACE).Info("Successfully parsed supplementary volume descriptor")

	logger.V(logging.TRACE).Info("Volume descriptor type", "type", svd.Type())
	if svd.Type() != VolumeDescriptorSupplementary {
		logger.Error(nil, "WARNING: Invalid supplementary volume descriptor",
			"actualType", svd.Type(), "expectedType", VolumeDescriptorSupplementary)
	}

	logger.V(logging.TRACE).Info("Standard identifier", "identifier", svd.Identifier())
	if svd.Identifier() != consts.ISO9660_STD_IDENTIFIER {
		logger.Error(nil, "WARNING: Invalid standard identifier",
			"actualIdentifier", svd.Identifier(), "expectedIdentifier", consts.ISO9660_STD_IDENTIFIER)
	}

	// Note: The version number can be 1 or 2, depending on standard vs. enhanced.
	v := svd.Version()
	logger.V(logging.TRACE).Info("Volume descriptor version", "version", v)

	switch v {
	case 1:
		logger.V(logging.TRACE).Info("Volume descriptor version indicates a standard ISO9660 descriptor")
	case 2:
		logger.V(logging.TRACE).Info("Volume descriptor version indicates an enhanced (Joliet) descriptor")
	default:
		logger.Error(nil, "WARNING: Invalid volume descriptor version",
			"actualVersion", v, "expectedVersions", "1 (standard) or 2 (enhanced)")
	}

	// Log remaining fields
	logSupplementaryVolumeFields(logger, svd)

	return svd, nil
}

func logSupplementaryVolumeFields(logger logr.Logger, svd *SupplementaryVolumeDescriptor) {
	logger.V(logging.TRACE).Info("Volume flags", "volumeFlags", svd.VolumeFlags)
	logger.V(logging.TRACE).Info("System identifier", "systemIdentifier", svd.SystemIdentifier)
	logger.V(logging.TRACE).Info("Volume identifier", "volumeIdentifier", svd.VolumeIdentifier)
	logger.V(logging.TRACE).Info("Volume space size", "volumeSpaceSize", svd.VolumeSpaceSize)
	logger.V(logging.TRACE).Info("Volume set size", "volumeSetSize", svd.VolumeSetSize)
	logger.V(logging.TRACE).Info("Volume sequence number", "volumeSequenceNumber", svd.VolumeSequenceNumber)
	logger.V(logging.TRACE).Info("Logical block size", "logicalBlockSize", svd.LogicalBlockSize)
	logger.V(logging.TRACE).Info("Path table size", "pathTableSize", svd.PathTableSize())
	logger.V(logging.TRACE).Info("Path table location (L)", "lPathTableLocation", svd.LPathTableLocation)
	logger.V(logging.TRACE).Info("Path table location (M)", "mPathTableLocation", svd.MPathTableLocation)
	logger.V(logging.TRACE).Info("Root directory entry", "rootDirectoryEntry", svd.RootDirectoryEntry)
	logger.V(logging.TRACE).Info("Volume set identifier", "volumeSetIdentifier", svd.VolumeSetIdentifier)
	logger.V(logging.TRACE).Info("Publisher identifier", "publisherIdentifier", svd.PublisherIdentifier)
	logger.V(logging.TRACE).Info("Data preparer identifier", "dataPreparerIdentifier", svd.DataPreparerIdentifier)
	logger.V(logging.TRACE).Info("Application identifier", "applicationIdentifier", svd.ApplicationIdentifier)
	logger.V(logging.TRACE).Info("Copyright file identifier", "copyrightFileIdentifier", svd.CopyRightFileIdentifier)
	logger.V(logging.TRACE).Info("Abstract file identifier", "abstractFileIdentifier", svd.AbstractFileIdentifier)
	logger.V(logging.TRACE).Info("Bibliographic file identifier", "bibliographicFileIdentifier", svd.BibliographicFileIdentifier)
	logger.V(logging.TRACE).Info("Volume creation date", "volumeCreationDate", svd.VolumeCreationDate)
	logger.V(logging.TRACE).Info("Volume modification date", "volumeModificationDate", svd.VolumeModificationDate)
	logger.V(logging.TRACE).Info("Volume expiration date", "volumeExpirationDate", svd.VolumeExpirationDate)
	logger.V(logging.TRACE).Info("Volume effective date", "volumeEffectiveDate", svd.VolumeEffectiveDate)
	logger.V(logging.TRACE).Info("File structure version", "fileStructureVersion", svd.FileStructureVersion)
	logger.V(logging.TRACE).Info("Application use", "applicationUse", strings.TrimSpace(string(svd.ApplicationUse[:])))

	logger.V(logging.TRACE).Info("Optional path table location (L)", "lOptionalPathTableLocation", svd.LOptionalPathTableLocation)
	logger.V(logging.TRACE).Info("Optional path table location (M)", "mOptionalPathTableLocation", svd.MOptionalPathTableLocation)

	// Escape sequences (unique to SVD)
	logger.V(logging.TRACE).Info("Escape sequences", "escapeSequences", svd.EscapeSequences)
	switch string(svd.EscapeSequences[0:3]) {
	case consts.JOLIET__LEVEL_1_ESCAPE:
		logger.V(logging.TRACE).Info("Level 1 Joliet escape sequence detected")
	case consts.JOLIET__LEVEL_2_ESCAPE:
		logger.V(logging.TRACE).Info("Level 2 Joliet escape sequence detected")
	case consts.JOLIET__LEVEL_3_ESCAPE:
		logger.V(logging.TRACE).Info("Level 3 Joliet escape sequence detected")
	}

	// Log any unused fields if needed for debugging
	logger.V(logging.TRACE).Info("Unused field 2", "unusedField2", svd.UnusedField2)
	logger.V(logging.TRACE).Info("Unused field 4", "unusedField4", svd.UnusedField4)
	logger.V(logging.TRACE).Info("Unused field 5", "unusedField5", svd.UnusedField5)
}

// SupplementaryVolumeDescriptor represents a supplementary volume descriptor in an ISO file.
type SupplementaryVolumeDescriptor struct {
	rawData                     [2048]byte                // Raw data from the volume descriptor
	vdType                      VolumeDescriptorType      // Numeric value
	standardIdentifier          string                    // Always "CD001"
	volumeDescriptorVersion     int8                      // Numeric value
	VolumeFlags                 [1]byte                   // 8 bits of flags
	SystemIdentifier            string                    // Identifier of the system that can act upon the volume
	VolumeIdentifier            string                    // Identifier of the volume
	UnusedField2                [8]byte                   // Unused field should be 0x00
	VolumeSpaceSize             int32                     // Size of the volume in logical blocks
	EscapeSequences             [32]byte                  // Should be 0x00
	VolumeSetSize               int16                     // Number of volumes in the volume set
	VolumeSequenceNumber        int16                     // Number of this volume in the volume set
	LogicalBlockSize            int16                     // Size of the logical blocks in bytes
	pathTableSize               int32                     // Size of the path table in bytes
	LPathTableLocation          uint32                    // Location of the path table for the first directory record
	LOptionalPathTableLocation  uint32                    // Location of the optional path table
	MPathTableLocation          uint32                    // Location of the path table for the second directory record
	MOptionalPathTableLocation  uint32                    // Location of the optional path table
	RootDirectoryEntry          *directory.DirectoryEntry // Directory entry for the root directory
	VolumeSetIdentifier         string                    // Identifier of the volume set
	PublisherIdentifier         string                    // Identifier of the publisher
	DataPreparerIdentifier      string                    // Identifier of the data preparer
	ApplicationIdentifier       string                    // Identifier of the application
	CopyRightFileIdentifier     string                    // Identifier of the copyright file
	AbstractFileIdentifier      string                    // Identifier of the abstract file
	BibliographicFileIdentifier string                    // Identifier of the bibliographic file
	VolumeCreationDate          string                    // Date and time the volume was created
	VolumeModificationDate      string                    // Date and time the volume was last modified
	VolumeExpirationDate        string                    // Date and time the volume expires
	VolumeEffectiveDate         string                    // Date and time the volume is effective
	FileStructureVersion        byte                      // Version of the file structure
	UnusedField4                byte                      // Unused field should be 0x00
	ApplicationUse              [512]byte                 // Application-specific data
	UnusedField5                [653]byte                 // Unused field should be 0x00
	pathTable                   []*path.PathTableRecord   // Path Table
	isoFile                     io.ReaderAt               // Reader for the ISO file
	isJoliet                    bool                      // Whether this is a Joliet SVD
	logger                      logr.Logger               // Logger
	useRR                       bool                      // Use RockRdige Extensions
}

// Type returns the volume descriptor type for the SVD.
func (svd *SupplementaryVolumeDescriptor) Type() VolumeDescriptorType {
	return svd.vdType
}

// Identifier returns the standard identifier for the SVD.
func (svd *SupplementaryVolumeDescriptor) Identifier() string {
	return svd.standardIdentifier
}

// Version returns the volume descriptor version for the SVD.
func (svd *SupplementaryVolumeDescriptor) Version() int8 {
	return svd.volumeDescriptorVersion
}

// Data returns the raw data for the SVD.
func (svd *SupplementaryVolumeDescriptor) Data() [2048]byte {
	return svd.rawData
}

// SystemID returns the path table location for the SVD.
func (svd *SupplementaryVolumeDescriptor) PathTableLocation() uint32 {
	return svd.LPathTableLocation
}

// PathTableSize returns the size of the path table for the SVD.
func (svd *SupplementaryVolumeDescriptor) PathTableSize() int32 {
	return svd.pathTableSize
}

// PathTable returns the path table records for the SVD.
func (svd *SupplementaryVolumeDescriptor) PathTable() *[]*path.PathTableRecord {
	if svd.pathTable == nil {
		svd.pathTable = make([]*path.PathTableRecord, 0)
	}

	return &svd.pathTable
}

// IsJoliet returns true if the SVD is a Joliet SVD.
func (svd *SupplementaryVolumeDescriptor) IsJoliet() bool {
	return svd.isJoliet
}

// Unmarshal parses the given byte slice and populates the PrimaryVolumeDescriptor struct.
func (svd *SupplementaryVolumeDescriptor) Unmarshal(data [consts.ISO9660_SECTOR_SIZE]byte, isoFile io.ReaderAt) (err error) {

	svd.logger.V(logging.TRACE).Info("Unmarshalling supplementary volume descriptor", "len", len(data))

	svd.rawData = data

	// Handle escape sequences early to determine if Joliet is in use
	copy(svd.EscapeSequences[:], data[88:120])
	if string(svd.EscapeSequences[0:3]) == consts.JOLIET__LEVEL_1_ESCAPE ||
		string(svd.EscapeSequences[0:3]) == consts.JOLIET__LEVEL_2_ESCAPE ||
		string(svd.EscapeSequences[0:3]) == consts.JOLIET__LEVEL_3_ESCAPE {
		svd.isJoliet = true
	}

	rootRecord := directory.NewRecord(svd.logger)
	rootRecord.Joliet = svd.isJoliet
	err = rootRecord.Unmarshal(data[156:190], isoFile)
	if err != nil {
		return err
	}

	svd.vdType = VolumeDescriptorType(data[0])
	svd.standardIdentifier = string(data[1:6])
	svd.volumeDescriptorVersion = int8(data[6])
	copy(svd.VolumeFlags[:], data[7:8])
	svd.SystemIdentifier = string(data[8:40])
	svd.VolumeIdentifier = string(data[40:72])
	copy(svd.UnusedField2[:], data[72:80])
	svd.VolumeSpaceSize, err = encoding.UnmarshalInt32LSBMSB(data[80:88])
	if err != nil {
		return err
	}
	svd.VolumeSetSize, err = encoding.UnmarshalInt16LSBMSB(data[120:124])
	if err != nil {
		return err
	}
	svd.VolumeSequenceNumber, err = encoding.UnmarshalInt16LSBMSB(data[124:128])
	if err != nil {
		return err
	}
	svd.LogicalBlockSize, err = encoding.UnmarshalInt16LSBMSB(data[128:132])
	if err != nil {
		return err
	}
	svd.pathTableSize, err = encoding.UnmarshalInt32LSBMSB(data[132:140])
	if err != nil {
		return err
	}
	svd.LPathTableLocation = binary.LittleEndian.Uint32(data[140:144])
	svd.LOptionalPathTableLocation = binary.LittleEndian.Uint32(data[144:148])
	svd.MPathTableLocation = binary.BigEndian.Uint32(data[148:152])
	svd.MOptionalPathTableLocation = binary.BigEndian.Uint32(data[152:156])
	svd.RootDirectoryEntry = directory.NewEntry(rootRecord, isoFile, svd.useRR, svd.logger)
	svd.VolumeSetIdentifier = string(data[190:318])
	svd.PublisherIdentifier = string(data[318:446])
	svd.DataPreparerIdentifier = string(data[446:574])
	svd.ApplicationIdentifier = string(data[574:702])
	svd.CopyRightFileIdentifier = string(data[702:739])
	svd.AbstractFileIdentifier = string(data[739:776])
	svd.BibliographicFileIdentifier = string(data[776:813])
	svd.VolumeCreationDate = string(data[813:830])
	svd.VolumeModificationDate = string(data[830:847])
	svd.VolumeExpirationDate = string(data[847:864])
	svd.VolumeEffectiveDate = string(data[864:881])
	svd.FileStructureVersion = data[881]
	svd.UnusedField4 = data[882]
	copy(svd.ApplicationUse[:], data[883:1395])
	copy(svd.UnusedField5[:], data[1395:2048])

	return nil
}
