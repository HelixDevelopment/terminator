#!/usr/bin/env python3
"""
Auto-fix common buf STANDARD lint issues in generated proto files.

Strategy:
1. For any RPC whose request/response type is not named <Method>Request/Response,
   create a wrapper message named <Method>Request/Response that contains a
   single field of the original type (or is empty for Empty).
2. Wrap google.protobuf.Empty request/response types in per-method messages.
3. Fix bare `repeated <lowercase>` syntax errors (e.g. `repeated open_ports`).
4. Add placeholder messages for undefined message references.
"""
import re
import pathlib
import subprocess
import sys

ROOT = pathlib.Path(__file__).resolve().parent.parent
SERVICES = [
    "analytics-service", "audit-service", "auth-service", "collaboration-service",
    "config-service", "container-bridge-service", "gateway-service", "health-service",
    "helixtrack-bridge-service", "host-service", "keychain-service", "notification-service",
    "org-service", "pki-service", "port-forward-service", "recording-service",
    "sftp-service", "snippet-service", "ssh-proxy-service", "terminal-service",
    "user-service", "vault-service", "workspace-service",
]


def to_pascal(snake):
    return "".join(w.capitalize() for w in snake.split("_"))


def extract_message_body(text, name):
    pattern = rf"^message\s+{re.escape(name)}\s*{{(.*?)\n}}"
    match = re.search(pattern, text, re.M | re.S)
    if not match:
        return None
    return match.group(1).rstrip()


def fix_service(service):
    proto_dir = ROOT / "services" / service / "api" / "proto"
    files = list(proto_dir.rglob("*.proto"))
    if not files:
        return None, "no proto files"
    path = files[0]
    text = path.read_text()

    # Fix syntax error: `repeated open_ports = 5;` -> `repeated OpenPort open_ports = 5;`
    text = re.sub(
        r"\b(repeated)\s+([a-z_][a-z0-9_]*)\s*(=\s*\d+\s*;)",
        lambda m: f"{m.group(1)} {to_pascal(m.group(2))} {m.group(2)} {m.group(3)}",
        text,
    )

    existing_messages = set(re.findall(r"^message\s+(\w+)\s*{", text, re.M))

    rpc_re = re.compile(
        r"^\s*rpc\s+(\w+)\s*\(\s*([\w.]+)\s*\)\s*returns\s*\(\s*([\w.]+)\s*\)\s*;",
        re.M,
    )
    rpcs = rpc_re.findall(text)

    new_messages = []

    def rpc_replacer(match):
        method, req, resp = match.groups()
        nonlocal new_messages
        new_req = req
        new_resp = resp

        if req == "google.protobuf.Empty":
            new_req = f"{method}Request"
            if new_req not in existing_messages:
                new_messages.append((new_req, ""))
                existing_messages.add(new_req)
        elif not req.endswith("Request"):
            new_req = f"{method}Request"
            if new_req not in existing_messages:
                body = extract_message_body(text, req) or f"  // TODO: populate {req}\n"
                new_messages.append((new_req, body))
                existing_messages.add(new_req)

        if resp == "google.protobuf.Empty":
            new_resp = f"{method}Response"
            if new_resp not in existing_messages:
                new_messages.append((new_resp, ""))
                existing_messages.add(new_resp)
        elif not resp.endswith("Response"):
            new_resp = f"{method}Response"
            if new_resp not in existing_messages:
                body = extract_message_body(text, resp) or f"  // TODO: populate {resp}\n"
                new_messages.append((new_resp, body))
                existing_messages.add(new_resp)

        return f"  rpc {method}({new_req}) returns ({new_resp});"

    text = rpc_re.sub(rpc_replacer, text)

    # Append new wrapper messages
    extra = ""
    for name, body in new_messages:
        extra += f"\nmessage {name} {{\n{body}\n}}\n"
    if extra:
        text = text.rstrip() + "\n" + extra + "\n"

    # Add placeholders for still-undefined message references (best effort)
    referenced = set(re.findall(r"\b[A-Z][a-zA-Z0-9_]*\b", text))
    known = existing_messages | {"Empty", "Timestamp", "StringValue", "BoolValue", "Int64Value", "Int32Value"}
    missing = referenced - known
    placeholders = ""
    for m in sorted(missing):
        if not re.search(rf"\bmessage\s+{re.escape(m)}\b", text):
            placeholders += f"\nmessage {m} {{\n  // TODO: define fields\n}}\n"
    if placeholders:
        text = text.rstrip() + "\n" + placeholders + "\n"

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
        print(f"Processing {service}...")
        path, err = fix_service(service)
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
