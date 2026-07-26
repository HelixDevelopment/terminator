# Cosign Image Signing Configuration

## Setup

### Generate Key Pair
```bash
cosign generate-key-pair
```

This creates:
- `cosign.key` (private key - store in AWS Secrets Manager)
- `cosign.pub` (public key - commit to repo)

## Signing Images

### CI/CD Pipeline
```bash
# Sign image after build
cosign sign --key env://COSIGN_PRIVATE_KEY \\
  $REGISTRY/helixterminator/$SERVICE:$TAG

# Verify signature before deployment
cosign verify --key cosign.pub \\
  $REGISTRY/helixterminator/$SERVICE:$TAG
```

### GitHub Actions
```yaml
- name: Sign image
  uses: sigstore/cosign-installer@v3
- run: |
    cosign sign --key env://COSIGN_PRIVATE_KEY \\
      --yes ${{ env.REGISTRY }}/${{ github.repository }}/${{ matrix.service }}:${{ github.sha }}
  env:
    COSIGN_PRIVATE_KEY: ${{ secrets.COSIGN_PRIVATE_KEY }}
    COSIGN_PASSWORD: ${{ secrets.COSIGN_PASSWORD }}
```

## Verification in Kubernetes

### Kyverno Policy
```yaml
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: verify-image-signatures
spec:
  validationFailureAction: Enforce
  rules:
    - name: verify-cosign-signature
      match:
        resources:
          kinds:
            - Pod
      verifyImages:
        - imageReferences:
            - "ghcr.io/helixdevelopment/*"
          attestors:
            - entries:
                - keys:
                    publicKeys: |
                      -----BEGIN PUBLIC KEY-----
                      MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE...
                      -----END PUBLIC KEY-----
```

## Verification Checklist

- [x] Generate key pair per environment.
- [x] Store private key in a KMS (AWS Secrets Manager / GCP Secret Manager / Azure Key Vault).
- [x] Commit `cosign.pub` to the repository under `infrastructure/security/cosign/`.
- [x] Sign images in CI immediately after build.
- [x] Verify signatures in the deployment pipeline before `kubectl apply`.
- [x] Enforce verification with Kyverno or OPA Gatekeeper.
