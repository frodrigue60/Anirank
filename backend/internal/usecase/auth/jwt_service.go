package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService handles creation and verification of tokens
type JWTService struct {
	secretKey []byte
	issuer    string
}

// Claims matches the Laravel Sanctum / Auth payload plus our custom RBAC fields
type Claims struct {
	UserUUID string   `json:"user_uuid"`
	Roles    []string `json:"roles"` // All role slugs assigned to the user
	jwt.RegisteredClaims
}

func NewJWTService() *JWTService {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "fallback_insecure_secret" // Better fallback for dev
	}

	return &JWTService{
		secretKey: []byte(secret),
		issuer:    "anirank",
	}
}

// GenerateToken creates a new token with 24 hours validity
func (s *JWTService) GenerateToken(userUUID string, roles []string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		UserUUID: userUUID,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    s.issuer,
			Subject:   "user_auth",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ValidateToken parses and validates a raw JWT token string
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token format")
}

// GenerateTempToken creates a short-lived token for multi-step flows (like registration)
func (s *JWTService) GenerateTempToken(data map[string]interface{}, duration time.Duration) (string, error) {
	expirationTime := time.Now().Add(duration)

	claims := jwt.MapClaims{
		"exp": jwt.NewNumericDate(expirationTime),
		"iat": jwt.NewNumericDate(time.Now()),
		"iss": s.issuer,
	}

	for k, v := range data {
		claims[k] = v
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ValidateTempToken parses a MapClaims token (generic payload)
func (s *JWTService) ValidateTempToken(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid temporary token")
}
