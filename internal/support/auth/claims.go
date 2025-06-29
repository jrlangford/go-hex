package auth

import (
	"go_hex/internal/core/domain/authorization"
	"go_hex/internal/core/domain/shared"
	"go_hex/internal/support/validation"
)

// Claims represents the authentication claims extracted from a valid authentication token.
// This is a shared authentication model used across the application infrastructure.
type Claims struct {
	UserID   string            `json:"user_id" validate:"required,min=1"`
	Username string            `json:"username" validate:"required,min=2,max=50"`
	Email    string            `json:"email,omitempty" validate:"omitempty,email"`
	Roles    []string          `json:"roles,omitempty" validate:"omitempty,dive,role"`
	Metadata map[string]string `json:"metadata,omitempty" validate:"omitempty"`
}

// NewClaims creates a new Claims instance with validation.
func NewClaims(userID, username, email string, roles []string, metadata map[string]string) (*Claims, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Roles:    roles,
		Metadata: metadata,
	}

	if err := validation.Validate(claims); err != nil {
		return nil, shared.NewDomainValidationError("invalid authentication claims", err)
	}

	if claims.Metadata == nil {
		claims.Metadata = make(map[string]string)
	}

	return claims, nil
}

// IsAuthorized checks if the claims have any of the required roles.
func (c *Claims) IsAuthorized(requiredRoles ...string) bool {
	if len(requiredRoles) == 0 {
		return true // No specific roles required
	}

	roleMap := make(map[string]bool)
	for _, role := range c.Roles {
		roleMap[role] = true
	}

	for _, required := range requiredRoles {
		if roleMap[required] {
			return true
		}
	}

	return false
}

// ToAuthorizationContext converts authentication claims to domain authorization context.
func (c *Claims) ToAuthorizationContext() authorization.AuthorizationContext {
	return authorization.NewAuthorizationContext(c.UserID, c.Username, c.Roles)
}
