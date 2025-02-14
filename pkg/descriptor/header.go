package descriptor

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
