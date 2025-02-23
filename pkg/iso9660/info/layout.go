package info

import (
	"fmt"
	"github.com/fatih/color"
	"io"
	"slices"
)

func NewISOLayout() *ISOLayout {
	return &ISOLayout{
		Objects: make([]ImageObject, 0),
	}
}

type ISOLayout struct {
	Objects []ImageObject
}

func (i *ISOLayout) AddObject(obj ImageObject) {
	i.Objects = append(i.Objects, obj)
}

func (i *ISOLayout) GetObjects() []ImageObject {
	return i.Objects
}

// Print outputs the ISO layout details to the given writer (w).
// - `verbose` prints all details if true; otherwise, prints a high-level summary.
// - `useColor` controls whether colored output is used.
// - `useHexOffset` prints offsets in hexadecimal if true.
func (i *ISOLayout) Print(w io.Writer, verbose, useColor, useHexOffset bool) {
	objects := i.GetObjects()

	// Remove nil objects and sort by Offset()
	var filteredObjects []ImageObject
	for _, obj := range objects {
		if obj != nil {
			filteredObjects = append(filteredObjects, obj)
		}
	}

	slices.SortFunc(filteredObjects, func(a, b ImageObject) int {
		return int(a.Offset() - b.Offset())
	})

	// Define color functions (use plain text if useColor is false)
	colorMap := map[string]func(a ...interface{}) string{
		"System Area":       color.New(color.FgBlue, color.Bold).SprintFunc(),
		"Volume Descriptor": color.New(color.FgYellow, color.Bold).SprintFunc(),
		"Path Table":        color.New(color.FgMagenta, color.Bold).SprintFunc(),
		"Directory Record":  color.New(color.FgCyan, color.Bold).SprintFunc(),
		"Directory Extent":  color.New(color.FgGreen, color.Bold).SprintFunc(),
		"File Extent":       color.New(color.FgRed, color.Bold).SprintFunc(),
	}

	offsetColor := color.New(color.FgGreen).SprintFunc()
	lengthColor := color.New(color.FgGreen).SprintFunc()

	if !useColor {
		for key := range colorMap {
			colorMap[key] = func(a ...interface{}) string { return fmt.Sprint(a...) }
		}
		offsetColor = func(a ...interface{}) string { return fmt.Sprint(a...) }
		lengthColor = func(a ...interface{}) string { return fmt.Sprint(a...) }
	}

	// Print header
	fmt.Fprintln(w, color.New(color.FgCyan, color.Bold).Sprint("\n=== ISO Layout Details ==="))

	// Fixed width settings
	offsetWidth := 14   // Width for [Offset: x]
	categoryWidth := 20 // Width for category names
	lengthWidth := 12   // Width for [Length: x bytes]

	if useHexOffset {
		offsetWidth = 18 // Width for [Offset: 0x...]
	}

	// Print each object in sorted order
	for _, obj := range filteredObjects {
		// Format offset as hex if requested
		offsetStr := fmt.Sprintf("Offset: %*d", offsetWidth-8, obj.Offset())
		if useHexOffset {
			offsetStr = fmt.Sprintf("Offset: %#*x", offsetWidth-8, obj.Offset())
		}

		fmt.Fprintf(w, "[%s] [%s] [%s] %s\n",
			offsetColor(offsetStr), // Offset (decimal or hex)
			colorMap[obj.Type()](fmt.Sprintf("%-*s", categoryWidth, obj.Type())), // Category
			lengthColor(fmt.Sprintf("%*s", lengthWidth, formatSize(obj.Size()))), // Size
			obj.Name(), // Object name
		)
	}

	// Print footer
	fmt.Fprintln(w, color.New(color.FgCyan, color.Bold).Sprint("=============================="))
}

//// PrettyJSON returns a pretty-printed JSON representation of the ISO layout.
//func (i *ISOLayout) PrettyJSON() string {
//	data, err := json.MarshalIndent(i, "", "  ")
//	if err != nil {
//		return fmt.Sprintf("Error generating JSON: %v", err)
//	}
//	return string(data)
//}
//
////TODO: This should be changed to not print directly and use a writer or return a string instead to decouple it.
//
//// Print outputs the ISO layout details in the order they occur in the ISO.
//// - `verbose` prints all details if true; otherwise, prints a high-level summary.
//// - `useColor` controls whether colored output is used.
//// - `useHexOffset` prints offsets in hexadecimal if true.
//func (i *ISOLayout) Print(verbose bool, useColor bool, useHexOffset bool) {
//	type layoutItem struct {
//		Offset   int
//		Length   int
//		Detail   string
//		Category string
//		IsDir    *bool // Pointer to handle directory indicator coloring
//	}
//
//	var items []layoutItem
//
//	// Define color functions (default to no color if useColor is false)
//	colorMap := map[string]func(a ...interface{}) string{
//		"System Area":       color.New(color.FgBlue, color.Bold).SprintFunc(),
//		"Volume Descriptor": color.New(color.FgYellow, color.Bold).SprintFunc(),
//		"Directory Record":  color.New(color.FgCyan, color.Bold).SprintFunc(),
//		"Path Table":        color.New(color.FgMagenta, color.Bold).SprintFunc(),
//		"Directory Extent":  color.New(color.FgGreen, color.Bold).SprintFunc(),
//	}
//
//	offsetColor := color.New(color.FgGreen).SprintFunc()
//	lengthColor := color.New(color.FgGreen).SprintFunc()
//	isDirTrueColor := color.New(color.FgHiYellow).SprintFunc() // Brown/Yellow for true
//	isDirFalseColor := color.New(color.FgBlue).SprintFunc()    // Blue for false
//
//	if !useColor {
//		for key := range colorMap {
//			colorMap[key] = func(a ...interface{}) string { return fmt.Sprint(a...) }
//		}
//		offsetColor = func(a ...interface{}) string { return fmt.Sprint(a...) }
//		lengthColor = func(a ...interface{}) string { return fmt.Sprint(a...) }
//		isDirTrueColor = offsetColor
//		isDirFalseColor = offsetColor
//	}
//
//	// System Area
//	items = append(items, layoutItem{
//		Offset:   i.SystemAreaOffset,
//		Length:   i.SystemAreaLength,
//		Detail:   "System Area",
//		Category: "System Area",
//	})
//
//	// Boot Catalog (if present)
//	if i.BootCatalogOffset > 0 {
//		items = append(items, layoutItem{
//			Offset:   i.BootCatalogOffset,
//			Length:   i.BootCatalogLength,
//			Detail:   fmt.Sprintf("Boot Catalog - System: %s", i.BootCatalogSystem),
//			Category: "Volume Descriptor",
//		})
//	}
//
//	// Volume Descriptor Set
//	for _, vd := range i.VolumeDescriptors {
//		items = append(items, layoutItem{
//			Offset:   vd.DescriptorOffset,
//			Length:   vd.DescriptorLength,
//			Detail:   fmt.Sprintf("%s (Version: %d)", vd.DescriptorType, vd.DescriptorVersion),
//			Category: "Volume Descriptor",
//		})
//	}
//
//	// Path Tables
//	for _, pt := range i.PathTables {
//		items = append(items, layoutItem{
//			Offset:   pt.PathTableOffset,
//			Length:   pt.PathTableLength,
//			Detail:   fmt.Sprintf("%s (Encoding: %s)", pt.PathTableSource, pt.PathTableEncoding),
//			Category: "Path Table",
//		})
//	}
//
//	// Directory Records
//	for _, dr := range i.DirectoryRecords {
//		isDir := dr.DirectoryRecordIsDirectory // Store isDir flag
//		items = append(items, layoutItem{
//			Offset:   dr.DirectoryRecordOffset,
//			Length:   dr.DirectoryRecordExtentLength,
//			Detail:   fmt.Sprintf("%s (Extent Location: %d)", dr.DirectoryRecordIdentifier, dr.DirectoryRecordExtentLocation),
//			Category: "Directory Record",
//			IsDir:    &isDir,
//		})
//	}
//
//	// Directory Extents
//	for _, de := range i.DirectoryExtents {
//		items = append(items, layoutItem{
//			Offset:   de.DirectoryExtentOffset,
//			Length:   de.DirectoryExtentLength,
//			Detail:   fmt.Sprintf("%s", de.DirectoryExtentIdentifier),
//			Category: "Directory Extent",
//		})
//	}
//
//	// Sort all items by Offset
//	slices.SortFunc(items, func(a, b layoutItem) int {
//		return a.Offset - b.Offset
//	})
//
//	// Print header
//	fmt.Println(color.New(color.FgCyan, color.Bold).Sprint("\n=== ISO Layout ==="))
//
//	// Fixed width settings
//	offsetWidth := 14   // Width for [Offset: x]
//	categoryWidth := 20 // Width for category names
//	lengthWidth := 12   // Width for [Length: x bytes]
//
//	if useHexOffset {
//		offsetWidth = 18 // Width for [Offset: 0x...]
//	}
//
//	// Print all sorted items with improved formatting
//	for _, item := range items {
//		// Format offset as hex if requested
//		offsetStr := fmt.Sprintf("Offset: %*d", offsetWidth-8, item.Offset)
//		if useHexOffset {
//			offsetStr = fmt.Sprintf("Offset: %#*x", offsetWidth-8, item.Offset)
//		}
//
//		// Apply color only to "true" or "false"
//		isDirStr := ""
//		if item.IsDir != nil {
//			if *item.IsDir {
//				isDirStr = fmt.Sprintf(" (IsDir: %s)", isDirTrueColor("true"))
//			} else {
//				isDirStr = fmt.Sprintf(" (IsDir: %s)", isDirFalseColor("false"))
//			}
//		}
//
//		fmt.Printf("[%s] [%s] [%s] %s%s\n",
//			offsetColor(offsetStr), // Offset (decimal or hex)
//			colorMap[item.Category](fmt.Sprintf("%-*s", categoryWidth, item.Category)), // Category
//			lengthColor(fmt.Sprintf("%*s", lengthWidth, formatSize(item.Length))),      // Human-readable length
//			item.Detail, // Description
//			isDirStr,    // Colored "true"/"false" value
//		)
//	}
//
//	// Print footer
//	fmt.Println(color.New(color.FgCyan, color.Bold).Sprint("=============================="))
//}

// formatSize converts a size in bytes to a human-readable format.
func formatSize(size int) string {
	const (
		MB = 1024 * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%8.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%8.2f MB", float64(size)/float64(MB))
	default:
		return fmt.Sprintf("%8d B ", size) // Ensures 'B' aligns with MB/GB
	}
}
