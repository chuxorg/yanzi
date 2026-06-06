package packs

// RoleBits is the permission bitmask type for role access checks.
type RoleBits uint8

// System role bit values. These are immutable and shipped with yanzi.
// User-visible labels are cosmetic; the bits are the reliable reference.
const (
	RoleObserver       RoleBits = 1   // 0b00000001 read only
	RoleAgent          RoleBits = 3   // 0b00000011 read + write artifacts
	RoleEngineer       RoleBits = 7   // 0b00000111 agent + code operations
	RoleQA             RoleBits = 15  // 0b00001111 engineer + validation
	RolePM             RoleBits = 31  // 0b00011111 qa + decisions
	RoleReleaseSteward RoleBits = 63  // 0b00111111 pm + release ops
	RoleAdmin          RoleBits = 255 // 0b11111111 all permissions
)

// Role describes a system role with its canonical name and display label.
type Role struct {
	Bits        RoleBits
	SystemName  string // canonical internal name
	Label       string // user-defined display label
	Description string
}

// systemRoles is the registry of all system roles in ascending bit order.
var systemRoles = []Role{
	{Bits: RoleObserver, SystemName: "observer", Label: "Observer", Description: "Read only, no destructive operations"},
	{Bits: RoleAgent, SystemName: "agent", Label: "Agent", Description: "Read and write artifacts"},
	{Bits: RoleEngineer, SystemName: "engineer", Label: "Engineer", Description: "Agent permissions plus code operations"},
	{Bits: RoleQA, SystemName: "qa", Label: "QA", Description: "Engineer permissions plus validation"},
	{Bits: RolePM, SystemName: "pm", Label: "PM", Description: "QA permissions plus decisions"},
	{Bits: RoleReleaseSteward, SystemName: "release-steward", Label: "Release Steward", Description: "PM permissions plus release operations"},
	{Bits: RoleAdmin, SystemName: "admin", Label: "Admin", Description: "All permissions"},
}

// Includes reports whether r has all bits set that required has.
func (r RoleBits) Includes(required RoleBits) bool {
	return r&required == required
}

// RoleFromBits returns the system role matching the given bits exactly.
// Returns false if no system role matches.
func RoleFromBits(bits RoleBits) (Role, bool) {
	for _, r := range systemRoles {
		if r.Bits == bits {
			return r, true
		}
	}
	return Role{}, false
}

// DefaultRole returns RoleObserver — the safe default (read only).
func DefaultRole() Role {
	return systemRoles[0]
}

// SystemRoles returns all system roles in ascending bit order.
func SystemRoles() []Role {
	out := make([]Role, len(systemRoles))
	copy(out, systemRoles)
	return out
}
