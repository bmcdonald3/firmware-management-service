# Firmware Management Service (FMS)

## Overview
FMS is a lightweight controller designed to manage firmware updates and inventory for devices via **Redfish** and **SSH**. It exposes a CRUD HTTP API (scaffolded via Fabrica) for managing lifecycle resources and utilizes dedicated reconcilers to orchestrate update and lookup workflows.

---

## Architecture & Implementation

### API and Framework
* **Fabrica Layers:** Implemented handlers, routes, OpenAPI specifications, and `ent` storage.
    * `fms/.fabrica.yaml`
    * `fms/apis.yaml`
    * `fms/pkg/apiversion/registry_generated.go`
* **Resource Types:** Spec/Status definitions with validation.
    * `DeviceProfile`, `UpdateProfile`, `FirmwareManifest`, `UpdateJob`, `LookupJob`

### Core Components
* **Reconcilers:**
    * `UpdateJob`: Manages a state machine (Preflight → Execute → Monitor).
    * `LookupJob`: Collects firmware inventory via Redfish endpoints.
    * `SSH Helper`: Facilitates SSH/SCP operations using `golang.org/x/crypto/ssh`.
* **Firmware Library (FMLS):** A file-based server for managing binaries under `/tmp/firmware`.
* **Storage:** Integrated `ent` schema with SQLite adapters and DB migration logic.
* **Auth:** Credential parsing from `/tmp/credentials.json`.

---

## Build and Execution

### Local Setup
```bash
# Build the binary
cd fms && go build -o bin/server ./cmd/server

# Initialize the database schema
cd fms && go run ./cmd/server init-db

# Start the service
FMS_HOST_IP=127.0.0.1 ./bin/server
```

### Smoke Testing
```bash
# Health Check
curl -s http://localhost:8080/health

# Create a Device Profile
curl -s -X POST http://localhost:8080/deviceprofiles \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "firmware.management.io/v1",
    "kind": "DeviceProfile",
    "metadata": {"name": "node1"},
    "spec": {
      "manufacturer": "Dell",
      "model": "R740",
      "redfishPath": "/redfish/v1/UpdateService",
      "managementIp": "192.168.1.10"
    }
  }'
```

---

## Verification Status

| Component | Status | Notes |
| :--- | :--- | :--- |
| **Toolchain** | Pass | `fabrica_init`, `add_resource`, and `generate` verified. |
| **Build** | Pass | `go mod tidy` and binary compilation successful. |
| **Database** | Pass | Migrations and schema initialization verified. |
| **API** | Pass | End-to-end CRUD functional for all resources. |

### Representative Logs

**Database Initialization:**
> `2026/04/15 15:32:28 Database schema initialized successfully`

**Health Check Response:**
```json
{
  "status": "healthy",
  "service": "fms"
}
```

---

## Roadmap & Production Hardening

### Resilience
* Implement exponential backoff and retry logic in reconcilers.
* Add Prometheus metrics for job latency and failure rates.
* Increase unit and integration test coverage for `ssh_helper.go`.

### Security & Infrastructure
* **Secret Management:** Transition from `/tmp/credentials.json` to a secure provider (e.g., Vault or K8s Secrets).
* **Database:** Migrate from SQLite to **PostgreSQL** for production persistence.
* **Storage:** Enhance FMLS to support S3-compatible backends instead of local `/tmp` storage.