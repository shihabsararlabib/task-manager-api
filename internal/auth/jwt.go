package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secretKey  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

type Claims struct {
	UserID    int    `json:"uid"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TokenType string `json:"typ"`
	jwt.RegisteredClaims
}

func NewJWTManager(secret string, accessTTL, refreshTTL time.Duration) *JWTManager {
	return &JWTManager{secretKey: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func (m *JWTManager) GenerateAccessToken(userID int, email, role string) (string, error) {
	return m.generate(userID, email, role, "access", m.accessTTL)
}

func (m *JWTManager) GenerateRefreshToken(userID int, email, role string) (string, error) {
	return m.generate(userID, email, role, "refresh", m.refreshTTL)
}

func (m *JWTManager) GeneratePair(userID int, email, role string) (string, string, error) {
	access, err := m.GenerateAccessToken(userID, email, role)
	if err != nil {
		return "", "", err
	}
	refresh, err := m.GenerateRefreshToken(userID, email, role)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func (m *JWTManager) generate(userID int, email, role, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func (m *JWTManager) Parse(tokenString string) (Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return m.secretKey, nil
	})
	if err != nil {
		return Claims{}, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return Claims{}, fmt.Errorf("invalid token claims")
	}
	return *claims, nil
}
