package main

import (
	"fmt"
	"github.com/bgrewell/iso-kit"
	"github.com/bgrewell/usage"
	"os"
)

// DisplayISOInfo prints general information about the ISO file.
func DisplayISOInfo(i iso.ISO, verbose bool) {
	//// Retrieve primary volume descriptor (PVD)
	//pvd := i.Pri
	//if pvd == nil {
	//	u.PrintError(fmt.Errorf("failed to retrieve primary volume descriptor"))
	//	return
	//}
	//
	//// Retrieve file system entries
	//entries, err := i.BuildFileSystemEntries(i.RootDirectory, true) // Rock Ridge enabled
	//if err != nil {
	//	u.PrintError(fmt.Errorf("failed to parse file system entries: %w", err))
	//	return
	//}
	//
	//// Count files and directories
	//fileCount, dirCount, symlinkCount := 0, 0, 0
	//totalSize := uint64(0)
	//
	//for _, entry := range entries {
	//	if entry.IsDir {
	//		dirCount++
	//	} else {
	//		fileCount++
	//		totalSize += uint64(entry.Size)
	//	}
	//
	//	// Count symbolic links if Rock Ridge is enabled
	//	if entry.Mode&os.ModeSymlink != 0 {
	//		symlinkCount++
	//	}
	//}

	// Print Basic Information
	fmt.Println("=== ISO Information ===")
	fmt.Printf("Volume Name: %s\n", i.GetVolumeID())
	fmt.Printf("Created By: %s\n", i.GetApplicationID())
	fmt.Printf("Preparer: %s\n", i.GetDataPreparerID())
	fmt.Printf("Publisher: %s\n", i.GetPublisherID())
	fmt.Printf("Volume Size: %d sectors (%d MB)\n", -1, (-1*2048)/1024/1024)
	fmt.Printf("Total Files: %d\n", -1)
	fmt.Printf("Total Directories: %d\n", -1)
	fmt.Printf("Total Size: %d bytes (%.2f MB)\n", -1, float64(-1)/1024/1024)

	if verbose {
		// Verbose output with additional metadata
		fmt.Println("\n=== Verbose Information ===")
		fmt.Printf("System Identifier: %s\n", i.GetSystemID())
		fmt.Printf("Volume Set Size: %d\n", i.GetVolumeSetID())
		fmt.Printf("Volume Sequence Number: %d\n", -1)
		fmt.Printf("Logical Block Size: %d bytes\n", -1)
		fmt.Printf("Number of Hard Links: %d\n", -1)
		fmt.Printf("Symbolic Links: %d\n", -1)
		//fmt.Printf("Root Directory Location: %d (LBA)\n", i.RootDirectory.Location)

		// Rock Ridge Support
		if i.HasRockRidge() {
			fmt.Println("\n--- Rock Ridge Extensions ---")
			fmt.Println("Rock Ridge Enabled: YES")
			fmt.Printf("Number of Files with Extended Attributes: %d\n", -1)
		} else {
			fmt.Println("\nRock Ridge Extensions: NOT PRESENT")
		}
	}

	fmt.Println("=========================")
}

func main() {

	u := usage.NewUsage()
	help := u.AddBooleanOption("h", "help", false, "Show this help message", "optional", nil)
	verbose := u.AddBooleanOption("v", "verbose", false, "Print verbose output", "", nil)
	path := u.AddArgument(1, "iso-path", "Path to the file within the ISO to read", "")
	parsed := u.Parse()

	if !parsed {
		u.PrintError(fmt.Errorf("failed to parse arguments"))
		os.Exit(1)
	}

	if *help {
		u.PrintUsage()
		os.Exit(0)
	}

	if path == nil || *path == "" {
		u.PrintError(fmt.Errorf("location of the iso file <path> must be provided"))
		os.Exit(1)
	}

	_ = verbose

	i, err := iso.Open(*path)
	if err != nil {
		u.PrintError(err)
		os.Exit(1)
	}

	DisplayISOInfo(i, *verbose)
	_ = i

}
