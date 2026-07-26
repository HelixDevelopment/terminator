# Falco Rules

This directory deploys real-time runtime security rules for the HelixTerminator platform via Falco. The rules detect container escape, privilege escalation, unauthorized data access, and anomalous outbound connections from sensitive services.

## Files

- `falco-rules.yaml` — ConfigMap with the rule set consumed by Falco (DaemonSet).

## Rule Summary

| Rule | Priority | Purpose |
|------|----------|---------|
| Terminal Shell in Container | WARNING | Detect shell execution inside containers |
| Unauthorized Database Access | NOTICE | Detect direct database client execution |
| Sensitive File Access | WARNING | Detect reads of /etc/shadow, .ssh/, etc. |
| Outbound Connection from Sensitive Service | CRITICAL | Detect unexpected egress from auth/vault/pki |
| Privilege Escalation | CRITICAL | Detect sudo/su/setuid use |
| Kubernetes Secret Access | WARNING | Detect access to service-account tokens |
| Crypto Mining Detection | CRITICAL | Detect known mining process names |
| Write to /etc | WARNING | Detect writes under /etc |
| SSH Connection from Container | NOTICE | Detect outbound port-22 connections |
| Modification of SSH Configuration | WARNING | Detect writes to /etc/ssh/ |

## Deployment

```bash
kubectl apply -n helixterminator -f infrastructure/security/falco/falco-rules.yaml
```

Falco must be configured to load `falco_rules.local.yaml` from the ConfigMap. Alerts should be shipped to the security SIEM via the configured Falco sidecar or `falcosidekick`.

## Validation

```bash
kubectl -n helixterminator get configmap falco-rules -o jsonpath='{.data.falco_rules\\.local\\.yaml}' | head -20
```
