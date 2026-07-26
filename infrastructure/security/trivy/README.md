# Trivy Security Scanning

## Configuration

### Scan All Services
```bash
# Scan filesystem
trivy fs --severity HIGH,CRITICAL .

# Scan Docker images
for service in services/*/; do
  trivy image "helixterminator/$(basename $service):latest"
done

# Scan Kubernetes manifests
trivy config infrastructure/kubernetes/

# Scan Terraform
trivy config infrastructure/terraform/
```

### CI/CD Integration
```yaml
- name: Trivy vulnerability scan
  uses: aquasecurity/trivy-action@master
  with:
    scan-type: 'fs'
    format: 'sarif'
    output: 'trivy-results.sarif'
    severity: 'CRITICAL,HIGH'
    exit-code: '1'
```

### Ignore File
Create `.trivyignore` for known acceptable vulnerabilities:
```
# CVE-2023-XXXX: Accepted risk - only affects debug endpoint
CVE-2023-XXXX
```

## CI Scanning Steps

1. **Repository checkout** — pull source code.
2. **Dependency cache restore** — restore Go module / npm / pip caches.
3. **Run Trivy filesystem scan**:
   ```bash
   trivy fs --severity HIGH,CRITICAL --exit-code 1 --scanners vuln,secret,config .
   ```
4. **Build and scan container images**:
   ```bash
   docker build -t helixterminator/$SERVICE:$TAG services/$SERVICE
   trivy image --severity HIGH,CRITICAL --exit-code 1 helixterminator/$SERVICE:$TAG
   ```
5. **Scan IaC**:
   ```bash
   trivy config --severity HIGH,CRITICAL --exit-code 1 infrastructure/helm infrastructure/terraform
   ```
6. **Upload SARIF** to the security dashboard.
7. **Fail the build** on any CRITICAL or HIGH finding unless listed in `.trivyignore` with documented justification.
