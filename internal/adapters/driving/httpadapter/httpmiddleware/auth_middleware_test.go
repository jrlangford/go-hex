package httpmiddleware

import (
	"context"
	"go_hex/internal/support/auth"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// createTestJWTToken creates a test JWT token for testing purposes using proper JWT signing
func createTestJWTToken(userID, username string, roles []string, issuer, audience string, expiresAt int64, secretKey string) string {
	// Create claims
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		Metadata: map[string]string{"test": "data"},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Audience:  []string{audience},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Unix(expiresAt, 0)),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret key
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		panic("Failed to create test token: " + err.Error())
	}

	return tokenString
}

func TestAuthMiddleware_RequireAuth(t *testing.T) {
	authMiddleware := NewAuthMiddleware("test-secret-that-is-at-least-32-characters-long", "go-hex-service", "go-hex-api")

	// Test handler that we'll protect
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}

	t.Run("Valid token", func(t *testing.T) {
		// Create a valid test token
		expiresAt := time.Now().Add(time.Hour).Unix()
		validToken := createTestJWTToken("user-123", "testuser", []string{"user"}, "go-hex-service", "go-hex-api", expiresAt, "test-secret-that-is-at-least-32-characters-long")

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rr := httptest.NewRecorder()

		protectedHandler := authMiddleware.RequireAuth(testHandler)
		protectedHandler(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %d", rr.Code)
		}
	})

	t.Run("Missing token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		protectedHandler := authMiddleware.RequireAuth(testHandler)
		protectedHandler(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status Unauthorized, got %d", rr.Code)
		}
	})

	t.Run("Invalid token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rr := httptest.NewRecorder()

		protectedHandler := authMiddleware.RequireAuth(testHandler)
		protectedHandler(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status Unauthorized, got %d", rr.Code)
		}
	})

	t.Run("Expired token", func(t *testing.T) {
		// Create an expired token
		expiredTime := time.Now().Add(-time.Hour).Unix()
		expiredToken := createTestJWTToken("user-123", "testuser", []string{"user"}, "go-hex-service", "go-hex-api", expiredTime, "test-secret-that-is-at-least-32-characters-long")

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+expiredToken)
		rr := httptest.NewRecorder()

		protectedHandler := authMiddleware.RequireAuth(testHandler)
		protectedHandler(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status Unauthorized for expired token, got %d", rr.Code)
		}
	})

	t.Run("Wrong issuer", func(t *testing.T) {
		// Create a token with wrong issuer
		expiresAt := time.Now().Add(time.Hour).Unix()
		wrongIssuerToken := createTestJWTToken("user-123", "testuser", []string{"user"}, "wrong-issuer", "go-hex-api", expiresAt, "test-secret-that-is-at-least-32-characters-long")

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+wrongIssuerToken)
		rr := httptest.NewRecorder()

		protectedHandler := authMiddleware.RequireAuth(testHandler)
		protectedHandler(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status Unauthorized for wrong issuer, got %d", rr.Code)
		}
	})

	t.Run("Wrong audience", func(t *testing.T) {
		// Create a token with wrong audience
		expiresAt := time.Now().Add(time.Hour).Unix()
		wrongAudienceToken := createTestJWTToken("user-123", "testuser", []string{"user"}, "go-hex-service", "wrong-audience", expiresAt, "test-secret-that-is-at-least-32-characters-long")

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+wrongAudienceToken)
		rr := httptest.NewRecorder()

		protectedHandler := authMiddleware.RequireAuth(testHandler)
		protectedHandler(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status Unauthorized for wrong audience, got %d", rr.Code)
		}
	})

	t.Run("Malformed Authorization header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Basic somevalue")
		rr := httptest.NewRecorder()

		protectedHandler := authMiddleware.RequireAuth(testHandler)
		protectedHandler(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status Unauthorized for malformed auth header, got %d", rr.Code)
		}
	})

	t.Run("Empty Bearer token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer ")
		rr := httptest.NewRecorder()

		protectedHandler := authMiddleware.RequireAuth(testHandler)
		protectedHandler(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status Unauthorized for empty bearer token, got %d", rr.Code)
		}
	})
}

func TestAuthMiddleware_RequireRole(t *testing.T) {
	authMiddleware := NewAuthMiddleware("test-secret-that-is-at-least-32-characters-long", "go-hex-service", "go-hex-api")

	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}

	t.Run("User with admin role", func(t *testing.T) {
		// Create a token with admin role
		expiresAt := time.Now().Add(time.Hour).Unix()
		adminToken := createTestJWTToken("admin-123", "admin", []string{"admin", "user"}, "go-hex-service", "go-hex-api", expiresAt, "test-secret-that-is-at-least-32-characters-long")

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rr := httptest.NewRecorder()

		protectedHandler := authMiddleware.RequireRole("admin")(testHandler)
		protectedHandler(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %d", rr.Code)
		}
	})

	t.Run("User without required role", func(t *testing.T) {
		// Create a token with only user role
		expiresAt := time.Now().Add(time.Hour).Unix()
		userToken := createTestJWTToken("user-123", "user", []string{"user"}, "go-hex-service", "go-hex-api", expiresAt, "test-secret-that-is-at-least-32-characters-long")

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		rr := httptest.NewRecorder()

		protectedHandler := authMiddleware.RequireRole("admin")(testHandler)
		protectedHandler(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("Expected status Forbidden, got %d", rr.Code)
		}
	})
}

func TestAuthMiddleware_OptionalAuth(t *testing.T) {
	authMiddleware := NewAuthMiddleware("test-secret-that-is-at-least-32-characters-long", "go-hex-service", "go-hex-api")

	testHandler := func(w http.ResponseWriter, r *http.Request) {
		claims := GetTokenClaims(r.Context())
		if claims != nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("authenticated"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("anonymous"))
		}
	}

	t.Run("With valid token", func(t *testing.T) {
		// Create a valid test token
		expiresAt := time.Now().Add(time.Hour).Unix()
		validToken := createTestJWTToken("user-123", "testuser", []string{"user"}, "go-hex-service", "go-hex-api", expiresAt, "test-secret-that-is-at-least-32-characters-long")

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rr := httptest.NewRecorder()

		optionalHandler := authMiddleware.OptionalAuth(testHandler)
		optionalHandler(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %d", rr.Code)
		}

		if rr.Body.String() != "authenticated" {
			t.Errorf("Expected 'authenticated', got %s", rr.Body.String())
		}
	})

	t.Run("Without token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		optionalHandler := authMiddleware.OptionalAuth(testHandler)
		optionalHandler(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %d", rr.Code)
		}

		if rr.Body.String() != "anonymous" {
			t.Errorf("Expected 'anonymous', got %s", rr.Body.String())
		}
	})
}

func TestGetTokenClaims(t *testing.T) {
	t.Run("With claims in context", func(t *testing.T) {
		claims, err := auth.NewClaims(
			"test-123",
			"testuser",
			"",
			[]string{"user"},
			nil,
		)
		if err != nil {
			t.Fatalf("Failed to create claims: %v", err)
		}

		ctx := context.WithValue(context.Background(), auth.ClaimsContextKey, claims)

		retrievedClaims := GetTokenClaims(ctx)
		if retrievedClaims == nil {
			t.Error("Expected claims, got nil")
			return
		}

		if retrievedClaims.UserID != "test-123" {
			t.Errorf("Expected UserID 'test-123', got %s", retrievedClaims.UserID)
		}
	})

	t.Run("Without claims in context", func(t *testing.T) {
		ctx := context.Background()
		claims := GetTokenClaims(ctx)

		if claims != nil {
			t.Error("Expected nil claims, got claims")
		}
	})
}

func TestHasRole(t *testing.T) {
	claims, err := auth.NewClaims(
		"test-123",
		"testuser",
		"",
		[]string{"user", "admin"}, // Use valid roles
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create claims: %v", err)
	}

	ctx := context.WithValue(context.Background(), auth.ClaimsContextKey, claims)

	t.Run("User has role", func(t *testing.T) {
		if !HasRole(ctx, "user") {
			t.Error("Expected user to have 'user' role")
		}

		if !HasRole(ctx, "admin") {
			t.Error("Expected user to have 'admin' role")
		}

		if !HasRole(ctx, "user", "admin") {
			t.Error("Expected user to have at least one of the roles")
		}
	})

	t.Run("User doesn't have role", func(t *testing.T) {
		if HasRole(ctx, "readonly") {
			t.Error("Expected user not to have 'readonly' role")
		}
	})

	t.Run("No claims in context", func(t *testing.T) {
		emptyCtx := context.Background()
		if HasRole(emptyCtx, "user") {
			t.Error("Expected false when no claims in context")
		}
	})
}

func TestErrorResponseFormat(t *testing.T) {
	authMiddleware := NewAuthMiddleware("test-secret-that-is-at-least-32-characters-long", "go-hex-service", "go-hex-api")

	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}

	t.Run("Error response contains proper JSON", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		protectedHandler := authMiddleware.RequireAuth(testHandler)
		protectedHandler(rr, req)

		if rr.Header().Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", rr.Header().Get("Content-Type"))
		}

		// Check if response is valid JSON with error field
		body := rr.Body.String()
		if !strings.Contains(body, `"error":`) {
			t.Errorf("Expected error field in JSON response, got %s", body)
		}
	})
}

func TestConcurrentAccess(t *testing.T) {
	authMiddleware := NewAuthMiddleware("test-secret-that-is-at-least-32-characters-long", "go-hex-service", "go-hex-api")

	testHandler := func(w http.ResponseWriter, r *http.Request) {
		claims := GetTokenClaims(r.Context())
		if claims != nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("authenticated"))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("unauthenticated"))
		}
	}

	t.Run("Concurrent authentication requests", func(t *testing.T) {
		expiresAt := time.Now().Add(time.Hour).Unix()
		validToken := createTestJWTToken("user-123", "testuser", []string{"user"}, "go-hex-service", "go-hex-api", expiresAt, "test-secret-that-is-at-least-32-characters-long")

		protectedHandler := authMiddleware.RequireAuth(testHandler)

		// Run multiple concurrent requests
		const numRequests = 10
		done := make(chan bool, numRequests)

		for i := 0; i < numRequests; i++ {
			go func() {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Authorization", "Bearer "+validToken)
				rr := httptest.NewRecorder()

				protectedHandler(rr, req)

				if rr.Code != http.StatusOK {
					t.Errorf("Expected status OK, got %d", rr.Code)
				}
				done <- true
			}()
		}

		// Wait for all requests to complete
		for i := 0; i < numRequests; i++ {
			<-done
		}
	})
}

// Benchmark tests to ensure the middleware performs well under load
func BenchmarkAuthMiddleware_RequireAuth(b *testing.B) {
	authMiddleware := NewAuthMiddleware("test-secret-that-is-at-least-32-characters-long", "go-hex-service", "go-hex-api")

	// Create a valid token
	expiresAt := time.Now().Add(time.Hour).Unix()
	validToken := createTestJWTToken("user-123", "testuser", []string{"user"}, "go-hex-service", "go-hex-api", expiresAt, "test-secret-that-is-at-least-32-characters-long")

	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	protectedHandler := authMiddleware.RequireAuth(testHandler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rr := httptest.NewRecorder()
		protectedHandler(rr, req)
	}
}

func BenchmarkAuthMiddleware_OptionalAuth(b *testing.B) {
	authMiddleware := NewAuthMiddleware("test-secret-that-is-at-least-32-characters-long", "go-hex-service", "go-hex-api")

	// Create a valid token
	expiresAt := time.Now().Add(time.Hour).Unix()
	validToken := createTestJWTToken("user-123", "testuser", []string{"user"}, "go-hex-service", "go-hex-api", expiresAt, "test-secret-that-is-at-least-32-characters-long")

	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	optionalHandler := authMiddleware.OptionalAuth(testHandler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rr := httptest.NewRecorder()
		optionalHandler(rr, req)
	}
}

func BenchmarkTokenValidation(b *testing.B) {
	authMiddleware := NewAuthMiddleware("test-secret-that-is-at-least-32-characters-long", "go-hex-service", "go-hex-api")

	// Create a valid token
	expiresAt := time.Now().Add(time.Hour).Unix()
	validToken := createTestJWTToken("user-123", "testuser", []string{"user"}, "go-hex-service", "go-hex-api", expiresAt, "test-secret-that-is-at-least-32-characters-long")

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = authMiddleware.validateJWTToken(ctx, validToken)
	}
}

func TestNewAuthMiddleware_Validation(t *testing.T) {
	t.Run("Valid configuration", func(t *testing.T) {
		middleware := NewAuthMiddleware("test-secret-that-is-at-least-32-characters-long", "issuer", "audience")
		if middleware == nil {
			t.Error("Expected valid middleware, got nil")
		}
	})

	t.Run("Empty secret key panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for empty secret key")
			}
		}()
		NewAuthMiddleware("", "issuer", "audience")
	})

	t.Run("Short secret key panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for short secret key")
			}
		}()
		NewAuthMiddleware("short", "issuer", "audience")
	})

	t.Run("Empty issuer panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for empty issuer")
			}
		}()
		NewAuthMiddleware("test-secret-that-is-at-least-32-characters-long", "", "audience")
	})

	t.Run("Empty audience panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for empty audience")
			}
		}()
		NewAuthMiddleware("test-secret-that-is-at-least-32-characters-long", "issuer", "")
	})
}

func TestAuthMiddleware_RequireAllRoles(t *testing.T) {
	authMiddleware := NewAuthMiddleware("test-secret-that-is-at-least-32-characters-long", "go-hex-service", "go-hex-api")

	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}

	t.Run("User has all required roles", func(t *testing.T) {
		expiresAt := time.Now().Add(time.Hour).Unix()
		token := createTestJWTToken("user-123", "user", []string{"admin", "user"}, "go-hex-service", "go-hex-api", expiresAt, "test-secret-that-is-at-least-32-characters-long")

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		protectedHandler := authMiddleware.RequireAllRoles("admin", "user")(testHandler)
		protectedHandler(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %d", rr.Code)
		}
	})

	t.Run("User missing one required role", func(t *testing.T) {
		expiresAt := time.Now().Add(time.Hour).Unix()
		token := createTestJWTToken("user-123", "user", []string{"admin", "user"}, "go-hex-service", "go-hex-api", expiresAt, "test-secret-that-is-at-least-32-characters-long")

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		protectedHandler := authMiddleware.RequireAllRoles("admin", "editor")(testHandler)
		protectedHandler(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("Expected status Forbidden, got %d", rr.Code)
		}
	})
}

func TestAuthMiddleware_RequireOwnershipOrRole(t *testing.T) {
	authMiddleware := NewAuthMiddleware("test-secret-that-is-at-least-32-characters-long", "go-hex-service", "go-hex-api")

	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}

	// Mock owner ID extractor that gets user ID from URL path
	ownerIDExtractor := func(r *http.Request) string {
		return r.URL.Query().Get("user_id")
	}

	t.Run("User is owner", func(t *testing.T) {
		expiresAt := time.Now().Add(time.Hour).Unix()
		token := createTestJWTToken("user-123", "user", []string{"user"}, "go-hex-service", "go-hex-api", expiresAt, "test-secret-that-is-at-least-32-characters-long")

		req := httptest.NewRequest("GET", "/test?user_id=user-123", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		protectedHandler := authMiddleware.RequireOwnershipOrRole(ownerIDExtractor, "admin")(testHandler)
		protectedHandler(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK for owner, got %d", rr.Code)
		}
	})

	t.Run("User is not owner but has admin role", func(t *testing.T) {
		expiresAt := time.Now().Add(time.Hour).Unix()
		token := createTestJWTToken("admin-456", "admin", []string{"admin", "user"}, "go-hex-service", "go-hex-api", expiresAt, "test-secret-that-is-at-least-32-characters-long")

		req := httptest.NewRequest("GET", "/test?user_id=user-123", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		protectedHandler := authMiddleware.RequireOwnershipOrRole(ownerIDExtractor, "admin")(testHandler)
		protectedHandler(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK for admin, got %d", rr.Code)
		}
	})

	t.Run("User is not owner and lacks admin role", func(t *testing.T) {
		expiresAt := time.Now().Add(time.Hour).Unix()
		token := createTestJWTToken("other-789", "other", []string{"user"}, "go-hex-service", "go-hex-api", expiresAt, "test-secret-that-is-at-least-32-characters-long")

		req := httptest.NewRequest("GET", "/test?user_id=user-123", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		protectedHandler := authMiddleware.RequireOwnershipOrRole(ownerIDExtractor, "admin")(testHandler)
		protectedHandler(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("Expected status Forbidden, got %d", rr.Code)
		}
	})
}
