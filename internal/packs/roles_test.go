package packs

import "testing"

func TestRoleBits_Includes(t *testing.T) {
	tests := []struct {
		r        RoleBits
		required RoleBits
		want     bool
	}{
		{RoleAdmin, RoleEngineer, true},
		{RoleEngineer, RoleAdmin, false},
		{RoleEngineer, RoleAgent, true},
		{RoleEngineer, RoleEngineer, true},
		{RoleObserver, RoleObserver, true},
		{RoleObserver, RoleAgent, false},
		{RoleAdmin, RoleAdmin, true},
	}
	for _, tt := range tests {
		got := tt.r.Includes(tt.required)
		if got != tt.want {
			t.Errorf("RoleBits(%d).Includes(%d) = %v, want %v", tt.r, tt.required, got, tt.want)
		}
	}
}

func TestDefaultRole(t *testing.T) {
	r := DefaultRole()
	if r.Bits != RoleObserver {
		t.Errorf("DefaultRole().Bits = %d, want %d (RoleObserver)", r.Bits, RoleObserver)
	}
}

func TestRoleFromBits(t *testing.T) {
	r, ok := RoleFromBits(RoleEngineer)
	if !ok {
		t.Fatal("RoleFromBits(RoleEngineer) returned false")
	}
	if r.SystemName != "engineer" {
		t.Errorf("SystemName = %q, want %q", r.SystemName, "engineer")
	}

	_, ok = RoleFromBits(42)
	if ok {
		t.Error("RoleFromBits(42) should return false for unknown bits")
	}
}

func TestSystemRoles_Order(t *testing.T) {
	roles := SystemRoles()
	if len(roles) == 0 {
		t.Fatal("SystemRoles() returned empty slice")
	}
	for i := 1; i < len(roles); i++ {
		if roles[i].Bits <= roles[i-1].Bits {
			t.Errorf("roles not in ascending order at index %d: %d <= %d", i, roles[i].Bits, roles[i-1].Bits)
		}
	}
}
