package xattr

import "fmt"

// GroupReadPermission represents the two-bit value stored in bits 8-9.
// According to the spec:
//
//	0 means that any user in the group may read the file,
//	2 means that only the owner (of that group) may read the file.
type GroupReadPermission uint8

const (
	GroupReadAllowed    GroupReadPermission = 0
	GroupReadRestricted GroupReadPermission = 2
)

// ExtendedAttrPermissions holds the permission flags from an Extended Attribute Record.
// The 16-bit field is broken out as follows (bit positions are 0-indexed):
//
//	Bit 0:  System class owner read denied (false means allowed; true means denied)
//	Bit 1:  Fixed; shall be 1
//	Bit 2:  System class owner execute denied (false allowed; true denied)
//	Bit 3:  Fixed; shall be 1
//	Bit 4:  Owner read denied (false allowed; true denied)
//	Bit 5:  Fixed; shall be 1
//	Bit 6:  Owner execute denied (false allowed; true denied)
//	Bit 7:  Fixed; shall be 1
//	Bits 8-9: Group read permission, where allowed values are:
//	          0 (GroupReadAllowed) means any group member may read,
//	          2 (GroupReadRestricted) means only the owner may read.
//	Bit 10: Group execute restricted (false means allowed; true means only owner may execute)
//	Bit 11: Fixed; shall be 1
//	Bit 12: Other (world) read denied (false allowed; true denied)
//	Bit 13: Fixed; shall be 1
//	Bit 14: Other (world) execute denied (false allowed; true denied)
//	Bit 15: Fixed; shall be 1
type ExtendedAttrPermissions struct {
	// Bit 0
	SystemReadDenied bool `json:"system_read_denied"`
	// Bit 2
	SystemExecuteDenied bool `json:"system_execute_denied"`
	// Bit 4
	OwnerReadDenied bool `json:"owner_read_denied"`
	// Bit 6
	OwnerExecuteDenied bool `json:"owner_execute_denied"`
	// Bits 8-9; allowed values: 0 (allowed) or 2 (restricted)
	GroupReadPermission GroupReadPermission `json:"group_read_permission"`
	// Bit 10
	GroupExecuteRestricted bool `json:"group_execute_restricted"`
	// Bit 12
	OtherReadDenied bool `json:"other_read_denied"`
	// Bit 14
	OtherExecuteDenied bool `json:"other_execute_denied"`
}

// Marshal returns the 16-bit value (as a uint16) representing the permission flags.
// Fixed bits are forced to 1 as specified.
func (eap ExtendedAttrPermissions) Marshal() uint16 {
	var flags uint16 = 0

	// Bit 0: SystemReadDenied
	if eap.SystemReadDenied {
		flags |= 1 << 0
	}
	// Bit 1: fixed to 1.
	flags |= 1 << 1

	// Bit 2: SystemExecuteDenied
	if eap.SystemExecuteDenied {
		flags |= 1 << 2
	}
	// Bit 3: fixed to 1.
	flags |= 1 << 3

	// Bit 4: OwnerReadDenied
	if eap.OwnerReadDenied {
		flags |= 1 << 4
	}
	// Bit 5: fixed to 1.
	flags |= 1 << 5

	// Bit 6: OwnerExecuteDenied
	if eap.OwnerExecuteDenied {
		flags |= 1 << 6
	}
	// Bit 7: fixed to 1.
	flags |= 1 << 7

	// Bits 8-9: Group Read Permission.
	// We place the value (expected to be 0 or 2) in bits 8-9.
	flags |= uint16(eap.GroupReadPermission) << 8

	// Bit 10: GroupExecuteRestricted.
	if eap.GroupExecuteRestricted {
		flags |= 1 << 10
	}
	// Bit 11: fixed to 1.
	flags |= 1 << 11

	// Bit 12: OtherReadDenied.
	if eap.OtherReadDenied {
		flags |= 1 << 12
	}
	// Bit 13: fixed to 1.
	flags |= 1 << 13

	// Bit 14: OtherExecuteDenied.
	if eap.OtherExecuteDenied {
		flags |= 1 << 14
	}
	// Bit 15: fixed to 1.
	flags |= 1 << 15

	return flags
}

// UnmarshalExtendedAttrPermissions decodes a 16-bit permission field into an ExtendedAttrPermissions struct.
// It verifies that the fixed bits (positions 1,3,5,7,11,13,15) are set.
func UnmarshalExtendedAttrPermissions(flags uint16) (ExtendedAttrPermissions, error) {
	// Verify fixed bits.
	fixedMask := uint16((1 << 1) | (1 << 3) | (1 << 5) | (1 << 7) | (1 << 11) | (1 << 13) | (1 << 15))
	if flags&fixedMask != fixedMask {
		return ExtendedAttrPermissions{}, fmt.Errorf("invalid permissions: fixed bits not all set in 0x%04X", flags)
	}

	// Extract configurable bits.
	eap := ExtendedAttrPermissions{
		SystemReadDenied:       flags&(1<<0) != 0,
		SystemExecuteDenied:    flags&(1<<2) != 0,
		OwnerReadDenied:        flags&(1<<4) != 0,
		OwnerExecuteDenied:     flags&(1<<6) != 0,
		GroupExecuteRestricted: flags&(1<<10) != 0,
		OtherReadDenied:        flags&(1<<12) != 0,
		OtherExecuteDenied:     flags&(1<<14) != 0,
	}
	// Bits 8-9: extract the 2-bit value.
	groupRead := uint8((flags >> 8) & 0x03)
	if groupRead != 0 && groupRead != 2 {
		return ExtendedAttrPermissions{}, fmt.Errorf("invalid group read permission value: %d (expected 0 or 2)", groupRead)
	}
	eap.GroupReadPermission = GroupReadPermission(groupRead)

	return eap, nil
}
