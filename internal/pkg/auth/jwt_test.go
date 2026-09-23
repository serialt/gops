package auth

import (
	"testing"
	"time"

	"gops/internal/config"
)

func TestJWTManagerTokenPairTypes(t *testing.T) {
	manager := NewJWTManager(&config.Config{Server: config.Server{JwtSecret: "test-secret", AccessTokenTimeout: time.Minute, RefreshTokenTimeout: time.Hour}})
	pair, err := manager.GeneratePair(42)
	if err != nil {
		t.Fatal(err)
	}
	access, err := manager.ParseAccess(pair.AccessToken)
	if err != nil || access.UserID != 42 || access.TokenType != "access" {
		t.Fatalf("invalid access token: claims=%+v err=%v", access, err)
	}
	refresh, err := manager.ParseRefresh(pair.RefreshToken)
	if err != nil || refresh.UserID != 42 || refresh.TokenType != "refresh" {
		t.Fatalf("invalid refresh token: claims=%+v err=%v", refresh, err)
	}
	if _, err := manager.ParseAccess(pair.RefreshToken); err == nil {
		t.Fatal("refresh token must not be accepted as access token")
	}
	if _, err := manager.ParseRefresh(pair.AccessToken); err == nil {
		t.Fatal("access token must not be accepted as refresh token")
	}
	if pair.ExpiresIn != 60 || pair.RefreshExpiresIn != 3600 {
		t.Fatalf("unexpected expiry values: %+v", pair)
	}
}
