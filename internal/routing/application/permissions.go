package application

import (
	"go_hex/internal/support/auth"
)

// RequireRoutingPermission checks if the user has the required routing permission
func RequireRoutingPermission(claims *auth.Claims, permission auth.RoutingPermission) error {
	if claims == nil {
		return auth.NewAuthenticationError("no authentication context found")
	}

	if claims.RoutingClaims == nil {
		return auth.NewAuthorizationError("no routing permissions available")
	}

	if !claims.RoutingClaims.HasPermission(permission) {
		return auth.NewAuthorizationError("insufficient permissions for routing operation")
	}

	return nil
}
