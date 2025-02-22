package descriptor

type VolumeDescriptorSet struct {
	Boot          *BootRecordDescriptor
	Primary       *PrimaryVolumeDescriptor
	Partition     []*VolumePartitionDescriptor
	Supplementary []*SupplementaryVolumeDescriptor
	Terminator    *VolumeDescriptorSetTerminator
}
