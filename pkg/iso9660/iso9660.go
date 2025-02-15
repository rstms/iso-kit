package iso9660

import (
	"github.com/bgrewell/iso-kit/pkg/file"
	"github.com/bgrewell/iso-kit/pkg/iso9660/consts"
	"github.com/bgrewell/iso-kit/pkg/iso9660/descriptor"
	"github.com/bgrewell/iso-kit/pkg/iso9660/directory"
	"github.com/bgrewell/iso-kit/pkg/iso9660/parser"
	"github.com/bgrewell/iso-kit/pkg/iso9660/systemarea"
	"github.com/bgrewell/iso-kit/pkg/option"
	"github.com/go-logr/logr"
	"io"
)

//10.1 Level 1
// At Level 1 the following restrictions shall apply to a volume identified by a Primary Volume Descriptor or by a
// Supplementary Volume Descriptor:
//  - each file shall consist of only one File Section;
//  - a File Name shall not contain more than eight d-characters or eight d1-characters;
//  - a File Name Extension shall not contain more than three d-characters or three d1-characters;
//  - a Directory Identifier shall not contain more than eight d-characters or eight d1-characters.
//
// At Level 1 the following restrictions shall apply to a volume identified by an Enhanced Volume Descriptor:
//  - each file shall consist of only one File Section.
//10.2 Level 2
// At Level 2 the following restriction shall apply:
//  - each file shall consist of only one File Section.
//10.3 Level 3
// At Level 3 no restrictions shall apply

func Open(isoReader io.ReaderAt, opts ...option.OpenOption) (*ISO9660, error) {

	// Set default open options
	emptyCallback := func(currentFilename string, bytesTransferred int64, totalBytes int64, currentFileNumber int, totalFileCount int) {
	}
	openOptions := &option.OpenOptions{
		ReadOnly:                   true,
		ParseOnOpen:                true,
		PreloadDir:                 true,
		StripVersionInfo:           true,
		PreferEnhancedVolumes:      true,
		BootFileExtractLocation:    "[BOOT]",
		ExtractionProgressCallback: emptyCallback,
		Logger:                     logr.Discard(),
	}

	for _, opt := range opts {
		opt(openOptions)
	}

	// Read the System Area
	saBuf := [consts.ISO9660_SECTOR_SIZE * consts.ISO9660_SYSTEM_AREA_SECTORS]byte{}
	if _, err := isoReader.ReadAt(saBuf[:], 0); err != nil {
		return nil, err
	}
	sa := systemarea.SystemArea{
		Contents: saBuf,
	}

	// Create a parser
	p := parser.NewParser(isoReader)

	// Read the boot record
	bootRecord, err := p.ReadBootRecord()
	if err != nil {
		return nil, err
	}

	// Read the primary volume descriptor
	pvd, err := p.ReadPrimaryVolumeDescriptor()
	if err != nil {
		return nil, err
	}

	// Read the supplementary volume descriptors
	svds, err := p.ReadSupplementaryVolumeDescriptors()
	if err != nil {
		return nil, err
	}

	var directoryRecords []*directory.DirectoryRecord
	if openOptions.PreferEnhancedVolumes && len(svds) > 0 {
		directoryRecords, err = p.ParseDirectoryRecords(svds[0].RootDirectoryRecord)
	} else {
		directoryRecords, err = p.ParseDirectoryRecords(pvd.RootDirectoryRecord)
	}

	iso := &ISO9660{
		isoReader:        isoReader,
		openOptions:      openOptions,
		systemArea:       sa,
		bootRecord:       bootRecord,
		pvd:              pvd,
		svds:             svds,
		directoryRecords: directoryRecords,
	}

	return iso, nil
}

func Create(filename string, rootPath string, opts ...option.CreateOption) (*ISO9660, error) {
	//TODO implement me
	panic("implement me")
}

type ISO9660 struct {
	isoReader        io.ReaderAt
	openOptions      *option.OpenOptions
	createOptions    *option.CreateOptions
	systemArea       systemarea.SystemArea
	bootRecord       *descriptor.BootRecordDescriptor
	pvd              *descriptor.PrimaryVolumeDescriptor
	svds             []*descriptor.SupplementaryVolumeDescriptor
	directoryRecords []*directory.DirectoryRecord
}

func (iso *ISO9660) GetVolumeID() string {
	//TODO implement me
	panic("implement me")
}

func (iso *ISO9660) GetSystemID() string {
	//TODO implement me
	panic("implement me")
}

func (iso *ISO9660) GetVolumeSize() uint32 {
	//TODO implement me
	panic("implement me")
}

func (iso *ISO9660) ListFiles() ([]file.FileEntry, error) {
	//TODO implement me
	panic("implement me")
}

func (iso *ISO9660) ReadFile(path string) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (iso *ISO9660) AddFile(path string, data []byte) error {
	//TODO implement me
	panic("implement me")
}

func (iso *ISO9660) RemoveFile(path string) error {
	//TODO implement me
	panic("implement me")
}

func (iso *ISO9660) Save(writer io.Writer) error {
	//TODO implement me
	panic("implement me")
}

func (iso *ISO9660) Close() error {
	//TODO implement me
	panic("implement me")
}
