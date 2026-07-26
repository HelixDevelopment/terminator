#!/usr/bin/env python3
import os
import subprocess
from pathlib import Path

REPO = Path(__file__).resolve().parent.parent


def main():
    failures = []
    for svc in sorted((REPO / "services").iterdir()):
        if not svc.is_dir() or not svc.name.endswith("-service"):
            continue
        if not (svc / "go.mod").exists():
            continue
        print(f"=== {svc.name} ===")
        result = subprocess.run(
            ["go", "test", "./..."],
            cwd=svc,
            capture_output=True,
            text=True,
        )
        if result.returncode != 0:
            print(result.stdout[-500:] if result.stdout else result.stderr[-500:])
            failures.append(svc.name)
        else:
            print("PASS")

    print("\n=== SUMMARY ===")
    print(f"Passed: {25 - len(failures)}/25")
    if failures:
        print(f"Failures: {', '.join(failures)}")


if __name__ == "__main__":
    main()
