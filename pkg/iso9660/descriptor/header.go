package descriptor

import (
	"fmt"
	"github.com/bgrewell/iso-kit/pkg/helpers"
	"github.com/bgrewell/iso-kit/pkg/iso9660/consts"
)

type VolumeDescriptorHeader struct {
	// Volume Descriptor Types.
	//  | 0 = Boot Record
	//  | 1 = Primary
	//  | 2 = Supplementary
	//  | 3 = Partition
	//  | 4 - 254 = Reserved
	//  | 255 = Terminator
	VolumeDescriptorType VolumeDescriptorType `json:"volume_descriptor_type"`
	// Standard Identifier should always be 'CD001' as a string or 0x4344303031.
	StandardIdentifier string `json:"standard_identifier"`
	// Volume Descriptor Version. The contents and interpretation depend on the Volume Descriptor Type field.
	VolumeDescriptorVersion uint8 `json:"volume_descriptor_version"`
}

func (h *VolumeDescriptorHeader) Type() VolumeDescriptorType {
	return h.VolumeDescriptorType
}

func (h *VolumeDescriptorHeader) Identifier() string {
	return h.StandardIdentifier
}

func (h *VolumeDescriptorHeader) Version() uint8 {
	return h.VolumeDescriptorVersion
}

// Marshal converts the VolumeDescriptorHeader into its 7-byte on-disk representation.
func (vdh *VolumeDescriptorHeader) Marshal() ([consts.ISO9660_VOLUME_DESC_HEADER_SIZE]byte, error) {
	var buf [consts.ISO9660_VOLUME_DESC_HEADER_SIZE]byte

	// Byte 0: Volume Descriptor Type.
	buf[0] = byte(vdh.VolumeDescriptorType)

	// Bytes 1-5: Standard Identifier.
	// Ensure the string is exactly 5 bytes (truncating or padding with spaces as needed).
	sid := helpers.PadString(vdh.StandardIdentifier, 5)
	copy(buf[1:6], sid)

	// Byte 6: Volume Descriptor Version.
	buf[6] = vdh.VolumeDescriptorVersion

	return buf, nil
}

// Unmarshal parses a 7-byte slice into the VolumeDescriptorHeader.
// It expects data to be exactly 7 bytes long.
func (vdh *VolumeDescriptorHeader) Unmarshal(data [consts.ISO9660_VOLUME_DESC_HEADER_SIZE]byte) error {
	vdh.VolumeDescriptorType = VolumeDescriptorType(data[0])
	vdh.StandardIdentifier = string(data[1:6])
	vdh.VolumeDescriptorVersion = data[6]

	// Optionally, you could verify that StandardIdentifier equals "CD001":
	if vdh.StandardIdentifier != "CD001" {
		return fmt.Errorf("unexpected standard identifier: %q", vdh.StandardIdentifier)
	}

	return nil
}
