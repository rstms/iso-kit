package descriptor

import "github.com/rstms/iso-kit/pkg/iso9660/directory"

type DirectoryRecordCollection struct {
	DirectoryRecords []*directory.DirectoryRecord
}
