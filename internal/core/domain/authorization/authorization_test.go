package authorization_test

import (
	"go_hex/internal/core/domain/authorization"
	"testing"
)

func TestRole_GetPermissions(t *testing.T) {
	tests := []struct {
		name     string
		role     authorization.Role
		expected []authorization.Permission
	}{
		{
			name: "Admin has all permissions",
			role: authorization.RoleAdmin,
			expected: []authorization.Permission{
				authorization.PermissionAddFriend,
				authorization.PermissionViewFriend,
				authorization.PermissionUpdateFriend,
				authorization.PermissionDeleteFriend,
				authorization.PermissionGreet,
			},
		},
		{
			name: "User has limited permissions",
			role: authorization.RoleUser,
			expected: []authorization.Permission{
				authorization.PermissionAddFriend,
				authorization.PermissionViewFriend,
				authorization.PermissionUpdateFriend,
				authorization.PermissionGreet,
			},
		},
		{
			name: "ReadOnly has minimal permissions",
			role: authorization.RoleReadOnly,
			expected: []authorization.Permission{
				authorization.PermissionViewFriend,
				authorization.PermissionGreet,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permissions := tt.role.GetPermissions()

			if len(permissions) != len(tt.expected) {
				t.Errorf("Expected %d permissions, got %d", len(tt.expected), len(permissions))
				return
			}

			// Check that all expected permissions are present
			for _, expectedPerm := range tt.expected {
				found := false
				for _, perm := range permissions {
					if perm == expectedPerm {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected permission %s not found", expectedPerm)
				}
			}
		})
	}
}

func TestRole_HasPermission(t *testing.T) {
	tests := []struct {
		name       string
		role       authorization.Role
		permission authorization.Permission
		expected   bool
	}{
		{"Admin can add friends", authorization.RoleAdmin, authorization.PermissionAddFriend, true},
		{"Admin can delete friends", authorization.RoleAdmin, authorization.PermissionDeleteFriend, true},
		{"User can view friends", authorization.RoleUser, authorization.PermissionViewFriend, true},
		{"User cannot delete friends", authorization.RoleUser, authorization.PermissionDeleteFriend, false},
		{"ReadOnly can view friends", authorization.RoleReadOnly, authorization.PermissionViewFriend, true},
		{"ReadOnly cannot add friends", authorization.RoleReadOnly, authorization.PermissionAddFriend, false},
		{"ReadOnly cannot delete friends", authorization.RoleReadOnly, authorization.PermissionDeleteFriend, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.role.HasPermission(tt.permission)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestNewAuthorizationContext(t *testing.T) {
	t.Run("Valid context creation", func(t *testing.T) {
		ctx := authorization.NewAuthorizationContext("user-123", "testuser", []string{"admin", "user"})

		if ctx.UserID != "user-123" {
			t.Errorf("Expected UserID 'user-123', got '%s'", ctx.UserID)
		}

		if ctx.Username != "testuser" {
			t.Errorf("Expected Username 'testuser', got '%s'", ctx.Username)
		}

		if len(ctx.Roles) != 2 {
			t.Errorf("Expected 2 roles, got %d", len(ctx.Roles))
		}

		expectedRoles := []authorization.Role{authorization.RoleAdmin, authorization.RoleUser}
		for i, expectedRole := range expectedRoles {
			if ctx.Roles[i] != expectedRole {
				t.Errorf("Expected role %s at index %d, got %s", expectedRole, i, ctx.Roles[i])
			}
		}
	})

	t.Run("Unknown roles default to readonly", func(t *testing.T) {
		ctx := authorization.NewAuthorizationContext("user-123", "testuser", []string{"unknown", "invalid"})

		expectedRoles := []authorization.Role{authorization.RoleReadOnly, authorization.RoleReadOnly}
		for i, expectedRole := range expectedRoles {
			if ctx.Roles[i] != expectedRole {
				t.Errorf("Expected role %s at index %d, got %s", expectedRole, i, ctx.Roles[i])
			}
		}
	})
}

func TestAuthorizationContext_HasPermission(t *testing.T) {
	tests := []struct {
		name       string
		roles      []string
		permission authorization.Permission
		expected   bool
	}{
		{"Admin context has delete permission", []string{"admin"}, authorization.PermissionDeleteFriend, true},
		{"User context has view permission", []string{"user"}, authorization.PermissionViewFriend, true},
		{"User context doesn't have delete permission", []string{"user"}, authorization.PermissionDeleteFriend, false},
		{"ReadOnly context has view permission", []string{"readonly"}, authorization.PermissionViewFriend, true},
		{"ReadOnly context doesn't have add permission", []string{"readonly"}, authorization.PermissionAddFriend, false},
		{"Multi-role context has permissions from any role", []string{"readonly", "user"}, authorization.PermissionAddFriend, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := authorization.NewAuthorizationContext("user-123", "testuser", tt.roles)
			result := ctx.HasPermission(tt.permission)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestAuthorizationContext_HasRole(t *testing.T) {
	ctx := authorization.NewAuthorizationContext("user-123", "testuser", []string{"admin", "user"})

	if !ctx.HasRole(authorization.RoleAdmin) {
		t.Error("Expected context to have admin role")
	}

	if !ctx.HasRole(authorization.RoleUser) {
		t.Error("Expected context to have user role")
	}

	if ctx.HasRole(authorization.RoleReadOnly) {
		t.Error("Expected context to not have readonly role")
	}
}

func TestAuthorizationContext_IsAdmin(t *testing.T) {
	t.Run("Admin context returns true", func(t *testing.T) {
		ctx := authorization.NewAuthorizationContext("user-123", "testuser", []string{"admin"})
		if !ctx.IsAdmin() {
			t.Error("Expected IsAdmin() to return true for admin context")
		}
	})

	t.Run("Non-admin context returns false", func(t *testing.T) {
		ctx := authorization.NewAuthorizationContext("user-123", "testuser", []string{"user"})
		if ctx.IsAdmin() {
			t.Error("Expected IsAdmin() to return false for non-admin context")
		}
	})
}

func TestAuthorizationContext_Validate(t *testing.T) {
	t.Run("Valid context passes validation", func(t *testing.T) {
		ctx := authorization.NewAuthorizationContext("user-123", "testuser", []string{"user"})
		if err := ctx.Validate(); err != nil {
			t.Errorf("Expected valid context to pass validation, got error: %v", err)
		}
	})

	t.Run("Empty UserID fails validation", func(t *testing.T) {
		ctx := authorization.AuthorizationContext{
			UserID:   "",
			Username: "testuser",
			Roles:    []authorization.Role{authorization.RoleUser},
		}
		if err := ctx.Validate(); err == nil {
			t.Error("Expected empty UserID to fail validation")
		}
	})

	t.Run("Short Username fails validation", func(t *testing.T) {
		ctx := authorization.AuthorizationContext{
			UserID:   "user-123",
			Username: "a",
			Roles:    []authorization.Role{authorization.RoleUser},
		}
		if err := ctx.Validate(); err == nil {
			t.Error("Expected short username to fail validation")
		}
	})

	t.Run("Empty roles fails validation", func(t *testing.T) {
		ctx := authorization.AuthorizationContext{
			UserID:   "user-123",
			Username: "testuser",
			Roles:    []authorization.Role{},
		}
		if err := ctx.Validate(); err == nil {
			t.Error("Expected empty roles to fail validation")
		}
	})
}
