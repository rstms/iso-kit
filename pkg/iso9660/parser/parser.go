package parser

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/bgrewell/iso-kit/pkg/filesystem"
	"github.com/bgrewell/iso-kit/pkg/iso9660/consts"
	"github.com/bgrewell/iso-kit/pkg/iso9660/descriptor"
	"github.com/bgrewell/iso-kit/pkg/iso9660/directory"
	"github.com/bgrewell/iso-kit/pkg/iso9660/extensions"
	"io"
)

func NewParser(r io.ReaderAt) *Parser {
	return &Parser{r: r}
}

type Parser struct {
	r io.ReaderAt
}

// GetBootRecord reads and validates the ISO9660 boot record.
func (p *Parser) GetBootRecord() (*descriptor.BootRecordDescriptor, error) {
	const sectorSize = consts.ISO9660_SECTOR_SIZE
	// The Volume Descriptor Set starts at logical sector 16.
	sector := int64(consts.ISO9660_SYSTEM_AREA_SECTORS)
	var buf [2048]byte

	for {
		offset := sector * int64(sectorSize)
		n, err := p.r.ReadAt(buf[:], offset)
		if err != nil {
			return nil, err
		}
		if n != len(buf) {
			return nil, errors.New("failed to read full sector")
		}

		// Unmarshal the Volume Descriptor Header (first 7 bytes)
		header := descriptor.VolumeDescriptorHeader{}
		if err = header.Unmarshal([7]byte(buf[:7])); err != nil {
			return nil, err
		}

		// A Volume Descriptor Set Terminator has type 255.
		if header.VolumeDescriptorType == descriptor.TYPE_TERMINATOR_DESCRIPTOR {
			return nil, errors.New("no boot record found in the volume descriptor set")
		}

		// Validate the ISO9660 signature.
		if string(buf[1:6]) != "CD001" {
			return nil, errors.New("invalid ISO9660 signature")
		}

		// If this is a Boot Record (type 0), unmarshal and return it.
		if header.VolumeDescriptorType == descriptor.TYPE_BOOT_RECORD {
			bootRecord := &descriptor.BootRecordDescriptor{
				VolumeDescriptorHeader: header,
			}
			if err = bootRecord.Unmarshal(buf); err != nil {
				return nil, err
			}
			return bootRecord, nil
		}

		// Otherwise, move to the next sector.
		sector++
	}
}

// GetPrimaryVolumeDescriptor reads and validates the ISO9660 PVD.
func (p *Parser) GetPrimaryVolumeDescriptor() (*descriptor.PrimaryVolumeDescriptor, error) {
	var buf [2048]byte
	_, err := p.r.ReadAt(buf[:], consts.ISO9660_SYSTEM_AREA_SECTORS*consts.ISO9660_SECTOR_SIZE)
	if err != nil {
		return nil, err
	}

	// Unmarshal the VolumeDescriptorHeader
	header := descriptor.VolumeDescriptorHeader{}
	if err = header.Unmarshal([7]byte(buf[:7])); err != nil {
		return nil, err
	}

	// Validate ISO9660 signature
	if string(buf[1:6]) != "CD001" {
		return nil, errors.New("invalid ISO9660 signature")
	}

	// Create a new PrimaryVolumeDescriptor
	pvd := &descriptor.PrimaryVolumeDescriptor{
		VolumeDescriptorHeader: header,
	}

	// Unmarshal the rest of the buffer
	if err = pvd.Unmarshal([2048]byte(buf[:])); err != nil {
		return nil, err
	}

	return pvd, nil
}

// GetSupplementaryVolumeDescriptors reads and validates the ISO9660 SVD.
func (p *Parser) GetSupplementaryVolumeDescriptors() ([]*descriptor.SupplementaryVolumeDescriptor, error) {
	const sectorSize = consts.ISO9660_SECTOR_SIZE
	// The Volume Descriptor Set starts at logical sector 16.
	sector := int64(consts.ISO9660_SYSTEM_AREA_SECTORS)
	var buf [2048]byte

	// Create a slice to hold the SupplementaryVolumeDescriptors
	var svds []*descriptor.SupplementaryVolumeDescriptor

	for {
		offset := sector * int64(sectorSize)
		n, err := p.r.ReadAt(buf[:], offset)
		if err != nil {
			return nil, err
		}
		if n != len(buf) {
			return nil, errors.New("failed to read full sector")
		}

		// Unmarshal the Volume Descriptor Header (first 7 bytes)
		header := descriptor.VolumeDescriptorHeader{}
		if err = header.Unmarshal([7]byte(buf[:7])); err != nil {
			return nil, err
		}

		// A Volume Descriptor Set Terminator has type 255.
		if header.VolumeDescriptorType == descriptor.TYPE_TERMINATOR_DESCRIPTOR {
			if len(svds) == 0 {
				return nil, errors.New("no supplementary volume descriptors found in the volume descriptor set")
			}
			return svds, nil
		}

		// Validate the ISO9660 signature.
		if string(buf[1:6]) != "CD001" {
			return nil, errors.New("invalid ISO9660 signature")
		}

		// If this is a Supplementary Volume Descriptor, unmarshal it and add to the collection.
		if header.VolumeDescriptorType == descriptor.TYPE_SUPPLEMENTARY_DESCRIPTOR {
			svd := &descriptor.SupplementaryVolumeDescriptor{
				VolumeDescriptorHeader: header,
			}

			if err = svd.Unmarshal(buf); err != nil {
				return nil, err
			}

			svds = append(svds, svd)
		}

		// Otherwise, move to the next sector.
		sector++
	}
}

// BuildFileSystemEntries walks the directory tree and converts entries into FileSystemEntry objects.
func (p *Parser) BuildFileSystemEntries(rootDir *directory.DirectoryRecord, RockRidgeEnabled bool) ([]*filesystem.FileSystemEntry, error) {
	if rootDir == nil {
		return nil, errors.New("rootDir cannot be nil")
	}

	visited := make(map[uint32]bool) // Prevent infinite recursion
	var entries []*filesystem.FileSystemEntry

	var walk func(dir *directory.DirectoryRecord, parentPath string) error
	walk = func(dir *directory.DirectoryRecord, parentPath string) error {
		if visited[dir.LocationOfExtent] {
			return nil
		}
		visited[dir.LocationOfExtent] = true

		// Read directory records
		dirRecords, err := p.ReadDirectoryRecords(dir.LocationOfExtent)
		if err != nil {
			return err
		}

		for _, record := range dirRecords {
			// Build full path
			fullPath := parentPath + "/" + record.GetBestName(RockRidgeEnabled)

			// Retrieve file attributes
			permissions := record.GetPermissions(RockRidgeEnabled)
			uid, gid := record.GetOwnership(RockRidgeEnabled)
			creationTime, modificationTime := record.GetTimestamps(RockRidgeEnabled)

			// Create FileSystemEntry
			entry := filesystem.FileSystemEntry{
				Name:       record.GetBestName(RockRidgeEnabled),
				FullPath:   fullPath,
				IsDir:      record.IsDirectory(),
				Size:       record.DataLength,
				Location:   record.LocationOfExtent,
				Mode:       permissions,
				CreateTime: creationTime,
				ModTime:    modificationTime,
				UID:        uid,
				GID:        gid,
			}

			entries = append(entries, &entry)

			// Recursively walk directories
			if record.IsDirectory() && !record.IsSpecial() {
				if err := walk(record, fullPath); err != nil {
					return err
				}
			}
		}
		return nil
	}

	// Start walking from the root directory
	if err := walk(rootDir, ""); err != nil {
		return nil, err
	}

	return entries, nil
}

// TODO: Should this not be exported?
// WalkDirectoryRecords recursively walks the directory tree from a given directory record
// and returns a slice of fully populated DirectoryRecord pointers.
func (p *Parser) WalkDirectoryRecords(rootDir *directory.DirectoryRecord) ([]*directory.DirectoryRecord, error) {
	if rootDir == nil {
		return nil, errors.New("rootDir cannot be nil")
	}

	visited := make(map[uint32]bool) // Prevent infinite recursion
	var records []*directory.DirectoryRecord

	var walk func(dir *directory.DirectoryRecord) error
	walk = func(dir *directory.DirectoryRecord) error {
		// Prevent revisiting the same directory
		if visited[dir.LocationOfExtent] {
			return nil
		}
		visited[dir.LocationOfExtent] = true

		// Read directory records from this LBA
		dirRecords, err := p.ReadDirectoryRecords(dir.LocationOfExtent)
		if err != nil {
			return err
		}

		for _, record := range dirRecords {
			records = append(records, record)

			// If the record is a directory (excluding `.` and `..` entries), recurse
			if record.IsDirectory() && !record.IsSpecial() {
				if err := walk(record); err != nil {
					return err
				}
			}
		}
		return nil
	}

	// Start walking from the provided root directory record
	if err := walk(rootDir); err != nil {
		return nil, err
	}

	return records, nil
}

// ReadDirectoryRecords reads directory records from a given LBA (logical block address)
// and processes Rock Ridge extensions if present.
func (p *Parser) ReadDirectoryRecords(lba uint32) ([]*directory.DirectoryRecord, error) {
	// Read a full 2048-byte sector from the given LBA
	offset := int64(lba) * consts.ISO9660_SECTOR_SIZE
	buf := make([]byte, consts.ISO9660_SECTOR_SIZE)

	_, err := p.r.ReadAt(buf, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory sector at LBA %d: %w", lba, err)
	}

	var records []*directory.DirectoryRecord
	reader := bytes.NewReader(buf)

	for reader.Len() > 0 {
		// Read length of this directory record (first byte)
		var length byte
		if err := binary.Read(reader, binary.LittleEndian, &length); err != nil {
			return nil, fmt.Errorf("failed to read directory record length: %w", err)
		}

		// If length is zero, we've reached padding or the end of records.
		if length == 0 {
			break
		}

		// Read the record data into a buffer
		recordBuf := make([]byte, length)
		recordBuf[0] = length // First byte already read
		if _, err := io.ReadFull(reader, recordBuf[1:]); err != nil {
			return nil, fmt.Errorf("failed to read directory record: %w", err)
		}

		// Parse directory record from raw data
		dr := &directory.DirectoryRecord{}
		err = dr.Unmarshal(recordBuf)
		if err != nil {
			return nil, fmt.Errorf("failed to parse directory record: %w", err)
		}

		// **Parse Rock Ridge extensions if present**
		if len(dr.SystemUse) > 0 {
			rr, err := extensions.UnmarshalRockRidge(dr.SystemUse)
			if err == nil {
				dr.RockRidge = rr
			}
		}

		records = append(records, dr)
	}

	return records, nil
}
