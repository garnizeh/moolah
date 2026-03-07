package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRole_Level(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		role     Role
		expected int
	}{
		{"Sysadmin level", RoleSysadmin, 3},
		{"Admin level", RoleAdmin, 2},
		{"Member level", RoleMember, 1},
		{"Unknown level", Role("guest"), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.role.Level())
		})
	}
}

func TestRole_CanAccess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		role     Role
		target   Role
		expected bool
	}{
		{"Admin can access Member", RoleAdmin, RoleMember, true},
		{"Admin can access Admin", RoleAdmin, RoleAdmin, true},
		{"Member cannot access Admin", RoleMember, RoleAdmin, false},
		{"Sysadmin can access Admin", RoleSysadmin, RoleAdmin, true},
		{"Unknown cannot access Member", Role("guest"), RoleMember, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.role.CanAccess(tt.target))
		})
	}
}
