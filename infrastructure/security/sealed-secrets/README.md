# Sealed Secrets

This directory documents the sealed-secrets workflow used to commit encrypted Kubernetes Secrets to Git without exposing plaintext values.

## SealedSecret Templates

The following template can be applied per environment. Replace `<ENV>` and values with the target environment and base64-encoded secret contents.

```yaml
---
apiVersion: bitnami.com/v1alpha1
kind: SealedSecret
metadata:
  name: helix-postgres-secret
  namespace: helixterminator
spec:
  encryptedData:
    password: AgB...
---
apiVersion: bitnami.com/v1alpha1
kind: SealedSecret
metadata:
  name: helix-redis-secret
  namespace: helixterminator
spec:
  encryptedData:
    redis-password: AgB...
---
apiVersion: bitnami.com/v1alpha1
kind: SealedSecret
metadata:
  name: helix-kafka-secret
  namespace: helixterminator
spec:
  encryptedData:
    password: AgB...
```

## Creating a SealedSecret

```bash
# Create a plain Secret in a temp file
kubectl create secret generic helix-example-secret \\
  --from-literal=token=super-secret-token \\
  --dry-run=client -o yaml > /tmp/secret.yaml

# Seal it with the cluster's public certificate
kubeseal --controller-namespace=sealed-secrets \\
  --controller-name=sealed-secrets \\
  --format yaml < /tmp/secret.yaml > infrastructure/security/sealed-secrets/example-sealed.yaml

# Apply the sealed secret to any cluster that owns the matching private key
kubectl apply -f infrastructure/security/sealed-secrets/example-sealed.yaml
```

## Encryption Key Management

- The Sealed Secrets controller holds the only private key.
- Backup the controller's private key to a KMS-encrypted object store.
- Rotate keys annually by generating a new key pair and re-sealing secrets.
- Never commit plaintext Secrets; only `SealedSecret` resources should be in Git.

## Per-Environment Layout

```text
infrastructure/security/sealed-secrets/
├── dev/
├── staging/
└── production/
```

Each environment directory contains a `kustomization.yaml` and the sealed secrets for that cluster.
