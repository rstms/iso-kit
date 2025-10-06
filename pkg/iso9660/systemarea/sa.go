package systemarea

import (
	"github.com/rstms/iso-kit/pkg/consts"
	"github.com/rstms/iso-kit/pkg/iso9660/info"
)

type SystemArea struct {
	// System Area's use isn't defined in the ISO 9660 standard. It is reserved for system use.
	Contents [consts.ISO9660_SECTOR_SIZE * consts.ISO9660_SYSTEM_AREA_SECTORS]byte
	// --- Fields that are not part of the ISO9660 object ---
	// Object Location (in bytes)
	ObjectLocation int64 `json:"object_location"`
	// Object Size (in bytes)
	ObjectSize uint32 `json:"object_size"`
}

func (s SystemArea) Type() string {
	return "System Area"
}

func (s SystemArea) Name() string {
	return "System Area"
}

func (s SystemArea) Description() string {
	return ""
}

func (s SystemArea) Properties() map[string]interface{} {
	return map[string]interface{}{}
}

func (s SystemArea) Offset() int64 {
	return s.ObjectLocation
}

func (s SystemArea) Size() int {
	return int(s.ObjectSize)
}

func (s SystemArea) GetObjects() []info.ImageObject {
	return []info.ImageObject{s}
}

func (s SystemArea) Marshal() ([]byte, error) {
	return s.Contents[:], nil
}
