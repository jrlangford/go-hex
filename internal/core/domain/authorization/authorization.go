package authorization

import (
	"go_hex/internal/core/domain/shared"
	"go_hex/internal/support/validation"
)

type Permission string

type Role string

const (
	PermissionAddFriend    Permission = "add_friend"
	PermissionViewFriend   Permission = "view_friend"
	PermissionUpdateFriend Permission = "update_friend"
	PermissionDeleteFriend Permission = "delete_friend"
	PermissionGreet        Permission = "greet"
)

const (
	RoleAdmin    Role = "admin"
	RoleUser     Role = "user"
	RoleReadOnly Role = "readonly"
)

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

func (r Role) GetPermissions() []Permission {
	permissions, exists := rolePermissions[r]
	if !exists {
		return []Permission{} // Return empty permissions for unknown roles
	}
	return permissions
}

func (r Role) HasPermission(permission Permission) bool {
	permissions := r.GetPermissions()
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

type AuthorizationContext struct {
	UserID   string `json:"user_id" validate:"required,min=1"`
	Username string `json:"username" validate:"required,min=2"`
	Roles    []Role `json:"roles" validate:"required,min=1"`
}

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

func (ac AuthorizationContext) HasPermission(permission Permission) bool {
	for _, role := range ac.Roles {
		if role.HasPermission(permission) {
			return true
		}
	}
	return false
}

func (ac AuthorizationContext) HasRole(role Role) bool {
	for _, r := range ac.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func (ac AuthorizationContext) IsAdmin() bool {
	return ac.HasRole(RoleAdmin)
}

func (ac AuthorizationContext) Validate() error {
	if err := validation.Validate(ac); err != nil {
		return shared.NewDomainValidationError("invalid authorization context", err)
	}
	return nil
}
