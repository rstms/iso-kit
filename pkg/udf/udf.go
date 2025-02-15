package udf

import (
	"github.com/bgrewell/iso-kit/pkg/file"
	"github.com/bgrewell/iso-kit/pkg/option"
	"io"
)

func Open(isoReader io.ReaderAt, opts ...option.OpenOption) (*UDF, error) {
	//TODO implement me
	panic("implement me")
}

type UDF struct {
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

func (U UDF) ListFiles() ([]file.FileEntry, error) {
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
