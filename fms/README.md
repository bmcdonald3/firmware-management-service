# Firmware Management Service (FMS) Demo

## Overview

The Firmware Management Service (FMS) is a demo implementation of a declarative, API-driven firmware management system for infrastructure environments. It is built using the [Fabrica](https://github.com/openchami/fabrica) framework and demonstrates resource modeling, reconciliation, and lifecycle management for firmware updates.

## Architecture

- **Fabrica-based API server** with CRUD endpoints for DeviceProfile, UpdateProfile, FirmwareProfile, and UpdateJob resources.
- **Reconciliation framework**: UpdateJob resources are processed by a custom reconciler that simulates the firmware update lifecycle.
- **File-based storage** for all resources.
- **Metrics** (on port 9090) and **event bus** enabled.
- **OpenAPI/Swagger** documentation auto-generated.

## Resource Model

- **DeviceProfile**: Describes device characteristics and verification logic.
- **UpdateProfile**: Describes update protocol, payload, and inventory paths.
- **FirmwareProfile**: Binds device and update profiles to firmware artifacts.
- **UpdateJob**: Tracks the execution and lifecycle of a firmware update.

## Usage

### Prerequisites

- Go 1.24+
- [Fabrica CLI](https://github.com/openchami/fabrica)
- curl, jq

### Running the Server

```sh
cd fms
go run ./cmd/server/
```

Server starts on `http://localhost:8080`.

### API Endpoints

- `POST /deviceprofiles`
- `POST /updateprofiles`
- `POST /firmwareprofiles`
- `POST /updatejobs`
- `GET /updatejobs`

See [VERIFY.md](../VERIFY.md) for full test plan.

### Example Workflow

#### 1. Create DeviceProfile

```sh
curl -X POST http://localhost:8080/deviceprofiles -H "Content-Type: application/json" -d '{
  "spec": {
    "profileName": "iLO",
    "protocol": "redfish",
    "verification": {
      "path": "/redfish/v1/UpdateService/FirmwareInventory/0",
      "filter": ".Name",
      "value": "iLO*"
    },
    "variables": [
      {"name": "model", "path": "/redfish/v1/Chassis/1", "filter": ".Model"}
    ]
  }
}'
```

**Output:**
```json
{"apiVersion":"v1","kind":"DeviceProfile","metadata":{"name":"","uid":"deviceprofile-680c97ae",...},"spec":{...},"status":{"ready":false}}
```

#### 2. Create UpdateProfile

```sh
curl -X POST http://localhost:8080/updateprofiles -H "Content-Type: application/json" -d '{
  "spec": {
    "profileName": "iLO",
    "protocol": "redfish",
    "pushPull": "pull",
    "updatePath": "/redfish/v1/UpdateService/Actions/SimpleUpdate",
    "payload": "{\"ImageURI\": \"%httpFileName%\"}",
    "firmwareInventory": "/redfish/v1/UpdateService/FirmwareInventory",
    "defaultTimeout": 300
  }
}'
```

**Output:**
```json
{"apiVersion":"v1","kind":"UpdateProfile","metadata":{"name":"","uid":"updateprofile-0d68a893",...},"spec":{...},"status":{"ready":false}}
```

#### 3. Create FirmwareProfile

```sh
curl -X POST http://localhost:8080/firmwareprofiles -H "Content-Type: application/json" -d '{
  "spec": {
    "deviceProfile": "iLO",
    "updateProfile": "iLO",
    "targets": ["System ROM"],
    "firmwareVersion": "A47 v3.70",
    "semanticFirmwareVersion": "3.70.0",
    "fileName": "A47_3.70.flash",
    "models": ["ProLiant XL675d"]
  }
}'
```

**Output:**
```json
{"apiVersion":"v1","kind":"FirmwareProfile","metadata":{"name":"","uid":"firmwareprofile-0b2cf39e",...},"spec":{...},"status":{"ready":false}}
```

#### 4. Initiate UpdateJob

```sh
curl -X POST http://localhost:8080/updatejobs -H "Content-Type: application/json" -d '{
  "spec": {
    "targetNode": "node-123",
    "firmwareProfileId": "A47_3.70.flash",
    "dryRun": false
  }
}'
```

**Output:**
```json
{"apiVersion":"v1","kind":"UpdateJob","metadata":{"name":"","uid":"updatejob-afaddcff",...},"spec":{...},"status":{}}
```

#### 5. Check UpdateJob Status

```sh
curl http://localhost:8080/updatejobs | jq
```

**Output:**
```json
[
  {
    "apiVersion": "v1",
    "kind": "UpdateJob",
    "metadata": {...},
    "spec": {
      "targetNode": "node-123",
      "firmwareProfileId": "A47_3.70.flash"
    },
    "status": {
      "state": "complete",
      "message": "Job completed successfully",
      "startTime": "2026-04-13T15:03:16-07:00",
      "endTime": "2026-04-13T15:03:18-07:00"
    }
  }
]
```

## What Is Implemented

- All resource types and OpenAPI endpoints
- Custom reconciliation logic for UpdateJob (simulated job lifecycle)
- Metrics and event bus
- File-based storage
- Full test plan from VERIFY.md

## What Is Left to Implement

- Real device and firmware inventory integration
- Real Redfish protocol communication
- Authentication/authorization
- Production-grade error handling and validation
- UI/dashboard

## Tests Run

All steps from [VERIFY.md](../VERIFY.md) were executed:

- Server startup
- DeviceProfile, UpdateProfile, FirmwareProfile, UpdateJob creation
- UpdateJob reconciliation and status transition

## Actual Command Output

See above for example outputs. The final UpdateJob status:

```json
{
  "state": "complete",
  "message": "Job completed successfully",
  "startTime": "2026-04-13T15:03:16-07:00",
  "endTime": "2026-04-13T15:03:18-07:00"
}
```

## How It Works

- Users POST resource specs to the API.
- The reconciler processes UpdateJobs, simulating the firmware update lifecycle.
- Status fields are updated to reflect job progress.
- All data is stored in the local file backend.

## Development

- Code generation: `fabrica generate`
- Dependency management: `go mod tidy`
- Server: `go run ./cmd/server/`
- All custom logic in `pkg/reconcilers/updatejob_reconciler.go`

## License

SPDX-License-Identifier: MIT