#!/usr/bin/env python3
"""
Fix buf STANDARD lint issues in the 23 generated service proto files.

Actions per proto:
- Remove unused google/protobuf/empty.proto imports.
- Ensure every RPC uses <Method>Request / <Method>Response.
- Duplicate shared request/response messages so each RPC has unique types.
- Create per-method request/response copies for streaming RPCs.

After fixing, runs `buf lint` and `go test ./...` per service and reports status.
"""
import re
import subprocess
from collections import Counter
from pathlib import Path

REPO = Path(__file__).resolve().parent.parent

SERVICES = [
    "analytics-service",
    "audit-service",
    "auth-service",
    "collaboration-service",
    "config-service",
    "container-bridge-service",
    "gateway-service",
    "health-service",
    "helixtrack-bridge-service",
    "host-service",
    "keychain-service",
    "notification-service",
    "org-service",
    "pki-service",
    "port-forward-service",
    "recording-service",
    "sftp-service",
    "snippet-service",
    "ssh-proxy-service",
    "terminal-service",
    "user-service",
    "vault-service",
    "workspace-service",
]


def find_closing_brace(text: str, open_idx: int) -> int:
    count = 0
    for i in range(open_idx, len(text)):
        if text[i] == "{":
            count += 1
        elif text[i] == "}":
            count -= 1
            if count == 0:
                return i
    return -1


def parse_blocks(text: str):
    blocks = []
    pattern = re.compile(r"^\s*(message|service)\s+(\w+)\s*\{", re.MULTILINE)
    for m in pattern.finditer(text):
        kind, name = m.group(1), m.group(2)
        open_idx = text.find("{", m.end())
        close_idx = find_closing_brace(text, open_idx)
        blocks.append(
            {
                "kind": kind,
                "name": name,
                "start": m.start(),
                "open": open_idx,
                "close": close_idx,
                "body": text[open_idx + 1 : close_idx],
            }
        )
    return blocks


RPC_LINE_RE = re.compile(
    r"^(\s*)rpc\s+(\w+)\s*\(\s*(stream\s+)?([\w.]+)\s*\)\s+returns\s+\(\s*(stream\s+)?([\w.]+)\s*\)\s*;(.*)$",
    re.MULTILINE,
)


def fix_proto(path: Path) -> tuple[bool, str]:
    text = path.read_text()

    # Remove unused google/protobuf/empty.proto imports.
    text = re.sub(
        r'^\s*import\s+"google/protobuf/empty\.proto"\s*;\s*\n',
        "",
        text,
        flags=re.MULTILINE,
    )

    blocks = parse_blocks(text)
    service_block = next((b for b in blocks if b["kind"] == "service"), None)
    if not service_block:
        return False, "no service block found"

    messages = {b["name"]: b["body"] for b in blocks if b["kind"] == "message"}
    svc_text = text[service_block["start"] : service_block["close"] + 1]

    rpcs = []
    for m in RPC_LINE_RE.finditer(svc_text):
        rpcs.append(
            {
                "method": m.group(2),
                "req_stream": m.group(3) is not None,
                "req_type": m.group(4),
                "resp_stream": m.group(5) is not None,
                "resp_type": m.group(6),
                "match": m,
            }
        )

    if not rpcs:
        return False, "no RPCs found"

    req_usage = Counter(r["req_type"] for r in rpcs)
    resp_usage = Counter(r["resp_type"] for r in rpcs)

    updates = {}
    new_messages = {}

    for r in rpcs:
        method = r["method"]
        expected_req = f"{method}Request"
        expected_resp = f"{method}Response"

        actual_req = r["req_type"]
        actual_resp = r["resp_type"]

        # Determine new request type
        if actual_req == expected_req and req_usage[actual_req] == 1:
            new_req = actual_req
        else:
            if expected_req not in messages and expected_req not in new_messages:
                new_messages[expected_req] = messages.get(actual_req, "")
            new_req = expected_req

        # Determine new response type
        if actual_resp == expected_resp and resp_usage[actual_resp] == 1:
            new_resp = actual_resp
        else:
            if expected_resp not in messages and expected_resp not in new_messages:
                new_messages[expected_resp] = messages.get(actual_resp, "")
            new_resp = expected_resp

        updates[method] = (new_req, new_resp)

    # Rebuild service block line-by-line.
    new_svc_lines = []
    for line in svc_text.splitlines(keepends=True):
        m = RPC_LINE_RE.match(line)
        if not m:
            new_svc_lines.append(line)
            continue

        method = m.group(2)
        req_stream = "stream " if m.group(3) else ""
        resp_stream = "stream " if m.group(5) else ""
        new_req, new_resp = updates[method]
        indent = m.group(1)
        remainder = m.group(7)
        new_line = f"{indent}rpc {method}({req_stream}{new_req}) returns ({resp_stream}{new_resp});{remainder}"
        if not new_line.endswith("\n"):
            new_line += "\n"
        new_svc_lines.append(new_line)

    new_svc_text = "".join(new_svc_lines)

    new_text = text[: service_block["start"]] + new_svc_text + text[service_block["close"] + 1 :]

    if new_messages:
        extra = []
        for name in new_messages:
            body = new_messages[name]
            extra.append(f"\nmessage {name} {{\n{body}\n}}\n")
        new_text = new_text.rstrip() + "".join(extra) + "\n"

    path.write_text(new_text)
    return True, ""


def run_buf_lint(service_dir: str):
    try:
        result = subprocess.run(
            ["buf", "lint", f"services/{service_dir}/api/proto"],
            cwd=REPO,
            capture_output=True,
            text=True,
            timeout=60,
        )
        return result.returncode == 0, (result.stdout + result.stderr).strip()
    except FileNotFoundError:
        return False, "buf not installed"
    except subprocess.TimeoutExpired:
        return False, "buf lint timed out"


def run_go_test(service_dir: str):
    try:
        result = subprocess.run(
            ["go", "test", "./..."],
            cwd=REPO / "services" / service_dir,
            capture_output=True,
            text=True,
            timeout=120,
        )
        return result.returncode == 0, (result.stdout + result.stderr).strip()
    except FileNotFoundError:
        return False, "go not installed"
    except subprocess.TimeoutExpired:
        return False, "go test timed out"


def main():
    results = {}
    passed = []

    for service in SERVICES:
        proto_dir = REPO / "services" / service / "api" / "proto"
        proto_files = list(proto_dir.rglob("*.proto"))
        if not proto_files:
            results[service] = {
                "lint": "FAIL",
                "test": "SKIP",
                "lint_output": "no proto files",
                "test_output": "",
            }
            continue

        proto_file = proto_files[0]
        print(f"Processing {service}...")

        ok, err = fix_proto(proto_file)
        if not ok:
            results[service] = {
                "lint": "FAIL",
                "test": "SKIP",
                "lint_output": err,
                "test_output": "",
            }
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
            passed.append(service)

    print("\n=== RESULTS ===")
    for service, status in results.items():
        print(f"{service}:")
        print(f"  lint: {status['lint']}")
        print(f"  test: {status['test']}")
        if status["lint"] == "FAIL":
            print(f"  lint_output: {status['lint_output']}")
        if status["test"] == "FAIL":
            print(f"  test_output: {status['test_output'][:500]}")

    print(f"\nCompleted {len(passed)}/{len(SERVICES)} services: {', '.join(passed)}")


if __name__ == "__main__":
    main()
