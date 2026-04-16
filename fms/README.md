# Firmware Management Service (FMS)

## Overview
Go-based Fabrica-generated service implementing firmware management resources and reconciliation logic.

## Implemented resources
- FirmwareProfile
- DeviceProfile
- UpdateProfile
- UpdateJob
- UpdateTask
- LookupJob

## Key features
- Library upload: POST /library/upload — ZIP extraction with zip-slip protection and creation of FirmwareProfile
- Reconciliation controllers: UpdateJob, UpdateTask, LookupJob
- Redfish helpers (lab/dev: InsecureSkipVerify), HMS mock, Ent + SQLite storage (./data.db)
- init-db command to create the database schema

## Quick start (run from repository root)

Initialize DB
```sh
go run ./cmd/server init-db
# Sample output:
# Database schema initialized successfully
```

Build server
```sh
go build -o bin/server ./cmd/server
```

Run server (default :8080)
```sh
./bin/server
# or
go run ./cmd/server
```

Health check
```sh
curl http://localhost:8080/health
# Sample output:
# {"status":"healthy","service":"fms"}
```

## Examples

Upload firmware library (multipart/form-data)
```sh
curl -X POST http://localhost:8080/library/upload -F "file=@/tmp/fw-bundle.zip"
```
Sample response (created FirmwareProfile)
```json
{
  "metadata": { "name": "example-fw", "uid": "firmwareprofile-501026e1" },
  "spec": { "versionString": "1.2.3", "softwareId": "com.example.fw", "targetComponent": "bios" }
}
```

Create an UpdateJob
```sh
curl -X POST http://localhost:8080/updatejobs \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": { "name": "update-1" },
    "spec": { "targetNodes": ["node-01"], "firmwareRef": "firmwareprofile-501026e1", "force": false }
  }'
```
Sample response (created UpdateJob)
```json
{
  "metadata": { "name": "update-1", "uid": "updatejob-549f5018" },
  "spec": { "targetNodes": ["node-01"], "firmwareRef": "firmwareprofile-501026e1" }
}
```

## Notes
- .gitignore added to prevent committing build artifacts (bin/, *.db, *.zip, logs, etc.).
- The service uses InsecureSkipVerify for Redfish calls in lab/test environments — do not enable this in production.
- If you need to remove the committed server binary from git history, perform a history rewrite (not done here).

## Contact
Repository remote: origin