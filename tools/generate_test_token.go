package main

import (
	"fmt"
	"go_hex/internal/support/config"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID   string            `json:"sub"`
	Username string            `json:"username"`
	Email    string            `json:"email"`
	Roles    []string          `json:"roles"`
	Metadata map[string]string `json:"metadata"`
	jwt.RegisteredClaims
}

func main() {
	// Check for predefined token sets
	if len(os.Args) == 2 {
		preset := os.Args[1]
		switch preset {
		case "admin":
			generatePresetToken("admin-001", "admin.user", "admin@cargo-shipping.com", []string{"admin"}, 24)
			return
		case "user":
			generatePresetToken("user-001", "standard.user", "user@cargo-shipping.com", []string{"user"}, 24)
			return
		case "readonly":
			generatePresetToken("readonly-001", "readonly.user", "readonly@cargo-shipping.com", []string{"readonly"}, 24)
			return
		case "super":
			generatePresetToken("super-001", "super.admin", "super@cargo-shipping.com", []string{"admin", "user", "readonly"}, 24)
			return
		}
	}

	if len(os.Args) < 4 {
		fmt.Println("Cargo Shipping System JWT Token Generator")
		fmt.Println("=========================================")
		fmt.Println("")
		fmt.Println("PREDEFINED TOKEN SETS:")
		fmt.Println("  go run generate_test_token.go admin    - Full access (all operations)")
		fmt.Println("  go run generate_test_token.go user     - Standard access (booking, viewing, tracking)")
		fmt.Println("  go run generate_test_token.go readonly - Read-only access (viewing only)")
		fmt.Println("  go run generate_test_token.go super    - All roles combined (for testing)")
		fmt.Println("")
		fmt.Println("CUSTOM TOKEN:")
		fmt.Println("  go run generate_test_token.go <userID> <username> <roles> [email] [expiresInHours]")
		fmt.Println("  Example: go run generate_test_token.go user-123 john.doe admin,user john@example.com 24")
		fmt.Println("")
		fmt.Println("ROLE PERMISSIONS:")
		fmt.Println("  admin    - All operations (booking, routing, handling, administration)")
		fmt.Println("           - Can book cargo, assign routes, submit handling events")
		fmt.Println("  user     - Standard operations (booking, viewing, basic routing)")
		fmt.Println("           - Can book cargo, view details, request routes")
		fmt.Println("  readonly - View-only access (tracking, viewing)")
		fmt.Println("           - Can view cargo, voyages, locations, handling events")
		fmt.Println("")
		fmt.Println("API ENDPOINTS TO TEST:")
		fmt.Println("  GET  /health                                    - System health (no auth)")
		fmt.Println("  GET  /info                                      - System info (no auth)")
		fmt.Println("  POST /api/v1/cargos                            - Book cargo (user, admin)")
		fmt.Println("  GET  /api/v1/cargos                            - List cargo (user, admin, readonly)")
		fmt.Println("  GET  /api/v1/cargos/{trackingId}               - Get cargo details (user, admin, readonly)")
		fmt.Println("  PUT  /api/v1/cargos/{trackingId}/route         - Assign route (user, admin)")
		fmt.Println("  POST /api/v1/route-candidates                  - Find routes (user, admin)")
		fmt.Println("  GET  /api/v1/voyages                           - List voyages (user, admin, readonly)")
		fmt.Println("  GET  /api/v1/locations                         - List locations (user, admin, readonly)")
		fmt.Println("  POST /api/v1/handling-events                   - Submit handling (user, admin)")
		fmt.Println("  GET  /api/v1/handling-events                   - View handling (user, admin, readonly)")
		fmt.Println("")
		os.Exit(1)
	}

	userID := os.Args[1]
	username := os.Args[2]
	rolesStr := os.Args[3]

	email := ""
	if len(os.Args) > 4 {
		email = os.Args[4]
	}

	expiresInHours := 24
	if len(os.Args) > 5 {
		if h, err := strconv.Atoi(os.Args[5]); err == nil {
			expiresInHours = h
		}
	}

	// Parse roles
	roles := []string{}
	if rolesStr != "" {
		roles = splitByComma(rolesStr)
	}

	// Create JWT token
	token := createJWTToken(userID, username, email, roles, expiresInHours)

	fmt.Printf("Generated JWT Token for user '%s':\n", username)
	fmt.Printf("Roles: %v\n", roles)
	fmt.Printf("Expires: %s\n", time.Now().Add(time.Duration(expiresInHours)*time.Hour).Format(time.RFC3339))
	fmt.Printf("\nToken:\n%s\n", token)
	fmt.Printf("\nTest with curl:\n")
	fmt.Printf("curl -H \"Authorization: Bearer %s\" http://localhost:8080/auth/me\n", token)
}

func splitByComma(s string) []string {
	return simpleStringSplit(s, ",")
}

// Simple string split function
func simpleStringSplit(s, sep string) []string {
	if s == "" {
		return []string{}
	}

	result := []string{}
	start := 0

	for i := 0; i <= len(s)-len(sep); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			if i > start {
				result = append(result, s[start:i])
			}
			start = i + len(sep)
			i += len(sep) - 1
		}
	}

	if start < len(s) {
		result = append(result, s[start:])
	}

	if len(result) == 0 {
		result = append(result, s)
	}

	return result
}

func createJWTToken(userID, username, email string, roles []string, expiresInHours int) string {
	// Load configuration (including JWT settings) from environment
	cfg, err := config.New()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Use the same secret key, issuer, and audience as the server
	secretKey := cfg.JWT.SecretKey
	issuer := cfg.JWT.Issuer
	audience := cfg.JWT.Audience

	// Create the claims
	now := time.Now()
	expiresAt := now.Add(time.Duration(expiresInHours) * time.Hour)

	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Roles:    roles,
		Metadata: map[string]string{"generated": "true"},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Audience:  []string{audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret key
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		fmt.Printf("Error creating token: %v\n", err)
		os.Exit(1)
	}

	return tokenString
}

func generatePresetToken(userID, username, email string, roles []string, expiresInHours int) {
	token := createJWTToken(userID, username, email, roles, expiresInHours)

	fmt.Printf("Generated %s Token:\n", roles[0])
	fmt.Printf("User: %s (%s)\n", username, email)
	fmt.Printf("Roles: %v\n", roles)
	fmt.Printf("Expires: %s\n", time.Now().Add(time.Duration(expiresInHours)*time.Hour).Format(time.RFC3339))
	fmt.Printf("\nToken:\n%s\n", token)
	fmt.Printf("\nTest with curl:\n")
	fmt.Printf("curl -H \"Authorization: Bearer %s\" http://localhost:8080/auth/me\n", token)
}
