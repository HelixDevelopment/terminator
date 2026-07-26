#!/usr/bin/env python3
import re
from pathlib import Path

REPO = Path(__file__).resolve().parent.parent
MODELS = [
    "clients/flutter/lib/models/ssh_key.dart",
    "clients/flutter/lib/models/audit_log.dart",
    "clients/flutter/lib/models/notification.dart",
    "clients/flutter/lib/models/organization.dart",
    "clients/flutter/lib/models/port_forward.dart",
    "clients/flutter/lib/models/recording.dart",
    "clients/flutter/lib/models/session.dart",
    "clients/flutter/lib/models/sftp_file.dart",
    "clients/flutter/lib/models/snippet.dart",
]


def snake_case(name: str) -> str:
    return re.sub(r"([A-Z])", r"_\1", name).lower().lstrip("_")


def process(path: Path):
    text = path.read_text()
    class_match = re.search(r"class\s+(\w+)\s*\{", text)
    if not class_match:
        print(f"skip {path}: no class")
        return
    class_name = class_match.group(1)
    fields = re.findall(r"^\s*final\s+([\w<>,?\s]+?)\s+(\w+);", text, re.M)

    from_lines = [f"  factory {class_name}.fromJson(Map<String, dynamic> json) {{"]
    from_lines.append(f"    return {class_name}(")
    to_lines = ["  Map<String, dynamic> toJson() {", "    return {"]

    for type_str, name in fields:
        type_clean = type_str.strip()
        json_key = snake_case(name)
        if type_clean == "String":
            from_lines.append(f"      {name}: json['{json_key}'] as String,")
            to_lines.append(f"      '{json_key}': {name},")
        elif type_clean == "int":
            from_lines.append(f"      {name}: json['{json_key}'] as int,")
            to_lines.append(f"      '{json_key}': {name},")
        elif type_clean == "bool":
            from_lines.append(f"      {name}: json['{json_key}'] as bool,")
            to_lines.append(f"      '{json_key}': {name},")
        elif type_clean == "DateTime":
            from_lines.append(f"      {name}: DateTime.parse(json['{json_key}'] as String),")
            to_lines.append(f"      '{json_key}': {name}.toIso8601String(),")
        elif type_clean == "Duration":
            from_lines.append(f"      {name}: Duration(milliseconds: json['{json_key}'] as int),")
            to_lines.append(f"      '{json_key}': {name}.inMilliseconds,")
        elif type_clean == "List<String>":
            from_lines.append(f"      {name}: (json['{json_key}'] as List<dynamic>?)?.cast<String>() ?? const [],")
            to_lines.append(f"      '{json_key}': {name},")
        elif type_clean == "String?":
            from_lines.append(f"      {name}: json['{json_key}'] as String?,")
            to_lines.append(f"      '{json_key}': {name},")
        else:
            from_lines.append(f"      {name}: json['{json_key}'], // TODO adapt type {type_clean}")
            to_lines.append(f"      '{json_key}': {name},")

    from_lines.append("    );")
    from_lines.append("  }")
    to_lines.append("    };")
    to_lines.append("  }")

    replacement = "\n" + "\n".join(from_lines) + "\n\n" + "\n".join(to_lines) + "\n"
    new_text = re.sub(r"\n\s*// TODO: add fromJson, toJson\n", replacement, text)
    path.write_text(new_text)
    print(f"updated {path}")


def main():
    for rel in MODELS:
        process(REPO / rel)


if __name__ == "__main__":
    main()
