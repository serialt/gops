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
	if access.SessionID == "" || access.SessionID != refresh.SessionID {
		t.Fatalf("token pair must share one non-empty session id: access=%q refresh=%q", access.SessionID, refresh.SessionID)
	}
	if !manager.IsSessionActive(access) || !manager.IsSessionActive(refresh) {
		t.Fatal("new token pair must belong to the active session")
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

func TestJWTManagerLatestLoginReplacesPreviousSession(t *testing.T) {
	manager := NewJWTManager(&config.Config{Server: config.Server{JwtSecret: "test-secret", AccessTokenTimeout: time.Minute, RefreshTokenTimeout: time.Hour}})
	firstPair, err := manager.GeneratePair(42)
	if err != nil {
		t.Fatal(err)
	}
	otherPair, err := manager.GeneratePair(7)
	if err != nil {
		t.Fatal(err)
	}
	secondPair, err := manager.GeneratePair(42)
	if err != nil {
		t.Fatal(err)
	}

	firstClaims, err := manager.ParseAccess(firstPair.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	secondClaims, err := manager.ParseAccess(secondPair.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	otherClaims, err := manager.ParseAccess(otherPair.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if manager.IsSessionActive(firstClaims) {
		t.Fatal("previous login session must become inactive")
	}
	if !manager.IsSessionActive(secondClaims) {
		t.Fatal("latest login session must remain active")
	}
	if !manager.IsSessionActive(otherClaims) {
		t.Fatal("logging in as one user must not affect another user's session")
	}
}
