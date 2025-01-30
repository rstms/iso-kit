package susp

import (
	"errors"
	"fmt"
	"github.com/bgrewell/iso-kit/pkg/logging"
	"github.com/bgrewell/iso-kit/pkg/rockridge"
	"github.com/go-logr/logr"
)

// NewSystemUseEntries creates a new SystemUseEntries instance
func NewSystemUseEntries(entries []*SystemUseEntry, logger logr.Logger) *SystemUseEntries {
	return &SystemUseEntries{
		entries: entries,
		logger:  logger,
	}
}

// SystemUseEntries is a slice of SystemUseEntry elements with some additional helper methods
type SystemUseEntries struct {
	entries []*SystemUseEntry
	logger  logr.Logger
}

func (e SystemUseEntries) Entries() []*SystemUseEntry {
	return e.entries
}

// Len returns the number of SystemUseEntry elements
func (e SystemUseEntries) Len() int {
	return len(e.entries)
}

func (e SystemUseEntries) GetExtensionRecords() (records []*ExtensionRecord, err error) {
	for _, entry := range e.entries {
		if entry.Type() == EXTENSION_REFERENCE {
			er, err := UnmarshalExtensionRecord(entry)
			if err != nil {
				return nil, err
			}
			records = append(records, er)
		}
	}
	return records, nil
}

// HasRockRidge returns true if the SystemUseEntries contains Rock Ridge extensions
func (e SystemUseEntries) HasRockRidge() bool {
	records, err := e.GetExtensionRecords()
	if err != nil {
		e.logger.Error(err, "Failed to get extension records")
		return false
	}

	for i, record := range records {
		e.logger.V(logging.TRACE).Info("ExtensionRecord", "record", record)
		if record.Identifier == rockridge.ROCK_RIDGE_IDENTIFIER && record.Version == rockridge.ROCK_RIDGE_VERSION {
			e.logger.V(logging.TRACE).Info("Found Rock Ridge extension", "index", i)
			return true
		}
	}

	// TODO: This is temporary until I figure out why the extension records aren't appearing for the actual items that
	//  have Rock Ridge extensions
	for _, entry := range e.entries {
		if entry.Type() == SystemUseEntryType(rockridge.POSIX_FILE_PERMS) ||
			entry.Type() == SystemUseEntryType(rockridge.ALTERNATE_NAME) ||
			entry.Type() == SystemUseEntryType(rockridge.TIME_STAMPS) {
			e.logger.V(logging.TRACE).Info("Found Rock Ridge extension")
			return true
		}
	}

	return false
}

// RockRidgeName returns the Rock Ridge name if present otherwise it returns nil
func (e SystemUseEntries) RockRidgeName() *string {
	for _, record := range e.entries {
		if record.Type() == SystemUseEntryType(rockridge.ALTERNATE_NAME) {
			var name string
			entry := rockridge.UnmarshalRockRidgeNameEntry(record.Length(), record.Data())
			if entry == nil {
				e.logger.Error(errors.New("failed to unmarshal Rock Ridge name entry"),
					"Failed to unmarshal Rock Ridge name entry", "data",
					fmt.Sprintf("%v", record.Data()))
				return nil
			}
			if entry.Current {
				name = "."
				return &name
			}
			if entry.Parent {
				name = ".."
				return &name
			}

			name = entry.Name
			e.logger.V(logging.TRACE).Info("Found Rock Ridge name", "name", name)
			return &name
		}
	}

	e.logger.V(logging.TRACE).Info("Rock Ridge alternate name not found")
	return nil
}

// RockRidgePermissions returns the Rock Ridge permissions if present otherwise it returns nil
func (e SystemUseEntries) RockRidgePermissions() *rockridge.RockRidgePosixEntry {
	for _, record := range e.entries {
		if record.Type() == SystemUseEntryType(rockridge.POSIX_FILE_PERMS) {
			e.logger.V(logging.TRACE).Info("Found Rock Ridge permissions")
			entry, err := rockridge.UnmarshalRockRidgePosixEntry(record.Data())
			if err != nil {
				e.logger.Error(err, "Failed to unmarshal Rock Ridge permissions entry")
				return nil
			}
			return entry
		}
	}

	e.logger.V(logging.TRACE).Info("Rock Ridge permissions not found")
	return nil
}

// RockRidgeTimestamps returns the Rock Ridge timestamps if present otherwise it returns nil
func (e SystemUseEntries) RockRidgeTimestamps() *rockridge.RockRidgeTimestamps {
	return nil
}
