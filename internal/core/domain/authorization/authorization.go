package authorization

import (
	"go_hex/internal/core/domain/common"
	"go_hex/internal/support/validation"
)

// Permission represents a specific action that can be performed in the system.
type Permission string

// Define domain permissions as constants
const (
	PermissionAddFriend    Permission = "add_friend"
	PermissionViewFriend   Permission = "view_friend"
	PermissionUpdateFriend Permission = "update_friend"
	PermissionDeleteFriend Permission = "delete_friend"
	PermissionGreet        Permission = "greet"
)

// Role represents a user role in the system.
type Role string

// Define domain roles as constants
const (
	RoleAdmin    Role = "admin"
	RoleUser     Role = "user"
	RoleReadOnly Role = "readonly"
)

// rolePermissions maps roles to their allowed permissions
var rolePermissions = map[Role][]Permission{
	RoleAdmin: {
		PermissionAddFriend,
		PermissionViewFriend,
		PermissionUpdateFriend,
		PermissionDeleteFriend,
		PermissionGreet,
	},
	RoleUser: {
		PermissionAddFriend,
		PermissionViewFriend,
		PermissionUpdateFriend,
		PermissionGreet,
	},
	RoleReadOnly: {
		PermissionViewFriend,
		PermissionGreet,
	},
}

// ParseRole converts a string to a Role enum.
func ParseRole(roleStr string) Role {
	switch roleStr {
	case "admin":
		return RoleAdmin
	case "user":
		return RoleUser
	case "readonly":
		return RoleReadOnly
	default:
		return RoleReadOnly // Default to most restrictive role
	}
}

// GetPermissions returns the permissions associated with a role.
func (r Role) GetPermissions() []Permission {
	permissions, exists := rolePermissions[r]
	if !exists {
		return []Permission{} // Return empty permissions for unknown roles
	}
	return permissions
}

// HasPermission checks if the role has a specific permission.
func (r Role) HasPermission(permission Permission) bool {
	permissions := r.GetPermissions()
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// AuthorizationContext contains user information for authorization decisions.
type AuthorizationContext struct {
	UserID   string `json:"user_id" validate:"required,min=1"`
	Username string `json:"username" validate:"required,min=2"`
	Roles    []Role `json:"roles" validate:"required,min=1"`
}

// NewAuthorizationContext creates a new authorization context from string roles.
func NewAuthorizationContext(userID, username string, roles []string) AuthorizationContext {
	domainRoles := make([]Role, len(roles))
	for i, roleStr := range roles {
		domainRoles[i] = ParseRole(roleStr)
	}

	return AuthorizationContext{
		UserID:   userID,
		Username: username,
		Roles:    domainRoles,
	}
}

// HasPermission checks if the authorization context has a specific permission.
func (ac AuthorizationContext) HasPermission(permission Permission) bool {
	for _, role := range ac.Roles {
		if role.HasPermission(permission) {
			return true
		}
	}
	return false
}

// HasRole checks if the authorization context has a specific role.
func (ac AuthorizationContext) HasRole(role Role) bool {
	for _, r := range ac.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// IsAdmin checks if the authorization context has admin privileges.
func (ac AuthorizationContext) IsAdmin() bool {
	return ac.HasRole(RoleAdmin)
}

// Validate validates the authorization context using the domain validation rules.
func (ac AuthorizationContext) Validate() error {
	if err := validation.Validate(ac); err != nil {
		return common.NewDomainValidationError("invalid authorization context", err)
	}
	return nil
}
