package consts

const (
	ISO9660_STD_IDENTIFIER      = "CD001"
	ISO9660_VOLUME_DESC_VERSION = 0x1
	ISO9660_SECTOR_SIZE         = 2048
	EL_TORITO_BOOT_SYSTEM_ID    = "EL TORITO SPECIFICATION"
	JOLIET__LEVEL_1_ESCAPE      = "%/@"
	JOLIET__LEVEL_2_ESCAPE      = "%/C"
	JOLIET__LEVEL_3_ESCAPE      = "%/E"

	// a-characters: 57 characters in the following positions of the International Reference Version
	//    2/0 - 2/2
	//    2/5 - 2/15
	//    3/0 - 3/15
	//    4/1 - 4/15
	//    5/0 - 5/10
	//    5/15
	A_CHARACTERS = " !\"%&'()*+,-./0123456789:;<=>?ABCDEFGHIJKLMNOPQRSTUVWXYZ_"

	// NOTE: This relates to SVD volumes, specifically Joliet
	// a1-characters: A subset of the c-characters. This subset shall be subject to agreement between the originator
	//                and the recipient of the volume.
	A1_CHARACTERS = C_CHARACTERS

	// NOTE: This relates to SVD volumes, specifically Joliet
	// c-characters: The characters of the coded graphic character sets identified by the escape sequences in a SVD
	C_CHARACTERS = ""

	// d-characters: 37 characters in the following positions of the International Reference Version
	//    3/0 - 3/9
	//    4/1 - 5/10
	//    5/15
	D_CHARACTERS = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_"

	// NOTE: This relates to SVD volumes, specifically Joliet
	// d1-characters: A subset of the a1-characters. This subset shall be subject to agreement between the originator
	//                and the recipient of the volume
	D1_CHARACTERS = A1_CHARACTERS

	SEPERATOR_1 = "." // 0x2E
	SEPERATOR_2 = ";" // 0x3B

	FILLER = " " // 0x20
)

// ISOType represents the type of ISO image
type ISOType int

const (
	TYPE_ISO9660 ISOType = iota
	TYPE_UDF
)
