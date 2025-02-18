package udf

import (
	"github.com/bgrewell/iso-kit/pkg/filesystem"
	"github.com/bgrewell/iso-kit/pkg/option"
	"io"
	"time"
)

func Open(isoReader io.ReaderAt, opts ...option.OpenOption) (*UDF, error) {
	//TODO implement me
	panic("implement me")
}

type UDF struct {
}

func (U UDF) RootDirectoryLocation() uint32 {
	//TODO implement me
	panic("implement me")
}

func (U UDF) GetVolumeSetID() string {
	//TODO implement me
	panic("implement me")
}

func (U UDF) GetPublisherID() string {
	//TODO implement me
	panic("implement me")
}

func (U UDF) GetDataPreparerID() string {
	//TODO implement me
	panic("implement me")
}

func (U UDF) GetApplicationID() string {
	//TODO implement me
	panic("implement me")
}

func (U UDF) GetCopyrightID() string {
	//TODO implement me
	panic("implement me")
}

func (U UDF) GetAbstractID() string {
	//TODO implement me
	panic("implement me")
}

func (U UDF) GetBibliographicID() string {
	//TODO implement me
	panic("implement me")
}

func (U UDF) GetCreationDateTime() time.Time {
	//TODO implement me
	panic("implement me")
}

func (U UDF) GetModificationDateTime() time.Time {
	//TODO implement me
	panic("implement me")
}

func (U UDF) GetExpirationDateTime() time.Time {
	//TODO implement me
	panic("implement me")
}

func (U UDF) GetEffectiveDateTime() time.Time {
	//TODO implement me
	panic("implement me")
}

func (U UDF) HasJoliet() bool {
	//TODO implement me
	panic("implement me")
}

func (U UDF) HasRockRidge() bool {
	//TODO implement me
	panic("implement me")
}

func (U UDF) GetVolumeID() string {
	//TODO implement me
	panic("implement me")
}

func (U UDF) GetSystemID() string {
	//TODO implement me
	panic("implement me")
}

func (U UDF) GetVolumeSize() uint32 {
	//TODO implement me
	panic("implement me")
}

func (U UDF) ListFiles() ([]*filesystem.FileSystemEntry, error) {
	//TODO implement me
	panic("implement me")
}

func (U UDF) ListDirectories() ([]*filesystem.FileSystemEntry, error) {
	//TODO implement me
	panic("implement me")
}

func (U UDF) ReadFile(path string) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (U UDF) AddFile(path string, data []byte) error {
	//TODO implement me
	panic("implement me")
}

func (U UDF) RemoveFile(path string) error {
	//TODO implement me
	panic("implement me")
}

func (U UDF) Save(writer io.Writer) error {
	//TODO implement me
	panic("implement me")
}

func (U UDF) Close() error {
	//TODO implement me
	panic("implement me")
}
