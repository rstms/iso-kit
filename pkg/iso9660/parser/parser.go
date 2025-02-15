package parser

import (
	"errors"
	"fmt"
	"github.com/bgrewell/iso-kit/pkg/iso9660/consts"
	"github.com/bgrewell/iso-kit/pkg/iso9660/descriptor"
	"github.com/bgrewell/iso-kit/pkg/iso9660/directory"
	"io"
)

func NewParser(r io.ReaderAt) *Parser {
	return &Parser{r: r}
}

type Parser struct {
	r io.ReaderAt
}

// ReadBootRecord reads and validates the ISO9660 boot record.
func (p *Parser) ReadBootRecord() (*descriptor.BootRecordDescriptor, error) {
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

// ReadPrimaryVolumeDescriptor reads and validates the ISO9660 PVD.
func (p *Parser) ReadPrimaryVolumeDescriptor() (*descriptor.PrimaryVolumeDescriptor, error) {
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
	if err = pvd.Unmarshal(buf[:]); err != nil {
		return nil, err
	}

	return pvd, nil
}

// ReadSupplementaryVolumeDescriptors reads and validates the ISO9660 SVD.
func (p *Parser) ReadSupplementaryVolumeDescriptors() ([]*descriptor.SupplementaryVolumeDescriptor, error) {
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

// ParseDirectoryRecords performs a breadth-first search of the directory hierarchy starting from the root directory entry.
// It uses the io.ReaderAt in the parser to read directory extents and returns a slice containing all encountered DirectoryRecords.
func (p *Parser) ParseDirectoryRecords(root *directory.DirectoryRecord) ([]*directory.DirectoryRecord, error) {
	var result []*directory.DirectoryRecord
	queue := []*directory.DirectoryRecord{root}

	for len(queue) > 0 {
		// Dequeue the first record.
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Only process children if this entry is a directory.
		if !current.IsDirectory() {
			continue
		}

		// Read the entire directory extent.
		extentSize := int(current.DataLength)
		data := make([]byte, extentSize)
		offset := int64(current.LocationOfExtent) * consts.ISO9660_SECTOR_SIZE
		n, err := p.r.ReadAt(data, offset)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("failed to read directory extent at offset %d: %w", offset, err)
		}
		if n < extentSize {
			data = data[:n]
		}

		record := &directory.DirectoryRecord{}
		err = record.Unmarshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal directory record: %w", err)
		}

		// Parse the directory records from the extent.
		children, err := p.ParseDirectoryRecords(record)
		if err != nil {
			return nil, fmt.Errorf("failed to parse directory records: %w", err)
		}

		// Enqueue the children, optionally skipping special entries like '.' and '..'
		for _, child := range children {
			if child.IsSpecial() {
				continue
			}
			queue = append(queue, child)
		}
	}

	return result, nil
}
