package filesystem

import (
	"github.com/bgrewell/iso-kit/pkg/iso9660/directory"
	"os"
	"time"
)

type FileSystemEntry struct {
	// The name of the file or directory and any extension
	Name string `json:"name"`
	// Full path, e.g., "dir/subdir/file.txt"
	FullPath string `json:"full_path"`
	// IsDir, true if it's a directory
	IsDir bool `json:"is_dir"`
	// Size of the file, 0 if it's a directory
	Size uint32 `json:"size"`
	// Location of the file in the iso
	Location uint32 `json:"location"`
	// UID, userid of the file/directory
	UID *uint32 `json:"uid"`
	// GID, groupid of the file/directory
	GID *uint32 `json:"gid"`
	// Mode, permissions of the file/directory
	Mode os.FileMode
	// CreateTime
	CreateTime time.Time
	// ModTime
	ModTime time.Time
	// RockRidge extended attributes
	HasRockRidge bool `json:"has_rock_ridge"`
	// Original DirectoryRecord
	DirectoryRecord *directory.DirectoryRecord
}
