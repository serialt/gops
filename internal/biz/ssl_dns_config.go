package biz

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gops/internal/pkg/cryptox"
)

var dnsProviderFields = map[string][]string{
	"alidns":       {"access_key_id", "access_key_secret", "security_token", "region"},
	"aliyun":       {"access_key_id", "access_key_secret", "security_token", "region"},
	"tencentcloud": {"secret_id", "secret_key", "session_token", "region"},
	"cloudflare":   {"api_token", "zone_token", "email", "api_key"},
	"huaweicloud":  {"access_key_id", "secret_access_key", "region"},
	"huawei":       {"access_key_id", "secret_access_key", "region"},
}

var dnsSecretFields = map[string]struct{}{
	"access_key_secret": {},
	"security_token":    {},
	"secret_key":        {},
	"session_token":     {},
	"api_token":         {},
	"zone_token":        {},
	"api_key":           {},
	"secret_access_key": {},
}

func parseDNSProviderConfig(provider, raw string) (map[string]string, []string, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	fields, ok := dnsProviderFields[provider]
	if !ok {
		return nil, nil, fmt.Errorf("不支持的 DNS 服务商 %q", provider)
	}
	values := make(map[string]string)
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil, nil, errors.New("DNS 配置必须是有效 JSON")
	}
	return values, fields, nil
}

func encryptDNSProviderConfig(aesKey, provider, raw, oldRaw string) (string, error) {
	values, fields, err := parseDNSProviderConfig(provider, raw)
	if err != nil {
		return "", err
	}
	oldValues := make(map[string]string)
	if strings.TrimSpace(oldRaw) != "" {
		if err = json.Unmarshal([]byte(oldRaw), &oldValues); err != nil {
			return "", errors.New("已保存的 DNS 配置无效")
		}
	}
	result := make(map[string]string, len(fields))
	for _, field := range fields {
		value := strings.TrimSpace(values[field])
		if _, secret := dnsSecretFields[field]; secret {
			if value == "" {
				value, err = cryptox.Decrypt(aesKey, oldValues[field])
				if err != nil {
					return "", fmt.Errorf("解密已有 DNS 凭据 %s: %w", field, err)
				}
			}
			value, err = cryptox.Encrypt(aesKey, value)
			if err != nil {
				return "", fmt.Errorf("加密 DNS 凭据 %s: %w", field, err)
			}
		}
		result[field] = value
	}
	encoded, err := json.Marshal(result)
	return string(encoded), err
}

func decryptDNSProviderConfig(aesKey, provider, raw string) (string, error) {
	values, fields, err := parseDNSProviderConfig(provider, raw)
	if err != nil {
		return "", err
	}
	for _, field := range fields {
		if _, secret := dnsSecretFields[field]; !secret {
			continue
		}
		values[field], err = cryptox.Decrypt(aesKey, values[field])
		if err != nil {
			return "", fmt.Errorf("解密 DNS 凭据 %s: %w", field, err)
		}
	}
	encoded, err := json.Marshal(values)
	return string(encoded), err
}

func maskDNSProviderConfig(provider, raw string) (string, error) {
	values, fields, err := parseDNSProviderConfig(provider, raw)
	if err != nil {
		return "", err
	}
	result := make(map[string]string, len(fields))
	for _, field := range fields {
		if _, secret := dnsSecretFields[field]; secret {
			result[field] = ""
		} else {
			result[field] = values[field]
		}
	}
	encoded, err := json.Marshal(result)
	return string(encoded), err
}
