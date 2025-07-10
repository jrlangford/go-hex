package auth

import (
	"context"
)

// ContextKey is used for context values to avoid collisions
type ContextKey string

const (
	ClaimsContextKey ContextKey = "token_claims"
)

// Claims represents authentication claims for a user aggregating all domain claims
type Claims struct {
	UserID   string            `json:"user_id"`
	Username string            `json:"username"`
	Email    string            `json:"email,omitempty"`
	Roles    []string          `json:"roles,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`

	// Domain-specific claims (can override role-based defaults)
	BookingClaims  *BookingClaims  `json:"booking_claims,omitempty"`
	RoutingClaims  *RoutingClaims  `json:"routing_claims,omitempty"`
	HandlingClaims *HandlingClaims `json:"handling_claims,omitempty"`
}

// NewClaims creates a new Claims instance
func NewClaims(userID, username, email string, roles []string, metadata map[string]string) (*Claims, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Roles:    roles,
		Metadata: metadata,
	}

	if claims.Metadata == nil {
		claims.Metadata = make(map[string]string)
	}

	// Initialize domain claims based on roles (can be overridden later)
	claims.initializeDefaultDomainClaims()

	return claims, nil
}

// NewClaimsWithDomainOverrides creates a new Claims instance with specific domain claims that override role-based defaults
func NewClaimsWithDomainOverrides(
	userID, username, email string,
	roles []string,
	metadata map[string]string,
	bookingClaims *BookingClaims,
	routingClaims *RoutingClaims,
	handlingClaims *HandlingClaims,
) (*Claims, error) {
	claims, err := NewClaims(userID, username, email, roles, metadata)
	if err != nil {
		return nil, err
	}

	// Override with specific domain claims if provided
	if bookingClaims != nil {
		claims.BookingClaims = bookingClaims
	}
	if routingClaims != nil {
		claims.RoutingClaims = routingClaims
	}
	if handlingClaims != nil {
		claims.HandlingClaims = handlingClaims
	}

	return claims, nil
}

// initializeDefaultDomainClaims sets up default domain-specific claims based on user roles
func (c *Claims) initializeDefaultDomainClaims() {
	isAdmin := c.HasRole(string(RoleAdmin))
	isUser := c.HasRole(string(RoleUser))
	isReadOnly := c.HasRole(string(RoleReadOnly))

	// Default booking claims based on role
	c.BookingClaims = &BookingClaims{
		CanBookCargo:   isAdmin || isUser,
		CanViewCargo:   isAdmin || isUser || isReadOnly,
		CanTrackCargo:  isAdmin || isUser || isReadOnly,
		CanAssignRoute: isAdmin || isUser,
	}

	// Default routing claims based on role
	c.RoutingClaims = &RoutingClaims{
		CanPlanRoutes:    isAdmin || isUser,
		CanViewVoyages:   isAdmin || isUser || isReadOnly,
		CanViewLocations: isAdmin || isUser || isReadOnly,
	}

	// Default handling claims based on role
	c.HandlingClaims = &HandlingClaims{
		CanSubmitHandling: isAdmin || isUser,
		CanViewHandling:   isAdmin || isUser || isReadOnly,
	}
}

// HasRole checks if the user has a specific role
func (c *Claims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// IsAdmin checks if the user has admin role
func (c *Claims) IsAdmin() bool {
	return c.HasRole(string(RoleAdmin))
}

// IsUser checks if the user has user role
func (c *Claims) IsUser() bool {
	return c.HasRole(string(RoleUser))
}

// IsReadOnly checks if the user has readonly role
func (c *Claims) IsReadOnly() bool {
	return c.HasRole(string(RoleReadOnly))
}

// ExtractClaims extracts authentication claims from request context
func ExtractClaims(ctx context.Context) (*Claims, error) {
	claimsValue := ctx.Value(ClaimsContextKey)
	if claimsValue == nil {
		return nil, NewAuthenticationError("no authentication context found")
	}

	claims, ok := claimsValue.(*Claims)
	if !ok {
		return nil, NewAuthenticationError("invalid authentication context type")
	}

	return claims, nil
}

// Role represents a user role in the system
type Role string

const (
	RoleAdmin    Role = "admin"
	RoleUser     Role = "user"
	RoleReadOnly Role = "readonly"
)

// BookingClaims represents domain-specific claims for the booking context
type BookingClaims struct {
	CanBookCargo   bool `json:"can_book_cargo"`
	CanViewCargo   bool `json:"can_view_cargo"`
	CanTrackCargo  bool `json:"can_track_cargo"`
	CanAssignRoute bool `json:"can_assign_route"`
}

// BookingPermission represents permissions specific to the booking domain
type BookingPermission string

const (
	PermissionBookCargo   BookingPermission = "book_cargo"
	PermissionViewCargo   BookingPermission = "view_cargo"
	PermissionTrackCargo  BookingPermission = "track_cargo"
	PermissionAssignRoute BookingPermission = "assign_route"
)

// HasPermission checks if the booking claims include a specific permission
func (bc *BookingClaims) HasPermission(permission BookingPermission) bool {
	switch permission {
	case PermissionBookCargo:
		return bc.CanBookCargo
	case PermissionViewCargo:
		return bc.CanViewCargo
	case PermissionTrackCargo:
		return bc.CanTrackCargo
	case PermissionAssignRoute:
		return bc.CanAssignRoute
	default:
		return false
	}
}

// RoutingClaims represents domain-specific claims for the routing context
type RoutingClaims struct {
	CanPlanRoutes    bool `json:"can_plan_routes"`
	CanViewVoyages   bool `json:"can_view_voyages"`
	CanViewLocations bool `json:"can_view_locations"`
}

// RoutingPermission represents permissions specific to the routing domain
type RoutingPermission string

const (
	PermissionPlanRoutes    RoutingPermission = "plan_routes"
	PermissionViewVoyages   RoutingPermission = "view_voyages"
	PermissionViewLocations RoutingPermission = "view_locations"
)

// HasPermission checks if the routing claims include a specific permission
func (rc *RoutingClaims) HasPermission(permission RoutingPermission) bool {
	switch permission {
	case PermissionPlanRoutes:
		return rc.CanPlanRoutes
	case PermissionViewVoyages:
		return rc.CanViewVoyages
	case PermissionViewLocations:
		return rc.CanViewLocations
	default:
		return false
	}
}

// HandlingClaims represents domain-specific claims for the handling context
type HandlingClaims struct {
	CanSubmitHandling bool `json:"can_submit_handling"`
	CanViewHandling   bool `json:"can_view_handling"`
}

// HandlingPermission represents permissions specific to the handling domain
type HandlingPermission string

const (
	PermissionSubmitHandling HandlingPermission = "submit_handling"
	PermissionViewHandling   HandlingPermission = "view_handling"
)

// HasPermission checks if the handling claims include a specific permission
func (hc *HandlingClaims) HasPermission(permission HandlingPermission) bool {
	switch permission {
	case PermissionSubmitHandling:
		return hc.CanSubmitHandling
	case PermissionViewHandling:
		return hc.CanViewHandling
	default:
		return false
	}
}

// // DefaultPermissionsPerRole returns default domain permissions for each role
// func DefaultPermissionsPerRole() map[string]DomainPermissions {
// 	return map[string]DomainPermissions{
// 		string(RoleAdmin): {
// 			BookingClaims: &BookingClaims{
// 				CanBookCargo:   true,
// 				CanViewCargo:   true,
// 				CanTrackCargo:  true,
// 				CanAssignRoute: true,
// 			},
// 			RoutingClaims: &RoutingClaims{
// 				CanPlanRoutes:    true,
// 				CanViewVoyages:   true,
// 				CanViewLocations: true,
// 			},
// 			HandlingClaims: &HandlingClaims{
// 				CanSubmitHandling: true,
// 				CanViewHandling:   true,
// 			},
// 		},
// 		string(RoleUser): {
// 			BookingClaims: &BookingClaims{
// 				CanBookCargo:   true,
// 				CanViewCargo:   true,
// 				CanTrackCargo:  true,
// 				CanAssignRoute: true,
// 			},
// 			RoutingClaims: &RoutingClaims{
// 				CanPlanRoutes:    true,
// 				CanViewVoyages:   true,
// 				CanViewLocations: true,
// 			},
// 			HandlingClaims: &HandlingClaims{
// 				CanSubmitHandling: true,
// 				CanViewHandling:   true,
// 			},
// 		},
// 		string(RoleReadOnly): {
// 			BookingClaims: &BookingClaims{
// 				CanBookCargo:   false,
// 				CanViewCargo:   true,
// 				CanTrackCargo:  true,
// 				CanAssignRoute: false,
// 			},
// 			RoutingClaims: &RoutingClaims{
// 				CanPlanRoutes:    false,
// 				CanViewVoyages:   true,
// 				CanViewLocations: true,
// 			},
// 			HandlingClaims: &HandlingClaims{
// 				CanSubmitHandling: false,
// 				CanViewHandling:   true,
// 			},
// 		},
// 	}
// }

// // DomainPermissions aggregates all domain-specific claims
// type DomainPermissions struct {
// 	BookingClaims  *BookingClaims  `json:"booking_claims,omitempty"`
// 	RoutingClaims  *RoutingClaims  `json:"routing_claims,omitempty"`
// 	HandlingClaims *HandlingClaims `json:"handling_claims,omitempty"`
// }
