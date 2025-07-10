package handlingapplication

import (
	"go_hex/internal/support/auth"
)

// RequireHandlingPermission checks if the user has the required handling permission
func RequireHandlingPermission(claims *auth.Claims, permission auth.HandlingPermission) error {
	if claims == nil {
		return auth.NewAuthenticationError("no authentication context found")
	}

	if claims.HandlingClaims == nil {
		return auth.NewAuthorizationError("no handling permissions available")
	}

	if !claims.HandlingClaims.HasPermission(permission) {
		return auth.NewAuthorizationError("insufficient permissions for handling operation")
	}

	return nil
}
