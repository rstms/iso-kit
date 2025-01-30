package testing

import (
	"encoding/json"
	"fmt"
	"github.com/bgrewell/iso-kit/pkg/directory"
	"os"
	"strings"
)

// ContainsNonASCIIPrintable returns true if the string has any
// characters outside ASCII [32..126], i.e., not a standard printable.
func ContainsNonASCIIPrintable(s string) bool {
	for _, r := range s {
		// If it's outside the ASCII printable range, return true.
		if r < 32 || r > 126 {
			return true
		}
	}
	return false
}

// Validate compares BFS entries against ground-truth JSON.
//   - entries: slice of BFSAllEntries results
//   - gtPath: path to the ground_truth.json
func Validate(entries []*directory.DirectoryEntry, gtPath string) error {
	groundTruth, err := LoadGroundTruth(gtPath)
	if err != nil {
		return err
	}

	// Build a map for quick lookups in BFS results.
	// Key is the full path or name, value is the BFS DirectoryEntry.
	bfsMap := make(map[string]*directory.DirectoryEntry)
	for _, e := range entries {
		// Skip the root entry
		if e.IsRootEntry() {
			continue
		}

		bfsMap[e.FullPath()] = e
		if ContainsNonASCIIPrintable(e.Name()) {
			return fmt.Errorf("non-ASCII printable characters in entry: %s", e.Name())
		}
	}

	// Build a map for ground truth entries.
	// Key is the "Name" field.
	gtMap := make(map[string]GroundTruthEntry)
	for _, gt := range groundTruth {
		gtMap[gt.Name] = gt
	}

	// 1) Identify missing entries (in ground truth but not BFS)
	var missing []GroundTruthEntry
	for name, gt := range gtMap {
		if _, found := bfsMap[name]; !found {
			missing = append(missing, gt)
		}
	}

	// 2) Identify extra entries (in BFS but not ground truth)
	var extra []*directory.DirectoryEntry
	for name, bfs := range bfsMap {
		if _, found := gtMap[name]; !found {
			extra = append(extra, bfs)
		}
	}

	// Print a nicely formatted summary.
	fmt.Println(strings.Repeat("=", 40))
	fmt.Println("VALIDATION RESULTS")
	fmt.Println(strings.Repeat("=", 40))

	if len(missing) == 0 && len(extra) == 0 {
		fmt.Println("All entries match the ground truth!")
		return nil
	}

	if len(missing) > 0 {
		fmt.Println("Missing entries (in ground truth, not in BFS):")
		for _, m := range missing {
			t := "FILE"
			if m.IsDirectory {
				t = "DIR"
			}
			fmt.Printf("  - [%s] %s\n", t, m.Name)
		}
	} else {
		fmt.Println("No missing entries.")
	}

	if len(extra) > 0 {
		fmt.Println("\nExtra entries (in BFS, not in ground truth):")
		for _, x := range extra {
			t := "FILE"
			if x.IsDir() {
				t = "DIR"
			}
			fmt.Printf("  - [%s] %s\n", t, x.FullPath())
		}
	} else {
		fmt.Println("No extra entries.")
	}

	fmt.Println(strings.Repeat("=", 40))
	return nil
}

// GroundTruthEntry represents a single record from the JSON.
type GroundTruthEntry struct {
	Date           string `json:"date"`
	Time           string `json:"time"`
	Attr           string `json:"attr"`
	Size           int64  `json:"size"`
	CompressedSize int64  `json:"compressed_size"`
	Name           string `json:"name"`
	IsDirectory    bool   `json:"is_directory"`
}

// LoadGroundTruth reads the JSON from a file and unmarshals it into a slice.
func LoadGroundTruth(filePath string) ([]GroundTruthEntry, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var entries []GroundTruthEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return entries, nil
}
