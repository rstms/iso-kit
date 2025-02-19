package consts

const (
	// Number of system area sectors.
	ISO9660_SYSTEM_AREA_SECTORS = 16

	// Standard ISO9660 identifier.
	ISO9660_STD_IDENTIFIER = "CD001"

	// ISO9660 volume descriptor version (always 1).
	ISO9660_VOLUME_DESC_VERSION = 1

	// ISO9660 default sector size.
	ISO9660_SECTOR_SIZE = 2048

	// ISO9660 volume descriptor header size
	ISO9660_VOLUME_DESC_HEADER_SIZE = 7

	// ISO9660 application use area size
	ISO9660_APPLICATION_USE_SIZE = 512

	// JOLIET level 1, 2, and 3 escape sequences.
	JOLIET_LEVEL_1_ESCAPE = "%/@"
	JOLIET_LEVEL_2_ESCAPE = "%/C"
	JOLIET_LEVEL_3_ESCAPE = "%/E"

	// El Torito bootable cdrom system identifier.
	EL_TORITO_BOOT_SYSTEM_ID = "EL TORITO SPECIFICATION"

	// a-characters set which are specified in the International Reference Version at the following positions.
	//   | 2/0 - 2/2
	//   | 2/5 - 2/15
	//   | 3/0 - 3/15
	//   | 4/1 - 4/15
	//   | 5/0 - 5/10
	//   | 5/15
	A_CHARACTERS = " !\"%&'()*+,-./0123456789:;<=>?ABCDEFGHIJKLMNOPQRSTUVWXYZ_"

	// c-characters set which are the coded graphic character sets identified by the escape sequences in a Joliet SVD.
	// | All code points between (00)(00) and (00)(1F), inclusive. (Control Characters)
	// | (00)(2A) '*'(Asterisk)
	// | (00)(2F) '/' (Forward Slash)
	// | (00)(3A) ':' (Colon)
	// | (00)(3B) ';' (Semicolon)
	// | (00)(3F) '?' (Question Mark)
	// | (00)(5C) '\' (Backslash)

	// a1-characters set which are a subset of the c-characters. This subset shall be subject to agreement between the
	// originator and the recipient of the volume.

	// d-characters: 37 characters in the following positions of the International Reference Version
	// | 3/0 - 3/9
	// | 4/1 - 5/10
	// | 5/15
	D_CHARACTERS = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_"

	// Separators allowed by ISO9660 0x2E and 0x3B.
	ISO9660_SEPARATOR_1 = "."
	ISO9660_SEPARATOR_2 = ";"

	// ISO9660 Filler 0x20 (space)
	ISO9660_FILLER = " "

	// Standard UDF Identifier
	UDF_STD_IDENTIFIER = "BEA01"

	// UDF default sector size.
	UDF_SECTOR_SIZE = 2048
)
