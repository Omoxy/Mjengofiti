package auth

// Role represents an authorized role in Mjengofiti.
type Role string

const (
	RoleViewer               Role = "viewer"
	RoleDataResearcher       Role = "data_researcher"
	RoleDataSteward          Role = "data_steward"
	RoleInspector            Role = "inspector"
	RoleInspectionSupervisor Role = "inspection_supervisor"
	RoleAdmin                Role = "admin"
)

// User represents an authenticated session user.
type User struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	DisplayName string   `json:"display_name"`
	Role        Role     `json:"role"`
	Permissions []string `json:"permissions"`
}

// HasPermission checks if the user possesses the requested permission or wildcard.
func (u *User) HasPermission(permission string) bool {
	if u == nil {
		return false
	}
	for _, p := range u.Permissions {
		if p == "*" || p == permission {
			return true
		}
	}
	return false
}

// NormalizeRole maps role aliases from database seed and external inputs to canonical Roles.
func NormalizeRole(value string) Role {
	role, ok := ParseRole(value)
	if !ok {
		return RoleViewer
	}
	return role
}

// ParseRole maps recognized role names and aliases without granting a fallback role.
func ParseRole(value string) (Role, bool) {
	switch value {
	case "viewer":
		return RoleViewer, true
	case "steward", "data_steward":
		return RoleDataSteward, true
	case "researcher", "data_researcher":
		return RoleDataResearcher, true
	case "supervisor", "inspection_supervisor":
		return RoleInspectionSupervisor, true
	case "inspector":
		return RoleInspector, true
	case "admin":
		return RoleAdmin, true
	default:
		return "", false
	}
}

// DefaultPermissionsForRole returns the standard permission set for a given role.
func DefaultPermissionsForRole(role Role) []string {
	switch role {
	case RoleAdmin:
		return []string{"*"}
	case RoleDataSteward:
		return []string{"buildings:read", "datasets:read", "datasets:write", "datasets:promote", "standards:read", "audit:read"}
	case RoleDataResearcher:
		return []string{"buildings:read", "datasets:read", "indicators:read", "indicators:compute", "standards:read"}
	case RoleInspectionSupervisor:
		return []string{"buildings:read", "inspections:read", "inspections:triage", "inspections:assign", "inspections:findings", "inspections:close", "standards:read", "audit:read"}
	case RoleInspector:
		return []string{"buildings:read", "inspections:read", "inspections:findings", "standards:read"}
	default:
		return []string{"buildings:read", "datasets:read", "standards:read"}
	}
}

// UserProfileResponse returns current user profile and available roles.
type UserProfileResponse struct {
	User           User            `json:"user"`
	AvailableRoles []RoleDefinition `json:"available_roles"`
}

// RoleDefinition describes a role and its permitted actions.
type RoleDefinition struct {
	Role        Role     `json:"role"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}
