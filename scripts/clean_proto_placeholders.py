#!/usr/bin/env python3
"""
Remove incorrectly-added placeholder messages that conflict with service/method names.
"""
import re
import pathlib
import subprocess

ROOT = pathlib.Path(__file__).resolve().parent.parent
SERVICES = [
    "analytics-service", "audit-service", "auth-service", "collaboration-service",
    "config-service", "container-bridge-service", "gateway-service", "health-service",
    "helixtrack-bridge-service", "host-service", "keychain-service", "notification-service",
    "org-service", "pki-service", "port-forward-service", "recording-service",
    "sftp-service", "snippet-service", "ssh-proxy-service", "terminal-service",
    "user-service", "vault-service", "workspace-service",
]


def clean_service(service):
    proto_dir = ROOT / "services" / service / "api" / "proto"
    files = list(proto_dir.rglob("*.proto"))
    if not files:
        return None, "no proto"
    path = files[0]
    text = path.read_text()

    # Parse service name and method names
    service_match = re.search(r"^service\s+(\w+)\s*{", text, re.M)
    if not service_match:
        return path, "no service block"
    service_name = service_match.group(1)
    method_names = set(re.findall(r"^\s*rpc\s+(\w+)\s*\(", text, re.M))
    forbidden = method_names | {service_name}

    # Remove any message block whose name is a forbidden word
    def remove_message(match):
        name = match.group(1)
        if name in forbidden:
            return ""
        return match.group(0)

    # Match message blocks including nested braces (proto messages don't nest in our files)
    pattern = re.compile(r"^message\s+(\w+)\s*{(.*?)\n}\n", re.M | re.S)
    text = pattern.sub(remove_message, text)

    # Clean up excessive blank lines
    text = re.sub(r"\n{3,}", "\n\n", text)

    path.write_text(text)
    return path, None


def run_buf_lint(service):
    try:
        result = subprocess.run(
            ["buf", "lint", f"services/{service}/api/proto"],
            cwd=ROOT,
            capture_output=True,
            text=True,
        )
        return result.returncode == 0, (result.stdout + result.stderr).strip()
    except FileNotFoundError:
        return False, "buf not installed"


def run_go_test(service):
    try:
        result = subprocess.run(
            ["go", "test", "./..."],
            cwd=ROOT / "services" / service,
            capture_output=True,
            text=True,
        )
        return result.returncode == 0, (result.stdout + result.stderr).strip()
    except FileNotFoundError:
        return False, "go not installed"


def main():
    results = {}
    completed = []
    for service in SERVICES:
        print(f"Cleaning {service}...")
        path, err = clean_service(service)
        if err:
            results[service] = {"error": err}
            continue
        lint_ok, lint_out = run_buf_lint(service)
        test_ok, test_out = run_go_test(service)
        results[service] = {
            "lint": "PASS" if lint_ok else "FAIL",
            "test": "PASS" if test_ok else "FAIL",
            "lint_output": lint_out,
            "test_output": test_out,
        }
        if lint_ok and test_ok:
            completed.append(service)

    print("\n=== RESULTS ===")
    for service, status in results.items():
        print(f"{service}:")
        if "error" in status:
            print(f"  ERROR: {status['error']}")
        else:
            print(f"  lint: {status['lint']}")
            print(f"  test: {status['test']}")
            if status["lint"] == "FAIL":
                print(f"  lint_output: {status['lint_output'][:500]}")
            if status["test"] == "FAIL":
                print(f"  test_output: {status['test_output'][:500]}")

    print(f"\nCompleted {len(completed)}/{len(SERVICES)} services: {', '.join(completed)}")


if __name__ == "__main__":
    main()
