# Flutter Client Stream Ledger

**Stream:** Flutter Client Stream  
**Scope:** Phase 7 (Flutter Client) — tasks.md T075–T084  
**Created:** 2026-07-26

| ID | Phase | Description | Status | Evidence |
|---|---|---|---|---|
| T075 | 7 | `flutter pub get` | NOT STARTED | rootless podman |
| T076 | 7 | `flutter analyze` 0 issues | NOT STARTED | analyzer output |
| T077 | 7 | Implement real terminal view | NOT STARTED | `clients/flutter/lib/widgets/terminal_view.dart` |
| T078 | 7 | Burn down prod TODOs | NOT STARTED | 65 markers |
| T079 | 7 | Add JSON serialization | NOT STARTED | 10 models |
| T080 | 7 | Wire design system | NOT STARTED | tokens/theme |
| T081 | 7 | Replace 8 BLoC stub tests | NOT STARTED | `clients/flutter/test/` |
| T082 | 7 | Generate golden images | NOT STARTED | screen×state×theme |
| T083 | 7 | Golden image dual validation | NOT STARTED | diff + OCR |
| T084 | 7 | Real e2e integration tests | NOT STARTED | `integration_test/app_test.dart` |

### First task: T076 — `flutter analyze`
- Command: `cd /home/milos/Factory/projects/tools_and_research/helix_terminator/clients/flutter && flutter analyze`
- Success criteria: analyzer exits 0.
