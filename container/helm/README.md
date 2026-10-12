# Gops Helm Chart

```bash
helm upgrade --install gops ./container/helm \
  --namespace gops --create-namespace \
  --set image.repository=registry.example.com/team/gops \
  --set image.tag=v1.0.0 \
  --set-string config.jwtSecret='replace-with-a-random-secret' \
  --set-string config.aesKey='0123456789abcdef0123456789abcdef'
```

生产环境建议使用已有 Secret，Secret 中默认需要 `jwt-secret` 和 `aes-key`：

```yaml
config:
  existingSecret: gops-secrets
```

应用使用 SQLite，默认启用一个 `ReadWriteOnce` PVC，因此 Deployment 使用 `Recreate` 策略并建议保持单副本。通过 `persistence.existingClaim` 可复用已有 PVC。
