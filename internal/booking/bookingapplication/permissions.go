package bookingapplication

import (
	"go_hex/internal/support/auth"
)

// RequireBookingPermission checks if the user has the required booking permission
func RequireBookingPermission(claims *auth.Claims, permission auth.BookingPermission) error {
	if claims == nil {
		return auth.NewAuthenticationError("no authentication context found")
	}

	if claims.BookingClaims == nil {
		return auth.NewAuthorizationError("no booking permissions available")
	}

	if !claims.BookingClaims.HasPermission(permission) {
		return auth.NewAuthorizationError("insufficient permissions for booking operation")
	}

	return nil
}
