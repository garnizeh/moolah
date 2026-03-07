package domain

// Role represents a user's role within a household/tenant.
type Role string

const (
	// RoleSysadmin has full access to the entire system (global).
	RoleSysadmin Role = "sysadmin"
	// RoleAdmin has full access to a specific tenant.
	RoleAdmin Role = "admin"
	// RoleMember has standard member access to a specific tenant.
	RoleMember Role = "member"
)

// Level returns a numeric weight for the role to allow comparison.
func (r Role) Level() int {
	switch r {
	case RoleSysadmin:
		return 3
	case RoleAdmin:
		return 2
	case RoleMember:
		return 1
	default:
		return 0
	}
}

// CanAccess returns true if the current role is at least as privileged as the target role.
func (r Role) CanAccess(target Role) bool {
	return r.Level() >= target.Level()
}
