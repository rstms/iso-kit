package iso9660

import (
	"github.com/bgrewell/iso-kit/pkg/filesystem"
	"github.com/bgrewell/iso-kit/pkg/iso9660/consts"
	"github.com/bgrewell/iso-kit/pkg/iso9660/descriptor"
	"github.com/bgrewell/iso-kit/pkg/iso9660/directory"
	"github.com/bgrewell/iso-kit/pkg/iso9660/parser"
	"github.com/bgrewell/iso-kit/pkg/iso9660/systemarea"
	"github.com/bgrewell/iso-kit/pkg/option"
	"github.com/go-logr/logr"
	"io"
	"time"
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
		RockRidgeEnabled:           true,
		PreferJoliet:               false,
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
	bootRecord, err := p.GetBootRecord()
	if err != nil {
		return nil, err
	}

	// Read the primary volume descriptor
	pvd, err := p.GetPrimaryVolumeDescriptor()
	if err != nil {
		return nil, err
	}

	// Read the supplementary volume descriptors
	svds, err := p.GetSupplementaryVolumeDescriptors()
	if err != nil {
		return nil, err
	}

	// Handle processing volume descriptor
	var filesystemEntries []*filesystem.FileSystemEntry
	var directoryRecords []*directory.DirectoryRecord
	var activeVD descriptor.VolumeDescriptor
	if openOptions.PreferJoliet && len(svds) > 0 {
		// Open the Joliet filesystem
		filesystemEntries, err = p.BuildFileSystemEntries(svds[0].RootDirectoryRecord, false)
		directoryRecords, err = p.WalkDirectoryRecords(svds[0].RootDirectoryRecord)
		activeVD = svds[0]
	} else {
		filesystemEntries, err = p.BuildFileSystemEntries(pvd.RootDirectoryRecord, openOptions.RockRidgeEnabled)
		directoryRecords, err = p.WalkDirectoryRecords(pvd.RootDirectoryRecord)
		activeVD = pvd
	}

	iso := &ISO9660{
		isoReader:         isoReader,
		openOptions:       openOptions,
		systemArea:        sa,
		bootRecord:        bootRecord,
		pvd:               pvd,
		svds:              svds,
		directoryRecords:  directoryRecords,
		filesystemEntries: filesystemEntries,
		activeVD:          activeVD,
	}

	return iso, nil
}

func Create(filename string, rootPath string, opts ...option.CreateOption) (*ISO9660, error) {
	//TODO implement me
	panic("implement me")
}

type ISO9660 struct {
	isoReader         io.ReaderAt
	openOptions       *option.OpenOptions
	createOptions     *option.CreateOptions
	systemArea        systemarea.SystemArea
	bootRecord        *descriptor.BootRecordDescriptor
	pvd               *descriptor.PrimaryVolumeDescriptor
	svds              []*descriptor.SupplementaryVolumeDescriptor
	activeVD          descriptor.VolumeDescriptor
	directoryRecords  []*directory.DirectoryRecord
	filesystemEntries []*filesystem.FileSystemEntry
}

func (iso *ISO9660) GetVolumeID() string {
	if iso.activeVD == nil {
		return ""
	}
	return iso.activeVD.VolumeIdentifier()
}

func (iso *ISO9660) GetSystemID() string {
	if iso.activeVD == nil {
		return ""
	}
	return iso.activeVD.SystemIdentifier()
}

func (iso *ISO9660) GetVolumeSize() uint32 {
	return 0
}

func (iso *ISO9660) GetVolumeSetID() string {
	if iso.activeVD == nil {
		return ""
	}
	return iso.activeVD.VolumeSetIdentifier()
}

func (iso *ISO9660) GetPublisherID() string {
	if iso.activeVD == nil {
		return ""
	}
	return iso.activeVD.PublisherIdentifier()
}

func (iso *ISO9660) GetDataPreparerID() string {
	if iso.activeVD == nil {
		return ""
	}
	return iso.activeVD.DataPreparerIdentifier()
}

func (iso *ISO9660) GetApplicationID() string {
	if iso.activeVD == nil {
		return ""
	}
	return iso.activeVD.ApplicationIdentifier()
}

func (iso *ISO9660) GetCopyrightID() string {
	if iso.activeVD == nil {
		return ""
	}
	return iso.activeVD.CopyrightFileIdentifier()
}

func (iso *ISO9660) GetAbstractID() string {
	if iso.activeVD == nil {
		return ""
	}
	return iso.activeVD.AbstractFileIdentifier()
}

func (iso *ISO9660) GetBibliographicID() string {
	if iso.activeVD == nil {
		return ""
	}
	return iso.activeVD.BibliographicFileIdentifier()
}

func (iso *ISO9660) GetCreationDateTime() time.Time {
	if iso.activeVD == nil {
		return time.Time{}
	}
	return iso.activeVD.VolumeCreationDateTime()
}

func (iso *ISO9660) GetModificationDateTime() time.Time {
	if iso.activeVD == nil {
		return time.Time{}
	}
	return iso.activeVD.VolumeModificationDateTime()
}

func (iso *ISO9660) GetExpirationDateTime() time.Time {
	if iso.activeVD == nil {
		return time.Time{}
	}
	return iso.activeVD.VolumeExpirationDateTime()
}

func (iso *ISO9660) GetEffectiveDateTime() time.Time {
	if iso.activeVD == nil {
		return time.Time{}
	}
	return iso.activeVD.VolumeEffectiveDateTime()
}

// HasJoliet returns true if the ISO9660 filesystem has Joliet extensions.
func (iso *ISO9660) HasJoliet() bool {
	if iso.activeVD == nil {
		return false
	}
	return iso.activeVD.HasJoliet()
}

// HasRockRidge returns true if the ISO9660 filesystem has Rock Ridge extensions.
func (iso *ISO9660) HasRockRidge() bool {
	for _, rec := range iso.directoryRecords {
		if rec.RockRidge != nil {
			return true
		}
	}
	return false
}

func (iso *ISO9660) RootDirectoryLocation() uint32 {
	return iso.activeVD.RootDirectory().LocationOfExtent
}

func (iso *ISO9660) ListFiles() ([]*filesystem.FileSystemEntry, error) {
	files := make([]*filesystem.FileSystemEntry, 0)
	for _, entry := range iso.filesystemEntries {
		if !entry.IsDir {
			files = append(files, entry)
		}
	}

	return files, nil
}

func (iso *ISO9660) ListDirectories() ([]*filesystem.FileSystemEntry, error) {
	dirs := make([]*filesystem.FileSystemEntry, 0)
	for _, entry := range iso.filesystemEntries {
		if entry.IsDir {
			dirs = append(dirs, entry)
		}
	}

	return dirs, nil
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
