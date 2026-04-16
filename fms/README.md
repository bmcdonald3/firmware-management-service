# Firmware Management Service (FMS)

Summary
- Go-based Fabrica-generated service implementing firmware management resources:
  - FirmwareProfile, DeviceProfile, UpdateProfile, UpdateJob, UpdateTask, LookupJob
- Features implemented:
  - Library upload (POST /library/upload) — ZIP extraction with zip-slip protection and creation of FirmwareProfile
  - Reconciliation controllers for UpdateJob, UpdateTask, LookupJob
  - Redfish helpers (insecure TLS support for lab/dev), simple HMS mock, SQLite Ent storage (./data.db)
  - init-db command to create schema

Quick commands (run from repository root)

- Initialize DB
  go run ./cmd/server init-db
  Sample output:
  Database schema initialized successfully

- Build server
  go build -o bin/server ./cmd/server

- Run server (default :8080)
  ./bin/server
  or
  go run ./cmd/server

- Health check
  curl http://localhost:8080/health
  Sample output:
  {"status":"healthy","service":"fms"}

Example: upload firmware library
  curl -X POST http://localhost:8080/library/upload -F "file=@/tmp/fw-bundle.zip"
  Sample response (created FirmwareProfile):
  {
    "metadata": { "name": "example-fw", "uid": "firmwareprofile-501026e1" },
    "spec": { "versionString": "1.2.3", "softwareId": "com.example.fw", "targetComponent": "bios" }
  }

Example: create UpdateJob
  curl -X POST http://localhost:8080/updatejobs -H "Content-Type: application/json" -d '{
    "metadata": { "name": "update-1" },
    "spec": { "targetNodes": ["node-01"], "firmwareRef": "firmwareprofile-501026e1", "force": false }
  }'
  Sample response (created UpdateJob):
  {
    "metadata": { "name": "update-1", "uid": "updatejob-549f5018" },
    "spec": { "targetNodes": ["node-01"], "firmwareRef": "firmwareprofile-501026e1" }
  }

Notes
- A .gitignore was added to prevent committing build artifacts (bin/, *.db, *.zip, logs, etc.).
- The service intentionally uses InsecureSkipVerify for Redfish calls in lab/test environments; do not use this in production.
- The committed server binary should be removed from git history if long-term cleanup is required (not performed here).

Contact
- Repository: origin (see repo remote)