package boot

import (
	"encoding/binary"
	"fmt"
	"github.com/bgrewell/iso-kit/pkg/consts"
	"github.com/bgrewell/iso-kit/pkg/filesystem"
	"github.com/bgrewell/iso-kit/pkg/iso9660/info"
	"github.com/bgrewell/iso-kit/pkg/logging"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// Logical sector 17 containing El-Torito boot catalog
	EL_TORITO_SECTOR = 0x11
	// Default catalog name for non-Rock Ridge filesystems
	EL_TORITO_DEFAULT_CATALOG = "BOOT.CAT"
	// Default catalog name for Rock Ridge filesystems
	EL_TORITO_DEFAULT_CATALOG_RR = "boot.catalog"
)

// PartitionType represents the type of partition in the boot image.
type PartitionType byte

// List of GUID partition types
const (
	Empty         PartitionType = 0x00
	Fat12         PartitionType = 0x01
	XenixRoot     PartitionType = 0x02
	XenixUsr      PartitionType = 0x03
	Fat16         PartitionType = 0x04
	ExtendedCHS   PartitionType = 0x05
	Fat16b        PartitionType = 0x06
	NTFS          PartitionType = 0x07
	CommodoreFAT  PartitionType = 0x08
	Fat32CHS      PartitionType = 0x0b
	Fat32LBA      PartitionType = 0x0c
	Fat16bLBA     PartitionType = 0x0e
	ExtendedLBA   PartitionType = 0x0f
	Linux         PartitionType = 0x83
	LinuxExtended PartitionType = 0x85
	LinuxLVM      PartitionType = 0x8e
	Iso9660       PartitionType = 0x96
	MacOSXUFS     PartitionType = 0xa8
	MacOSXBoot    PartitionType = 0xab
	HFS           PartitionType = 0xaf
	Solaris8Boot  PartitionType = 0xbe
	//GPTProtective PartitionType = 0xef
	EFISystem  PartitionType = 0xef
	VMWareFS   PartitionType = 0xfb
	VMWareSwap PartitionType = 0xfc
)

func (p PartitionType) String() string {
	switch p {
	case Empty:
		return "Empty"
	case Fat12:
		return "FAT12"
	case XenixRoot:
		return "Xenix Root"
	case XenixUsr:
		return "Xenix User"
	case Fat16:
		return "FAT16"
	case ExtendedCHS:
		return "Extended (CHS)"
	case Fat16b:
		return "FAT16B"
	case NTFS:
		return "NTFS"
	case CommodoreFAT:
		return "Commodore FAT"
	case Fat32CHS:
		return "FAT32 (CHS)"
	case Fat32LBA:
		return "FAT32 (LBA)"
	case Fat16bLBA:
		return "FAT16B (LBA)"
	case ExtendedLBA:
		return "Extended (LBA)"
	case Linux:
		return "Linux"
	case LinuxExtended:
		return "Linux Extended"
	case LinuxLVM:
		return "Linux LVM"
	case Iso9660:
		return "ISO9660"
	case MacOSXUFS:
		return "MacOS X UFS"
	case MacOSXBoot:
		return "MacOS X Boot"
	case HFS:
		return "HFS"
	case Solaris8Boot:
		return "Solaris 8 Boot"
	case EFISystem:
		return "EFI System"
	case VMWareFS:
		return "VMWare FS"
	case VMWareSwap:
		return "VMWare Swap"
	default:
		return "Unknown"
	}
}

// Platform represents the target booting system for an El-Torito bootable ISO.
type Platform uint8

const (
	BIOS Platform = 0x0  // Classic PC-BIOS x86
	PPC  Platform = 0x1  // PowerPC
	Mac  Platform = 0x2  // Macintosh systems
	EFI  Platform = 0xef // Extensible Firmware Interface (EFI)
)

func (p Platform) String() string {
	switch p {
	case BIOS:
		return "BIOS"
	case PPC:
		return "PowerPC"
	case Mac:
		return "Macintosh"
	case EFI:
		return "EFI"
	default:
		return "Unknown"
	}
}

// Emulation represents the emulation mode used for booting.
type Emulation uint8

const (
	NoEmulation        Emulation = 0x0 // No emulation (default)
	Floppy12Emulation  Emulation = 0x1 // Emulate a 1.2 MB floppy
	Floppy144Emulation Emulation = 0x2 // Emulate a 1.44 MB floppy
	Floppy288Emulation Emulation = 0x3 // Emulate a 2.88 MB floppy
	HardDiskEmulation  Emulation = 0x4 // Emulate a hard disk
)

func (e Emulation) String() string {
	switch e {
	case NoEmulation:
		return "NoEmul"
	case Floppy12Emulation:
		return "1.2MFloppy"
	case Floppy144Emulation:
		return "1.44MFloppy"
	case Floppy288Emulation:
		return "2.88MFloppy"
	case HardDiskEmulation:
		return "HardDisk"
	default:
		return "Unknown"
	}
}

// ElTorito represents the El-Torito boot structure for a disk.
type ElTorito struct {
	BootCatalog     string           // Path to the boot catalog file
	HideBootCatalog bool             // Whether to hide the boot catalog in the filesystem
	Entries         []*ElToritoEntry // List of El-Torito boot entries
	Platform        Platform         // Target platform for booting
	// Object Location (in bytes)
	ObjectLocation int64 `json:"object_location"`
	// Object Size (in bytes)
	ObjectSize uint32          `json:"object_size"`
	Logger     *logging.Logger // Logger for debug output
}

func (et *ElTorito) Type() string {
	return "Boot Catalog"
}

func (et *ElTorito) Name() string {
	return "El Torito Boot Catalog"
}

func (et *ElTorito) Description() string {
	return fmt.Sprintf("%s Entries: %d", et.BootCatalog, len(et.Entries))
}

func (et *ElTorito) Properties() map[string]interface{} {

	type EntryDetails struct {
		Emulation     string
		Platform      string
		PartitionType string
		Location      uint32
		Size          uint16
	}

	entryDetails := make(map[string]EntryDetails)
	if len(et.Entries) > 0 {
		for _, entry := range et.Entries {
			entryDetails[entry.BootFile] = EntryDetails{
				Emulation:     entry.Emulation.String(),
				Platform:      entry.Platform.String(),
				PartitionType: entry.PartitionType.String(),
				Location:      entry.location,
				Size:          entry.size,
			}
		}
	}

	return map[string]interface{}{
		"Entries":         len(et.Entries),
		"Platform":        et.Platform,
		"HideBootCatalog": et.HideBootCatalog,
		"EntryDetails":    entryDetails,
	}
}

func (et *ElTorito) Offset() int64 {
	return et.ObjectLocation
}

func (et *ElTorito) Size() int {
	return int(et.ObjectSize)
}

func (et *ElTorito) GetObjects() []info.ImageObject {
	return []info.ImageObject{et}
}

func (et *ElTorito) Marshal() ([]byte, error) {
	if len(et.Entries) == 0 {
		return nil, fmt.Errorf("El Torito Boot Catalog has no entries")
	}

	// Boot Catalog is stored in 2048-byte sectors, ensure correct alignment
	data := make([]byte, consts.ISO9660_SECTOR_SIZE)

	// 1️⃣ Write Validation Entry (First 32 bytes)
	data[0] = 0x01                    // Header ID
	copy(data[1:6], "EL TORITO SPEC") // Identifier
	data[0x1E] = 0x55
	data[0x1F] = 0xAA

	// Compute checksum
	checksum := uint16(0)
	for i := 0; i < 32; i += 2 {
		checksum += binary.LittleEndian.Uint16(data[i : i+2])
	}
	binary.LittleEndian.PutUint16(data[0x1C:0x1E], -checksum) // Store negative checksum

	// 2️⃣ Write Initial Boot Entry (First Boot Entry, starts at offset 32)
	offset := 32
	for _, entry := range et.Entries {
		if offset+32 > len(data) {
			return nil, fmt.Errorf("Boot catalog exceeds sector size limit")
		}

		data[offset] = 0x88                    // Boot Indicator (0x88 = Bootable)
		data[offset+1] = byte(entry.Platform)  // Platform ID
		data[offset+2] = byte(entry.Emulation) // Emulation Type
		binary.LittleEndian.PutUint16(data[offset+4:], entry.LoadSegment)
		binary.LittleEndian.PutUint16(data[offset+6:], entry.size)     // Size in 512-byte blocks
		binary.LittleEndian.PutUint32(data[offset+8:], entry.location) // Location in 2048-byte sectors

		offset += 32 // Move to next entry
	}

	return data, nil
}

// UnmarshalBinary decodes an El-Torito Boot Catalog from binary form
func (et *ElTorito) UnmarshalBinary(data []byte) error {
	if et.Logger != nil {
		et.Logger.Debug("Starting El Torito Boot Catalog unmarshalling")
	}
	if len(data) < 32 {
		err := fmt.Errorf("Boot Catalog: data too short")
		if et.Logger != nil {
			et.Logger.Error(err, "Boot Catalog: data too short")
		}
		return err
	}

	// Parse Validation Entry
	err := parseValidationEntry(data[:32])
	if err != nil {
		if et.Logger != nil {
			et.Logger.Error(err, "Boot Catalog: invalid Validation Entry")
		}
		return fmt.Errorf("Boot Catalog: invalid Validation Entry: %w", err)
	}

	// Parse Boot Entries
	sectionCount := 0
	for offset := 32; offset < len(data); offset += 32 {
		entryData := data[offset : offset+32]

		// Check for End of Catalog
		if entryData[0] == 0x00 {
			if et.Logger != nil {
				et.Logger.Debug("End of El Torito Boot Catalog reached", "offset", offset)
			}
			break
		}

		// Handle Section Headers
		if entryData[0] == 0x90 || entryData[0] == 0x91 {
			sectionCount = int(binary.LittleEndian.Uint16(entryData[2:4]))
			if et.Logger != nil {
				et.Logger.Debug("Section header found", "offset", offset, "entries", sectionCount)
			}
			continue
		}

		// Parse Section Entries
		if sectionCount > 0 {
			entry := parseSectionEntry(entryData)
			if et.Logger != nil {
				et.Logger.Trace("Parsed section entry", "entry", entry)
			}
			et.Entries = append(et.Entries, entry)
			sectionCount--
			continue
		}

		// Parse Initial/Default Entry
		entry := parseInitialEntry(entryData)
		if et.Logger != nil {
			et.Logger.Trace("Parsed initial entry", "entry", entry)
		}
		et.Entries = append(et.Entries, entry)
	}
	if et.Logger != nil {
		et.Logger.Debug("Total El Torito entries discovered", "count", len(et.Entries))
	}
	return nil
}

// ElToritoEntry represents a single entry in an El-Torito boot catalog.
type ElToritoEntry struct {
	Platform      Platform      // Target platform
	Emulation     Emulation     // Emulation mode
	BootFile      string        // Path to the boot file
	HideBootFile  bool          // Whether to hide the boot file in the filesystem
	LoadSegment   uint16        // Open segment address
	PartitionType PartitionType // Partition type of the boot file
	size          uint16        // Size of the boot file in 512-byte blocks
	location      uint32        // Location of the boot file in 2048-byte sectors
}

// SectionHeader represents a header for grouping entries in the boot catalog.
type SectionHeader struct {
	Indicator byte     // Indicator byte (0x90 or 0x91 for the last section)
	Platform  Platform // Target platform
	Entries   uint16   // Number of entries in the section
}

// SelectionCriteria represents optional vendor-specific selection criteria.
type SelectionCriteria struct {
	Type       byte   // Selection criteria type
	VendorData []byte // Vendor-specific data
}

// ValidationEntry represents the validation entry at the start of the boot catalog.
type ValidationEntry struct {
	Platform    Platform // Target platform
	Identifier  string   // Identifier string
	Checksum    uint16   // Validation checksum
	KeyByte55AA uint16   // Fixed 0x55AA marker
}

// BuildBootImageEntries constructs a list of FileSystemEntry objects for all boot images.
func (et *ElTorito) BuildBootImageEntries() ([]*filesystem.FileSystemEntry, error) {
	var entries []*filesystem.FileSystemEntry

	if et.Logger != nil {
		et.Logger.Debug("Building boot image entries for El Torito catalog")
	}

	for i, entry := range et.Entries {
		// Skip non-bootable entries
		if entry.size == 0 || entry.location == 0 {
			if et.Logger != nil {
				et.Logger.Trace("Skipping non-bootable entry", "index", i, "entry", entry)
			}
			continue
		}

		// Construct a synthetic file name for the boot image
		filename := fmt.Sprintf("%d-Boot-%s.img", i+1, entry.Emulation)

		// TODO: The directory should be user configurable, for now we use the same default as 7z
		// Create a FileSystemEntry for the boot image
		fsEntry := &filesystem.FileSystemEntry{
			Name:       filename,
			FullPath:   "/[BOOT]/" + filename, // Logical path inside the ISO
			IsDir:      false,
			Size:       uint32(entry.size * 512), // Convert 512-byte block size
			Location:   entry.location,
			Mode:       0444,        // Read-only boot image
			CreateTime: time.Time{}, // No real timestamp in El Torito
			ModTime:    time.Time{},
			UID:        nil,
			GID:        nil,
		}

		if et.Logger != nil {
			et.Logger.Trace("Boot image entry created", "entry", fsEntry)
		}

		entries = append(entries, fsEntry)
	}

	if et.Logger != nil {
		et.Logger.Debug("Total boot image entries built", "count", len(entries))
	}

	return entries, nil
}

// ExtractBootImages extracts all bootable images to the specified directory.
func (et *ElTorito) ExtractBootImages(ra io.ReaderAt, outputDir string) error {
	if et.Logger != nil {
		et.Logger.Debug("Extracting El Torito boot images to directory", "outputDir", outputDir)
	}

	// Ensure the output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		if et.Logger != nil {
			et.Logger.Error(err, "Failed to create parent directories for boot image", "outputPath", outputPath)
		}
		return fmt.Errorf("failed to create parent directories for %s: %w", outputPath, err)
	}

	for i, entry := range et.Entries {
		// Skip non-bootable entries
		if entry.size == 0 || entry.location == 0 {
			if et.Logger != nil {
				et.Logger.Trace("Skipping non-bootable entry", "index", i, "entry", entry)
			}
			continue
		}

		// Create the file name
		filename := fmt.Sprintf("%d-Boot-%s.img", i+1, entry.Emulation)
		outputPath := filepath.Join(outputDir, filename)

		if et.Logger != nil {
			et.Logger.Debug("Extracting boot image", "outputPath", outputPath)
		}

		// Open the output file for writing
		outFile, err := os.Create(outputPath)
		if err != nil {
			if et.Logger != nil {
				et.Logger.Error(err, "Failed to create file", "outputPath", outputPath)
			}
			return fmt.Errorf("failed to create file %s: %w", outputPath, err)
		}
		defer outFile.Close()

		// Read the boot image data
		startOffset := int64(entry.location) * int64(consts.ISO9660_SECTOR_SIZE)
		data := make([]byte, int64(entry.size)*512) // Size is in 512-byte blocks
		if _, err := ra.ReadAt(data, startOffset); err != nil {
			if et.Logger != nil {
				et.Logger.Error(err, "Failed to read boot image", "offset", startOffset)
			}
			return fmt.Errorf("failed to read boot image at offset %d: %w", startOffset, err)
		}

		// Write the data to the file
		if _, err := outFile.Write(data); err != nil {
			if et.Logger != nil {
				et.Logger.Error(err, "Failed to write boot image", "outputPath", outputPath)
			}
			return fmt.Errorf("failed to write boot image to file %s: %w", outputPath, err)
		}

		// Save the boot file path in the entry
		entry.BootFile = outputPath
		if et.Logger != nil {
			et.Logger.Debug("Boot image successfully extracted", "outputPath", outputPath)
		}
	}
	if et.Logger != nil {
		et.Logger.Debug("All boot images extraction complete.")
	}
	return nil
}

func IsElTorito(bootSystemIdentifier string) bool {
	trimmed := strings.TrimRight(bootSystemIdentifier, "\x00")
	return trimmed == consts.EL_TORITO_BOOT_SYSTEM_ID
}

func parseInitialEntry(data []byte) *ElToritoEntry {
	return &ElToritoEntry{
		Platform:      Platform(data[1]),
		Emulation:     Emulation(data[2]),
		LoadSegment:   binary.LittleEndian.Uint16(data[4:6]),
		PartitionType: PartitionType(data[4]),
		size:          binary.LittleEndian.Uint16(data[6:8]),
		location:      binary.LittleEndian.Uint32(data[8:12]),
	}
}

func parseSectionEntry(data []byte) *ElToritoEntry {
	return &ElToritoEntry{
		Platform:      Platform(data[1]),
		Emulation:     Emulation(data[2]),
		LoadSegment:   binary.LittleEndian.Uint16(data[4:6]),
		PartitionType: PartitionType(data[4]),
		size:          binary.LittleEndian.Uint16(data[6:8]),
		location:      binary.LittleEndian.Uint32(data[8:12]),
	}
}

func parseValidationEntry(data []byte) error {
	if len(data) < 32 {
		return fmt.Errorf("Validation Entry: data too short")
	}
	if data[0] != 0x01 {
		return fmt.Errorf("Validation Entry: invalid header ID %x", data[0])
	}
	checksum := uint16(0)
	for i := 0; i < 32; i += 2 {
		checksum += binary.LittleEndian.Uint16(data[i : i+2])
	}
	if checksum != 0 {
		return fmt.Errorf("Validation Entry: checksum invalid")
	}
	if data[0x1E] != 0x55 || data[0x1F] != 0xAA {
		return fmt.Errorf("Validation Entry: invalid key bytes %x%x", data[0x1E], data[0x1F])
	}
	return nil
}
