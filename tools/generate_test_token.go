package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents JWT claims structure that implements jwt.Claims interface.
type JWTClaims struct {
	UserID   string            `json:"sub"`
	Username string            `json:"username"`
	Email    string            `json:"email"`
	Roles    []string          `json:"roles"`
	Metadata map[string]string `json:"metadata"`
	jwt.RegisteredClaims
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run generate_test_token.go <userID> <username> <roles> [email] [expiresInHours]")
		fmt.Println("Example: go run generate_test_token.go user-123 john.doe user,admin john@example.com 24")
		fmt.Println("")
		fmt.Println("Common role combinations:")
		fmt.Println("  admin,user   - Full access (admin can do everything)")
		fmt.Println("  user         - Standard user (can add, view, update, greet)")
		fmt.Println("  readonly     - Read-only access (can only view and greet)")
		fmt.Println("")
		fmt.Println("Test endpoints:")
		fmt.Println("  GET  /auth/me                 - View current user info")
		fmt.Println("  GET  /friends                 - List all friends (any auth)")
		fmt.Println("  POST /friends                 - Add friend (admin only)")
		fmt.Println("  GET  /friends/{id}            - Get specific friend (any auth)")
		fmt.Println("  PUT  /friends/{id}            - Update friend (any auth)")
		fmt.Println("  DELETE /friends/{id}          - Delete friend (admin only)")
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
	// Secret key used for signing (must match the one used in the application)
	secretKey := "your-secret-key-7890123456789012" // Must match main.go

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
			Issuer:    "go-hex-service",
			Audience:  []string{"go-hex-api"},
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
