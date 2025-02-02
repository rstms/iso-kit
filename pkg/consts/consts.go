package consts

const (
	ISO9660_STD_IDENTIFIER      = "CD001"
	ISO9660_VOLUME_DESC_VERSION = 0x1
	ISO9660_SECTOR_SIZE         = 2048
	EL_TORITO_BOOT_SYSTEM_ID    = "EL TORITO SPECIFICATION"
	JOLIET__LEVEL_1_ESCAPE      = "%/@"
	JOLIET__LEVEL_2_ESCAPE      = "%/C"
	JOLIET__LEVEL_3_ESCAPE      = "%/E"
)

// ISOType represents the type of ISO image
type ISOType int

const (
	TYPE_ISO9660 ISOType = iota
	TYPE_UDF
)
