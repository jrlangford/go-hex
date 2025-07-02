package middleware

import (
	"context"
	"go_hex/internal/support/auth"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ContextKey is a type for context keys to avoid collisions.
type ContextKey string

const (
	TokenClaimsKey ContextKey = "token_claims"
)

// JWTClaims represents JWT claims structure that implements jwt.Claims interface.
type JWTClaims struct {
	UserID   string            `json:"sub" validate:"required,min=1"`
	Username string            `json:"username" validate:"required,min=2,max=50"`
	Email    string            `json:"email" validate:"omitempty,email"`
	Roles    []string          `json:"roles" validate:"omitempty,dive,role"`
	Metadata map[string]string `json:"metadata" validate:"omitempty"`
	jwt.RegisteredClaims
}

// AuthMiddleware provides authentication middleware functionality with JWT validation.
type AuthMiddleware struct {
	secretKey string
	issuer    string
	audience  string
}

func NewAuthMiddleware(secretKey, issuer, audience string) *AuthMiddleware {
	if secretKey == "" {
		panic("secretKey cannot be empty")
	}
	if len(secretKey) < 32 {
		panic("secretKey must be at least 32 characters for HS256")
	}
	if issuer == "" {
		panic("issuer cannot be empty")
	}
	if audience == "" {
		panic("audience cannot be empty")
	}

	return &AuthMiddleware{
		secretKey: secretKey,
		issuer:    issuer,
		audience:  audience,
	}
}

func (m *AuthMiddleware) validateJWTToken(ctx context.Context, tokenString string) (*auth.Claims, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	tokenString = strings.TrimSpace(tokenString)

	if tokenString == "" {
		return nil, ErrMissingToken
	}

	if err := validateTokenLength(tokenString); err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(m.secretKey), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.Issuer != m.issuer {
		return nil, ErrInvalidToken
	}

	if len(claims.Audience) == 0 || !contains(claims.Audience, m.audience) {
		return nil, ErrInvalidToken
	}

	now := time.Now()
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(now) {
		return nil, ErrInvalidToken
	}

	if claims.IssuedAt != nil && claims.IssuedAt.Time.After(now.Add(time.Minute)) {
		return nil, ErrInvalidToken
	}

	tokenClaims, err := auth.NewClaims(
		claims.UserID,
		claims.Username,
		claims.Email,
		claims.Roles,
		claims.Metadata,
	)
	if err != nil {
		return nil, ErrInvalidToken
	}

	return tokenClaims, nil
}

func GetTokenClaims(ctx context.Context) *auth.Claims {
	claims, ok := ctx.Value(TokenClaimsKey).(*auth.Claims)
	if !ok {
		return nil
	}
	return claims
}

func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			writeErrorResponse(w, "Authentication token required", http.StatusUnauthorized)
			return
		}

		claims, err := m.validateJWTToken(r.Context(), token)
		if err != nil {
			switch err {
			case ErrInvalidToken:
				writeErrorResponse(w, "Invalid or expired token", http.StatusUnauthorized)
			case ErrMissingToken:
				writeErrorResponse(w, "Authentication token required", http.StatusUnauthorized)
			default:
				writeErrorResponse(w, "Authentication failed: "+err.Error(), http.StatusUnauthorized)
			}
			return
		}

		ctx := context.WithValue(r.Context(), TokenClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (m *AuthMiddleware) OptionalAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		claims, err := m.validateJWTToken(r.Context(), token)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), TokenClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func HasRole(ctx context.Context, roles ...string) bool {
	claims := GetTokenClaims(ctx)
	if claims == nil {
		return false
	}

	for _, userRole := range claims.Roles {
		for _, requiredRole := range roles {
			if userRole == requiredRole {
				return true
			}
		}
	}
	return false
}

func (m *AuthMiddleware) RequireRole(roles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return m.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(TokenClaimsKey).(*auth.Claims)
			if !ok || claims == nil {
				writeErrorResponse(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			hasRole := false
			for _, userRole := range claims.Roles {
				for _, requiredRole := range roles {
					if userRole == requiredRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				writeErrorResponse(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *AuthMiddleware) RequireAllRoles(roles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return m.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(TokenClaimsKey).(*auth.Claims)
			if !ok || claims == nil {
				writeErrorResponse(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			for _, requiredRole := range roles {
				hasRole := false
				for _, userRole := range claims.Roles {
					if userRole == requiredRole {
						hasRole = true
						break
					}
				}
				if !hasRole {
					writeErrorResponse(w, "Insufficient permissions", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *AuthMiddleware) RequireAnyRole(roles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return m.RequireRole(roles...)
}

func (m *AuthMiddleware) RequireOwnershipOrRole(ownerIDExtractor func(*http.Request) string, roles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return m.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(TokenClaimsKey).(*auth.Claims)
			if !ok || claims == nil {
				writeErrorResponse(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			resourceUserID := ownerIDExtractor(r)
			if resourceUserID == claims.UserID {
				next.ServeHTTP(w, r)
				return
			}

			hasRole := false
			for _, userRole := range claims.Roles {
				for _, requiredRole := range roles {
					if userRole == requiredRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				writeErrorResponse(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return r.URL.Query().Get("token")
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(`{"error":"` + message + `"}`))
}

const (
	MaxTokenLength             = 4096
	SecurityHeaderCacheControl = "no-cache, no-store, must-revalidate"
	SecurityHeaderPragma       = "no-cache"
)

func validateTokenLength(token string) error {
	if len(token) > MaxTokenLength {
		return ErrInvalidToken
	}
	return nil
}
