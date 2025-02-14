package descriptor

type VolumeDescriptorSet struct {
	Primary       *PrimaryVolumeDescriptor
	Supplementary []*SupplementaryVolumeDescriptor
	Boot          *BootRecordVolumeDescriptor
	Terminator    *VolumeDescriptorSetTerminator
}
