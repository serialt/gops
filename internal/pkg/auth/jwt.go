package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"gops/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID    uint   `json:"user_id"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret        []byte
	accessExpire  time.Duration
	refreshExpire time.Duration
}

func NewJWTManager(cfg *config.Config) *JWTManager {
	return &JWTManager{secret: []byte(cfg.Server.JwtSecret), accessExpire: cfg.Server.AccessTokenTimeout, refreshExpire: cfg.Server.RefreshTokenTimeout}
}

type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	ExpiresIn        int64
	RefreshExpiresIn int64
}

func (m *JWTManager) GeneratePair(userID uint) (*TokenPair, error) {
	access, err := m.generate(userID, "access", m.accessExpire)
	if err != nil {
		return nil, err
	}
	refresh, err := m.generate(userID, "refresh", m.refreshExpire)
	if err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: access, RefreshToken: refresh, ExpiresIn: int64(m.accessExpire.Seconds()), RefreshExpiresIn: int64(m.refreshExpire.Seconds())}, nil
}

func (m *JWTManager) generate(userID uint, tokenType string, expire time.Duration) (string, error) {
	now := time.Now()
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return "", err
	}
	claims := Claims{
		UserID: userID, TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expire)),
			ID:        hex.EncodeToString(idBytes),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *JWTManager) ParseAccess(tokenString string) (*Claims, error) {
	return m.parse(tokenString, "access")
}
func (m *JWTManager) ParseRefresh(tokenString string) (*Claims, error) {
	return m.parse(tokenString, "refresh")
}

func (m *JWTManager) parse(tokenString, expectedType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	// Existing access tokens issued before the dual-token rollout did not contain
	// token_type. Accept them only as access tokens until their original expiry.
	legacyAccess := expectedType == "access" && claims.TokenType == ""
	if claims.TokenType != expectedType && !legacyAccess {
		return nil, errors.New("invalid token type")
	}
	return claims, nil
}
