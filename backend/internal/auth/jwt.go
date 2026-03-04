package auth

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID string `json:"sub"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken    string
	RefreshToken   string
	RefreshTokenID string
	AccessExpiry   time.Time
	RefreshExpiry  time.Time
}

var (
	jwtSecret        []byte
	jwtRefreshSecret []byte
	secretsLoaded    bool
)

func loadSecrets() {
	if secretsLoaded {
		return
	}

	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
	jwtRefreshSecret = []byte(os.Getenv("JWT_REFRESH_SECRET"))

	if len(jwtSecret) == 0 {
		log.Fatal("JWT_SECRET environment variable is required")
	}
	if len(jwtRefreshSecret) == 0 {
		log.Fatal("JWT_REFRESH_SECRET environment variable is required")
	}

	secretsLoaded = true
}

func GenerateTokenPair(userID, email, role string) (*TokenPair, error) {
	loadSecrets()
	now := time.Now()

	// Access token - 15 minutes
	accessExpiry := now.Add(15 * time.Minute)
	accessClaims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   userID,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh token - opaque UUID token (not JWT)
	// Using UUID allows bcrypt hashing without 72-byte limit
	refreshExpiry := now.Add(7 * 24 * time.Hour)
	refreshTokenID := uuid.New().String()
	refreshTokenString := uuid.New().String() + uuid.New().String() // 72 chars total

	return &TokenPair{
		AccessToken:    accessTokenString,
		RefreshToken:   refreshTokenString,
		RefreshTokenID: refreshTokenID,
		AccessExpiry:   accessExpiry,
		RefreshExpiry:  refreshExpiry,
	}, nil
}

func ValidateAccessToken(tokenString string) (*Claims, error) {
	loadSecrets()
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

func GetTokenID(tokenString string) (string, error) {
	loadSecrets()
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtRefreshSecret, nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok {
		return claims.ID, nil
	}

	return "", fmt.Errorf("invalid token claims")
}
