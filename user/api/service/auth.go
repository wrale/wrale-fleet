package service

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "strings"
    "sync"
    "time"
)

// Role defines user permissions
type Role string

const (
    RoleAdmin  Role = "admin"
    RoleUser   Role = "user"
    RoleViewer Role = "viewer"
)

// Token claims structure
type tokenClaims struct {
    UserID    string    `json:"uid"`
    Roles     []Role    `json:"roles"`
    IssuedAt  time.Time `json:"iat"`
    ExpiresAt time.Time `json:"exp"`
}

// AuthService implements authentication and authorization
type AuthService struct {
    secretKey []byte
    tokens    map[string]tokenClaims
    mu        sync.RWMutex
}

// NewAuthService creates a new auth service
func NewAuthService(secretKey string) *AuthService {
    return &AuthService{
        secretKey: []byte(secretKey),
        tokens:    make(map[string]tokenClaims),
    }
}

// GenerateToken creates a new authentication token
func (s *AuthService) GenerateToken(userID string, roles []string) (string, error) {
    // Convert string roles to Role type
    typedRoles := make([]Role, len(roles))
    for i, r := range roles {
        typedRoles[i] = Role(r)
    }

    // Create claims
    claims := tokenClaims{
        UserID:    userID,
        Roles:     typedRoles,
        IssuedAt:  time.Now(),
        ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hour expiry for v1.0
    }

    // Marshal claims
    claimsJSON, err := json.Marshal(claims)
    if err != nil {
        return "", fmt.Errorf("failed to marshal claims: %w", err)
    }

    // Encode claims
    claimsB64 := base64.StdEncoding.EncodeToString(claimsJSON)

    // Create signature
    h := hmac.New(sha256.New, s.secretKey)
    h.Write([]byte(claimsB64))
    signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

    // Create token
    token := fmt.Sprintf("%s.%s", claimsB64, signature)

    // Store token
    s.mu.Lock()
    s.tokens[token] = claims
    s.mu.Unlock()

    return token, nil
}

// Authenticate verifies token validity
func (s *AuthService) Authenticate(token string) (bool, error) {
    // Check if token exists
    s.mu.RLock()
    claims, exists := s.tokens[token]
    s.mu.RUnlock()

    if !exists {
        // Try to validate and parse token
        parsedClaims, err := s.validateToken(token)
        if err != nil {
            return false, nil // Invalid token
        }
        claims = parsedClaims
    }

    // Check expiration
    if time.Now().After(claims.ExpiresAt) {
        s.mu.Lock()
        delete(s.tokens, token)
        s.mu.Unlock()
        return false, nil
    }

    return true, nil
}

// Authorize checks if token has required permissions
func (s *AuthService) Authorize(token string, resource string, action string) (bool, error) {
    // Get claims
    s.mu.RLock()
    claims, exists := s.tokens[token]
    s.mu.RUnlock()

    if !exists {
        claims, err := s.validateToken(token)
        if err != nil {
            return false, nil
        }
        s.mu.Lock()
        s.tokens[token] = claims
        s.mu.Unlock()
    }

    // Check roles
    for _, role := range claims.Roles {
        if s.hasPermission(role, resource, action) {
            return true, nil
        }
    }

    return false, nil
}

// validateToken validates and parses a token
func (s *AuthService) validateToken(token string) (tokenClaims, error) {
    // Split token into claims and signature
    parts := strings.Split(token, ".")
    if len(parts) != 2 {
        return tokenClaims{}, fmt.Errorf("invalid token format")
    }

    claimsB64, signatureB64 := parts[0], parts[1]

    // Verify signature
    h := hmac.New(sha256.New, s.secretKey)
    h.Write([]byte(claimsB64))
    expectedSignature := base64.StdEncoding.EncodeToString(h.Sum(nil))
    if signatureB64 != expectedSignature {
        return tokenClaims{}, fmt.Errorf("invalid signature")
    }

    // Decode claims
    claimsJSON, err := base64.StdEncoding.DecodeString(claimsB64)
    if err != nil {
        return tokenClaims{}, fmt.Errorf("invalid claims encoding")
    }

    // Parse claims
    var claims tokenClaims
    if err := json.Unmarshal(claimsJSON, &claims); err != nil {
        return tokenClaims{}, fmt.Errorf("invalid claims format")
    }

    return claims, nil
}

// hasPermission checks if a role has permission for an action
func (s *AuthService) hasPermission(role Role, resource string, action string) bool {
    // For v1.0, implement simple role-based permissions
    switch role {
    case RoleAdmin:
        return true // Admins can do everything
    case RoleUser:
        // Users can do most things except sensitive operations
        return !strings.Contains(action, "DELETE") &&
            !strings.Contains(resource, "config")
    case RoleViewer:
        // Viewers can only read
        return action == "GET"
    default:
        return false
    }
}

// Utility functions

// CleanupTokens removes expired tokens
func (s *AuthService) CleanupTokens() {
    s.mu.Lock()
    defer s.mu.Unlock()

    now := time.Now()
    for token, claims := range s.tokens {
        if now.After(claims.ExpiresAt) {
            delete(s.tokens, token)
        }
    }
}
