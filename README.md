# Firmware Management Service (FMS)

## Overview
FMS is a small controller/service for managing firmware updates and inventory for devices using Redfish and SSH. It exposes a CRUD HTTP API (generated via Fabrica) for DeviceProfiles, UpdateProfiles, FirmwareManifests, UpdateJobs and LookupJobs, and includes reconcilers to drive update and lookup workflows.

## What was implemented
- Fabrica project and generated layers (handlers, routes, OpenAPI, ent storage)
  - fms/.fabrica.yaml
  - fms/apis.yaml
  - fms/pkg/apiversion/registry_generated.go
- API types (Spec + Status + validation)
  - fms/apis/firmware.management.io/v1/deviceprofile_types.go
  - fms/apis/firmware.management.io/v1/updateprofile_types.go
  - fms/apis/firmware.management.io/v1/firmwaremanifest_types.go
  - fms/apis/firmware.management.io/v1/updatejob_types.go
  - fms/apis/firmware.management.io/v1/lookupjob_types.go
- CLI/server wiring
  - fms/cmd/server/main.go (init-db command, viper defaults)
- Storage and DB migrations
  - internal/storage/ent/* (generated ent schema + adapters)
- Credentials parsing
  - fms/pkg/credentials/parser.go (reads /tmp/credentials.json)
- File-based firmware library server (simple FMLS)
  - fms/pkg/fmls/server.go — upload/download/delete under /tmp/firmware
- Reconcilers (core logic)
  - fms/pkg/reconcilers/updatejob_reconciler.go (state machine: Preflight, Execute, Monitor)
  - fms/pkg/reconcilers/lookupjob_reconciler.go (collect firmware inventory via Redfish)
  - fms/pkg/reconcilers/ssh_helper.go (SSH/SCP runner using golang.org/x/crypto/ssh)

## What was tested
- Fabrica toolchain: fabrica_init, fabrica_add_resource (force), fabrica_generate
- go mod tidy / go build
- Database init and migrations via server init-db
- Basic API end-to-end smoke tests with curl (create/list resources)
- Start/stop server and verified endpoints

## How to build and run (local)
- Build:
  - cd fms && go build -o bin/server ./cmd/server
- Initialize DB:
  - cd fms && go run ./cmd/server init-db
- Run server (background):
  - cd fms && FMS_HOST_IP=127.0.0.1 ./bin/server &
- Smoke tests (examples):
  - curl -s http://localhost:8080/health
  - curl -s -X POST http://localhost:8080/deviceprofiles -H "Content-Type: application/json" -d '{"apiVersion":"firmware.management.io/v1","kind":"DeviceProfile","metadata":{"name":"node1"},"spec":{"manufacturer":"Dell","model":"R740","redfishPath":"/redfish/v1/UpdateService","managementIp":"192.168.1.10"}}'
  - curl -s -X POST http://localhost:8080/updateprofiles -H "Content-Type: application/json" -d '{"apiVersion":"firmware.management.io/v1","kind":"UpdateProfile","metadata":{"name":"redfish-profile"},"spec":{"commandType":"Redfish","payloadPath":"bios.bin","successCriteria":"TaskState=Completed"}}'
  - curl -s -X POST http://localhost:8080/firmwaremanifests -H "Content-Type: application/json" -d '{"apiVersion":"firmware.management.io/v1","kind":"FirmwareManifest","metadata":{"name":"bios-v2"},"spec":{"versionString":"2.0.0","versionNumber":"2","targetComponent":"BIOS","updateProfileRef":"redfish-profile"}}'
  - curl -s -X POST http://localhost:8080/updatejobs -H "Content-Type: application/json" -d '{"apiVersion":"firmware.management.io/v1","kind":"UpdateJob","metadata":{"name":"job1"},"spec":{"targetNode":"node1","targetComponent":"BIOS","firmwareRef":"bios-v2","dryRun":true}}'
  - curl -s -X POST http://localhost:8080/lookupjobs -H "Content-Type: application/json" -d '{"apiVersion":"firmware.management.io/v1","kind":"LookupJob","metadata":{"name":"lookup1"},"spec":{"targetNode":"node1"}}'

## Representative command outputs
Health check:
{
  "status":"healthy",
  "service":"fms"
}

Database init (log):
2026/04/15 15:32:28 Database schema initialized successfully

DeviceProfile creation (201 response excerpt):
{"apiVersion":"v1","kind":"DeviceProfile","metadata":{"name":"node1","uid":"deviceprofile-01dbd1a0","createdAt":"2026-04-15T15:42:52.09609-07:00"},"spec":{"manufacturer":"Dell","model":"R740","redfishPath":"/redfish/v1/UpdateService","managementIp":"192.168.1.10"},"status":{"ready":false}}

Build:
- bin/server produced successfully (go build completed with no errors)

## Notes and next steps
- Reconcilers are functional but lightweight; further production hardening required:
  - Retries, backoff, metrics, better error reporting
  - Tests (unit + integration) for reconcilers and SSH helper
  - Secure storage for credentials (avoid /tmp in production)
- Ent/SQLite currently used for local dev; consider Postgres for production ent storage.

## Exact user request fulfilled
"Can you add a README detailing what this service is, what's been done, what's tested, and show some output of the commands being run?"