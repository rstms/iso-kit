package systemarea

import (
	"github.com/bgrewell/iso-kit/pkg/iso9660/consts"
)

type SystemArea struct {
	// System Area's use isn't defined in the ISO 9660 standard. It is reserved for system use.
	Contents [consts.ISO9660_SECTOR_SIZE * consts.ISO9660_SYSTEM_AREA_SECTORS]byte
}
