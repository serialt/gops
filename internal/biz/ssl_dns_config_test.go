package biz

import (
	"encoding/json"
	"strings"
	"testing"
)

const testDNSAESKey = "0123456789abcdef0123456789abcdef"

func TestDNSProviderConfigEncryptDecryptAndMask(t *testing.T) {
	raw := `{"access_key_id":"ak-id","access_key_secret":"secret-value","security_token":"token-value","region":"cn-hangzhou"}`
	encrypted, err := encryptDNSProviderConfig(testDNSAESKey, "alidns", raw, "")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(encrypted, "secret-value") || strings.Contains(encrypted, "token-value") || !strings.Contains(encrypted, "aesgcm:") {
		t.Fatalf("sensitive DNS values were not encrypted: %s", encrypted)
	}

	plain, err := decryptDNSProviderConfig(testDNSAESKey, "alidns", encrypted)
	if err != nil {
		t.Fatal(err)
	}
	var plainValues map[string]string
	if err = json.Unmarshal([]byte(plain), &plainValues); err != nil {
		t.Fatal(err)
	}
	if plainValues["access_key_secret"] != "secret-value" || plainValues["security_token"] != "token-value" {
		t.Fatalf("decrypted DNS values do not match: %#v", plainValues)
	}

	masked, err := maskDNSProviderConfig("alidns", encrypted)
	if err != nil {
		t.Fatal(err)
	}
	var maskedValues map[string]string
	if err = json.Unmarshal([]byte(masked), &maskedValues); err != nil {
		t.Fatal(err)
	}
	if maskedValues["access_key_id"] != "ak-id" || maskedValues["access_key_secret"] != "" || maskedValues["security_token"] != "" {
		t.Fatalf("DNS response was not masked correctly: %#v", maskedValues)
	}
}

func TestDNSProviderConfigPreservesBlankSecretOnUpdate(t *testing.T) {
	oldRaw := `{"secret_id":"secret-id","secret_key":"old-secret","session_token":"","region":"ap-guangzhou"}`
	oldEncrypted, err := encryptDNSProviderConfig(testDNSAESKey, "tencentcloud", oldRaw, "")
	if err != nil {
		t.Fatal(err)
	}
	updatedRaw := `{"secret_id":"new-id","secret_key":"","session_token":"","region":"ap-shanghai"}`
	updated, err := encryptDNSProviderConfig(testDNSAESKey, "tencentcloud", updatedRaw, oldEncrypted)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := decryptDNSProviderConfig(testDNSAESKey, "tencentcloud", updated)
	if err != nil {
		t.Fatal(err)
	}
	var values map[string]string
	if err = json.Unmarshal([]byte(plain), &values); err != nil {
		t.Fatal(err)
	}
	if values["secret_id"] != "new-id" || values["secret_key"] != "old-secret" || values["region"] != "ap-shanghai" {
		t.Fatalf("blank secret was not preserved: %#v", values)
	}
}
