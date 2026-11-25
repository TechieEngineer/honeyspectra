# HoneySpectra — AI Development Handover

## 1. Instructions for the AI

Act as a Senior Cybersecurity Engineer and Full-Stack
Systems Architect. Continue developing this existing project
instead of rebuilding it. Inspect the repository before
modifying anything.

Provide exact file paths, complete code, PowerShell commands,
verification steps and Git commit commands after every
significant development milestone.

## 2. Project Identity

Name: HoneySpectra
Version: V1.0.0 baseline
Internal Go module: honeypot-core
Environment: Windows, PowerShell, VS Code, Docker
Local path: D:\Projects\honeypot-core

Repository:
https://github.com/TechieEngineer/honeyspectra

Intended public demo:
https://techieengineer.github.io/honeyspectra/

## 3. Architecture

Go concurrent networking engine
    -> Buffered structured JSONL logger
    -> Python FastAPI analytics backend
    -> Server-Sent Events (SSE)
    -> Interactive JavaScript SOC dashboard

A separate static dashboard uses synthetic data and runs
independently through GitHub Pages.

## 4. Filesystem

cmd/honeypot/main.go
    Concurrent networking engine and listener lifecycle.

configs/honeypot.yaml
    Protocol definitions, ports and resource configuration.

internal/config/
    Configuration parsing and validation.

internal/logger/
    Thread-safe asynchronous JSONL event logging.

internal/protocols/
    Unified protocol interface, adapter registry and handlers.

backend/app.py
    FastAPI, event processing, analytics and SSE streaming.

backend/static/
    Local dashboard HTML, CSS, JavaScript and theme management.

docs/
    Standalone public dashboard, synthetic fixtures,
    sample telemetry and threat report template.

data/
    Local runtime telemetry. Excluded from public publication.

simulator.py
    Synthetic multi-protocol event generator.

tests/
    Python API and analytics tests.

.github/workflows/
    Automated CI testing.

Dockerfile and docker-compose.yml
    Containerization and resource isolation configuration.

README.md, SECURITY.md, CONTINUATION.md
    Documentation, security boundaries and development handover.

requirements.txt, requirements-dev.txt, go.mod
    Python and Go dependencies.

## 5. Current Implementation

- Configuration-driven concurrent Go networking.
- SSH, HTTP, Telnet, FTP, MySQL and TCP protocol models.
- Actual local HTTP/TCP listeners and protocol demonstrations.
- Bounded asynchronous JSON logging.
- FastAPI analytics and real-time SSE.
- Dark/light theme with persistent preferences.
- Protocol filtering, pagination and session replay.
- Synthetic MITRE ATT&CK, credential and payload analytics.
- Standalone GitHub Pages-compatible dashboard.
- Automated Go and Python tests.
- Docker configuration and GitHub Actions.

The local dashboard worked during initial development.
Verify the latest GitHub publication, release tag, Pages
deployment and test results before assuming completion.

## 6. Restart Commands

Clone the repository if the local copy no longer exists.

Initialize Python:
py -3 -m venv .venv
.\.venv\Scripts\python.exe -m pip install -r requirements-dev.txt

Validate:
go mod tidy
go test ./...
go vet ./...
.\.venv\Scripts\python.exe -m pytest -v

Terminal 1:
go run ./cmd/honeypot

Terminal 2:
.\.venv\Scripts\python.exe -m uvicorn backend.app:app --host 127.0.0.1 --port 8000 --limit-concurrency 32

Terminal 3:
.\.venv\Scripts\python.exe simulator.py --count 100 --interval 0.25

Local dashboard:
http://127.0.0.1:8000

Static preview:
.\.venv\Scripts\python.exe -m http.server 5500 --bind 127.0.0.1 --directory docs

## 7. Future Development Roadmap

Phase 1 - Reliability
- Audit the existing source code and dependencies.
- Expand unit, integration and concurrency tests.
- Verify Docker configuration.
- Implement log rotation and retention policies.

Phase 2 - Security Hardening
- Add dashboard authentication and authorization.
- Add per-IP rate limits and stronger resource caps.
- Validate input handling and network isolation.
- Conduct independent security assessments.

Phase 3 - Threat Intelligence
- Introduce local GeoIP/ASN enrichment.
- Improve MITRE ATT&CK classification.
- Add historical analytics and reporting.
- Implement an interactive geographical visualization.
- Improve synthetic protocol and session models.

Phase 4 - Performance and UI
- Benchmark Go concurrency and memory consumption.
- Implement virtualized event tables.
- Optimize SSE streaming and chart rendering.
- Improve session replay and accessibility.
- Introduce persistent indexed telemetry storage.

Phase 5 - Release Engineering
- Improve CI/CD and dependency security scanning.
- Expand documentation and architecture diagrams.
- Maintain the standalone public demo.
- Publish tested, versioned GitHub releases.

## 8. Non-Negotiable Security Requirements

Keep the current engine isolated and localhost-only.
Never execute untrusted commands or download real malware.
Never publish real credentials, secrets or private telemetry.
Use synthetic fixtures for public demonstrations.
Do not expose V1 listeners directly to the internet.
Any future live research deployment requires professional
supervision, independent auditing and verified containment.

## 9. Development Workflow

Inspect -> Plan -> Implement -> Test -> Commit -> Push.

Preserve existing functionality and Git history.
Avoid unnecessary rewrites and dependency bloat.
Never claim production readiness without verification.
Document all limitations honestly.

Start by auditing the latest repository state and identifying
the first unfinished roadmap task.
