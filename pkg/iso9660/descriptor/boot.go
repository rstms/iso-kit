package descriptor

import (
	"fmt"
	"github.com/rstms/iso-kit/pkg/consts"
	"github.com/rstms/iso-kit/pkg/helpers"
	"github.com/rstms/iso-kit/pkg/iso9660/directory"
	"github.com/rstms/iso-kit/pkg/iso9660/info"
	"github.com/rstms/iso-kit/pkg/logging"
	"strings"
	"time"
)

const (
	// Boot System Use Size is the size of a sector minus 71 bytes
	BOOT_SYSTEM_USE_SIZE = consts.ISO9660_SECTOR_SIZE - 71
)

type BootRecordDescriptor struct {
	VolumeDescriptorHeader
	BootRecordBody
}

func (d *BootRecordDescriptor) DescriptorType() VolumeDescriptorType {
	return TYPE_BOOT_RECORD
}

func (d *BootRecordDescriptor) VolumeIdentifier() string {
	//TODO implement me
	panic("implement me")
}

func (d *BootRecordDescriptor) SystemIdentifier() string {
	//TODO implement me
	panic("implement me")
}

func (d *BootRecordDescriptor) VolumeSetIdentifier() string {
	//TODO implement me
	panic("implement me")
}

func (d *BootRecordDescriptor) PublisherIdentifier() string {
	//TODO implement me
	panic("implement me")
}

func (d *BootRecordDescriptor) DataPreparerIdentifier() string {
	//TODO implement me
	panic("implement me")
}

func (d *BootRecordDescriptor) ApplicationIdentifier() string {
	//TODO implement me
	panic("implement me")
}

func (d *BootRecordDescriptor) CopyrightFileIdentifier() string {
	//TODO implement me
	panic("implement me")
}

func (d *BootRecordDescriptor) AbstractFileIdentifier() string {
	//TODO implement me
	panic("implement me")
}

func (d *BootRecordDescriptor) BibliographicFileIdentifier() string {
	//TODO implement me
	panic("implement me")
}

func (d *BootRecordDescriptor) VolumeCreationDateTime() time.Time {
	//TODO implement me
	panic("implement me")
}

func (d *BootRecordDescriptor) VolumeModificationDateTime() time.Time {
	//TODO implement me
	panic("implement me")
}

func (d *BootRecordDescriptor) VolumeExpirationDateTime() time.Time {
	//TODO implement me
	panic("implement me")
}

func (d *BootRecordDescriptor) VolumeEffectiveDateTime() time.Time {
	//TODO implement me
	panic("implement me")
}

func (d *BootRecordDescriptor) HasJoliet() bool {
	//TODO implement me
	panic("implement me")
}

func (d *BootRecordDescriptor) HasRockRidge() bool {
	//TODO implement me
	panic("implement me")
}

func (d *BootRecordDescriptor) RootDirectory() *directory.DirectoryRecord {
	//TODO implement me
	panic("implement me")
}

func (d *BootRecordDescriptor) GetObjects() []info.ImageObject {
	return []info.ImageObject{d}
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
	// --- Fields that are not part of the ISO9660 object ---
	// Object Location (in bytes)
	ObjectLocation int64 `json:"object_location"`
	// Object Size (in bytes)
	ObjectSize uint32 `json:"object_size"`
	// Logger
	Logger *logging.Logger
}

func (b BootRecordBody) Type() string {
	return "Volume Descriptor"
}

func (b BootRecordBody) Name() string {
	return "Boot Record Volume Descriptor"
}

func (b BootRecordBody) Description() string {
	return fmt.Sprintf("%s: %s", b.BootSystemIdentifier, b.BootIdentifier)
}

func (b BootRecordBody) Properties() map[string]interface{} {
	return map[string]interface{}{}
}

func (b BootRecordBody) Offset() int64 {
	return b.ObjectLocation
}

func (b BootRecordBody) Size() int {
	return int(b.ObjectSize)
}

// Marshal converts the BootRecordDescriptor into its 2048-byte on-disk representation.
func (d *BootRecordDescriptor) Marshal() ([]byte, error) {
	var buf [consts.ISO9660_SECTOR_SIZE]byte
	offset := 0

	// 1. Marshal the VolumeDescriptorHeader (first 7 bytes).
	headerBytes, err := d.VolumeDescriptorHeader.Marshal()
	if err != nil {
		return buf[:], fmt.Errorf("failed to marshal VolumeDescriptorHeader: %w", err)
	}
	copy(buf[0:7], headerBytes[:])
	offset += 7

	// 2. Boot System Identifier: 32 bytes.
	sysIDBytes := helpers.PadString(d.BootRecordBody.BootSystemIdentifier, 32)
	copy(buf[offset:offset+32], sysIDBytes)
	offset += 32

	// 3. Boot Identifier: 32 bytes.
	bootIDBytes := helpers.PadString(d.BootRecordBody.BootIdentifier, 32)
	copy(buf[offset:offset+32], bootIDBytes)
	offset += 32

	// 4. Boot System Use: remaining bytes.
	copy(buf[offset:offset+BOOT_SYSTEM_USE_SIZE], d.BootRecordBody.BootSystemUse[:])
	offset += BOOT_SYSTEM_USE_SIZE

	if offset != consts.ISO9660_SECTOR_SIZE {
		return buf[:], fmt.Errorf("marshal BootRecordDescriptor: incorrect offset %d", offset)
	}
	return buf[:], nil
}

// Unmarshal parses a 2048-byte sector into the BootRecordDescriptor.
func (d *BootRecordDescriptor) Unmarshal(data [consts.ISO9660_SECTOR_SIZE]byte) error {
	offset := 0

	// 1. Unmarshal the VolumeDescriptorHeader (first 7 bytes).
	var headerBytes [7]byte
	copy(headerBytes[:], data[0:7])
	if err := d.VolumeDescriptorHeader.Unmarshal(headerBytes); err != nil {
		return fmt.Errorf("failed to unmarshal VolumeDescriptorHeader: %w", err)
	}
	offset += 7

	// 2. Boot System Identifier: 32 bytes.
	// Trim trailing spaces.
	d.BootRecordBody.BootSystemIdentifier = strings.TrimRight(string(data[offset:offset+32]), " ")
	offset += 32

	// 3. Boot Identifier: 32 bytes.
	d.BootRecordBody.BootIdentifier = strings.TrimRight(string(data[offset:offset+32]), " ")
	offset += 32

	// 4. Boot System Use: remaining BOOT_SYSTEM_USE_SIZE bytes.
	copy(d.BootRecordBody.BootSystemUse[:], data[offset:offset+BOOT_SYSTEM_USE_SIZE])
	offset += BOOT_SYSTEM_USE_SIZE

	if offset != consts.ISO9660_SECTOR_SIZE {
		return fmt.Errorf("unmarshal BootRecordDescriptor: incorrect offset %d", offset)
	}
	return nil
}
