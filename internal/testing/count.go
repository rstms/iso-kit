package testing

import "github.com/bgrewell/iso-kit/pkg/directory"

func GetFileAndFolderCounts(rootDir *directory.DirectoryEntry) (int, int) {
	var folderCount, fileCount int

	// Function needs to be declared before it is assigned to the anonymous function so that it can
	// be called recursively.
	var walk func(d *directory.DirectoryEntry)

	// Anonymous function to walk the directory tree and count folders and files.
	walk = func(d *directory.DirectoryEntry) {
		if !d.IsRootEntry() {
			folderCount++
		}

		children, err := d.GetChildren()
		if err != nil {
			panic(err)
		}
		for _, child := range children {
			if child.IsDir() {
				walk(child)
			} else {
				if child.Record.FileIdentifier != "\x00" && child.Record.FileIdentifier != "\x01" {
					fileCount++
				}
			}
		}
	}

	walk(rootDir)
	return folderCount, fileCount
}
