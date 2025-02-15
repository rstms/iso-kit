package descriptor

type VolumeDescriptorSet struct {
	Primary       *PrimaryVolumeDescriptor
	Supplementary []*SupplementaryVolumeDescriptor
	Boot          *BootRecordDescriptor
	Terminator    *VolumeDescriptorSetTerminator
}
