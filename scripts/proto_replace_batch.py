#!/usr/bin/env python3
"""
Replace placeholder protos in services with real contracts derived from
reference protos under docs/research/mvp/final/implementation/04-api-specification/proto.

Usage:
    python3 scripts/proto_replace_batch.py [service_dir_basename ...]

If no service is specified, processes all services except ai-service and billing-service.
"""
import os
import re
import subprocess
import sys
from pathlib import Path

REPO = Path(__file__).resolve().parent.parent
SERVICES_DIR = REPO / "services"
REF_DIR = REPO / "docs" / "research" / "mvp" / "final" / "implementation" / "04-api-specification" / "proto"

SERVICE_CONFIG = {
    "analytics-service": ("analytics", "analytics_service.v1", "AnalyticsService", "analytics_servicev1"),
    "audit-service": ("audit", "audit_service.v1", "AuditService", "audit_servicev1"),
    "auth-service": ("auth", "auth_service.v1", "AuthService", "auth_servicev1"),
    "collaboration-service": ("collab", "collaboration_service.v1", "CollaborationService", "collaboration_servicev1"),
    "config-service": ("config", "config_service.v1", "ConfigService", "config_servicev1"),
    "container-bridge-service": ("container-bridge", "container_bridge_service.v1", "ContainerBridgeService", "container_bridge_servicev1"),
    "gateway-service": ("gateway", "gateway_service.v1", "GatewayService", "gateway_servicev1"),
    "health-service": ("health", "health_service.v1", "HealthService", "health_servicev1"),
    "helixtrack-bridge-service": ("helixtrack-bridge", "helixtrack_bridge_service.v1", "HelixtrackBridgeService", "helixtrack_bridge_servicev1"),
    "host-service": ("host", "host_service.v1", "HostService", "host_servicev1"),
    "keychain-service": ("keychain", "keychain_service.v1", "KeychainService", "keychain_servicev1"),
    "notification-service": ("notification", "notification_service.v1", "NotificationService", "notification_servicev1"),
    "org-service": ("org", "org_service.v1", "OrgService", "org_servicev1"),
    "pki-service": ("pki", "pki_service.v1", "PkiService", "pki_servicev1"),
    "port-forward-service": ("port-forward", "port_forward_service.v1", "PortForwardService", "port_forward_servicev1"),
    "recording-service": ("recording", "recording_service.v1", "RecordingService", "recording_servicev1"),
    "sftp-service": ("sftp", "sftp_service.v1", "SftpService", "sftp_servicev1"),
    "snippet-service": ("snippet", "snippet_service.v1", "SnippetService", "snippet_servicev1"),
    "ssh-proxy-service": ("session", "ssh_proxy_service.v1", "SshProxyService", "ssh_proxy_servicev1"),
    "terminal-service": ("session", "terminal_service.v1", "TerminalService", "terminal_servicev1"),
    "user-service": ("user", "user_service.v1", "UserService", "user_servicev1"),
    "vault-service": ("vault", "vault_service.v1", "VaultService", "vault_servicev1"),
    "workspace-service": ("workspace", "workspace_service.v1", "WorkspaceService", "workspace_servicev1"),
}


def to_snake_case(name):
    s = re.sub(r'[^a-zA-Z0-9]', '_', name)
    return re.sub(r'_+', '_', s).strip('_').lower()


def read_reference(ref_name):
    path = REF_DIR / f"{ref_name}.proto"
    if not path.exists():
        raise FileNotFoundError(f"Reference proto not found: {path}")
    return path.read_text()


def normalize_content(content, package, service_name, go_package_suffix, service_dir):
    # Replace package declaration
    content = re.sub(r'^\s*package\s+[^;]+;', f'package {package};', content, count=1, flags=re.MULTILINE)
    # Replace go_package
    go_package = f"github.com/helixdevelopment/{service_dir}/api/proto/{package.replace('.', '/')};{go_package_suffix}"
    content = re.sub(
        r'^\s*option\s+go_package\s*=\s*"[^"]+";',
        f'option go_package = "{go_package}";',
        content,
        count=1,
        flags=re.MULTILINE,
    )
    # Replace service declaration name
    content = re.sub(
        r'^\s*service\s+(\w+)\s*\{',
        f"service {service_name} {{",
        content,
        count=1,
        flags=re.MULTILINE,
    )
    # Remove any TODO comment line about define service RPCs
    content = re.sub(r'//\s*TODO:\s*define service RPCs and messages\s*\n', '', content)
    return content


def write_proto(service_dir, package, content):
    parts = package.split(".")
    proto_dir = SERVICES_DIR / service_dir / "api" / "proto" / parts[0] / parts[1]
    proto_dir.mkdir(parents=True, exist_ok=True)
    proto_file = proto_dir / f"{parts[0]}.proto"
    proto_file.write_text(content)
    return proto_file


def cleanup_old_proto(service_dir):
    old = SERVICES_DIR / service_dir / "api" / "proto" / "notification-service.proto"
    if old.exists():
        old.unlink()


def run_buf_lint(service_dir):
    try:
        result = subprocess.run(
            ["buf", "lint", f"services/{service_dir}/api/proto"],
            cwd=REPO,
            capture_output=True,
            text=True,
        )
        return result.returncode == 0, result.stdout + result.stderr
    except FileNotFoundError:
        return False, "buf not installed"


def run_go_test(service_dir):
    try:
        result = subprocess.run(
            ["go", "test", "./..."],
            cwd=SERVICES_DIR / service_dir,
            capture_output=True,
            text=True,
        )
        return result.returncode == 0, result.stdout + result.stderr
    except FileNotFoundError:
        return False, "go not installed"


def main():
    if not REF_DIR.exists():
        print(f"Reference proto directory not found: {REF_DIR}")
        sys.exit(1)

    requested = sys.argv[1:]
    if requested:
        services = {k: v for k, v in SERVICE_CONFIG.items() if k.replace("-service", "") in requested or k in requested}
    else:
        services = SERVICE_CONFIG

    results = {}
    completed = []

    for service_dir, (ref_name, package, service_name, go_package_suffix) in services.items():
        print(f"Processing {service_dir}...")
        try:
            ref_content = read_reference(ref_name)
            new_content = normalize_content(ref_content, package, service_name, go_package_suffix, service_dir)
            cleanup_old_proto(service_dir)
            proto_file = write_proto(service_dir, package, new_content)
            print(f"  Wrote {proto_file}")

            lint_ok, lint_output = run_buf_lint(service_dir)
            test_ok, test_output = run_go_test(service_dir)
            results[service_dir] = {
                "lint": "PASS" if lint_ok else "FAIL",
                "test": "PASS" if test_ok else "FAIL",
                "lint_output": lint_output,
                "test_output": test_output,
            }
            if lint_ok and test_ok:
                completed.append(service_dir)
        except Exception as e:
            results[service_dir] = {"error": str(e)}

    print("\n=== RESULTS ===")
    for service, status in results.items():
        print(f"{service}:")
        if "error" in status:
            print(f"  ERROR: {status['error']}")
        else:
            print(f"  lint: {status['lint']}")
            print(f"  test: {status['test']}")
            if status.get("lint") == "FAIL":
                print(f"  lint_output: {status['lint_output'][:500]}")
            if status.get("test") == "FAIL":
                print(f"  test_output: {status['test_output'][:500]}")

    print(f"\nCompleted {len(completed)}/{len(services)} services: {', '.join(completed)}")


if __name__ == "__main__":
    main()
