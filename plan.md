# Firmware Management Service (FMS) Implementation Plan

You are an autonomous agent tasked with building the Firmware Management Service. You must execute this plan sequentially. Do not move to the next step until the current step's success criteria are met. 

## Step 1: Initialize Project and Generate Resources via MCP
**Description:** Use the Fabrica MCP server tools to bootstrap the project. 
1. Call `fabrica_init` to create a project named `fms`. 
   - Group: `firmware.management.io`
   - Storage Type: `ent`
   - DB: `sqlite`
   - Enable reconciliation.
2. Change your working directory to the newly created `fms` folder for all subsequent commands.
3. Call `fabrica_add_resource` to add the following 6 resources: `DeviceProfile`, `FirmwareProfile`, `UpdateProfile`, `UpdateJob`, `UpdateTask`, `LookupJob`.
4. Modify the generated `*_types.go` files in `apis/firmware.management.io/v1/` to include these exact fields with `json` and `validate` tags:
   - **DeviceProfileSpec:** `manufacturer` (req), `model` (req), `redfishPath` (req), `managementIp` (req).
   - **FirmwareProfileSpec:** `versionString` (req), `versionNumber` (req), `targetComponent` (req), `preConditions`, `postConditions`, `softwareId`.
   - **UpdateProfileSpec:** `commandType` (req, oneof=Redfish SSH), `payloadPath`, `successCriteria`.
   - **UpdateJobSpec:** `targetNodes` (array of strings), `targetComponent`, `firmwareRef`, `dryRun` (bool), `force` (bool).
   - **UpdateTaskSpec:** `targetNode`, `targetComponent`, `firmwareRef`, `updateJobId`.
   - **UpdateTaskStatus:** `state` (oneof=Pending Running Success Failed), `startTime` (int64), `taskUri`.
   - **LookupJobSpec:** `targetNode`.
   - **LookupJobStatus:** `state` (oneof=Pending Running Complete Failed), `jobId`, `firmwareData`.
5. Call `fabrica_generate`.
6. Run `go mod tidy` and `go run ./cmd/server init-db`.
7. **Framework Guardrail:** Open `.fabrica.yaml` and `cmd/server/main.go` to ensure API authentication (like Tokensmith) is disabled or bypassed for local development.

## Step 2: Implement FMLS Binary and ZIP Parsing
**Description:** Implement the Library Service capabilities.
1. **Framework Guardrail (Routing):** Do NOT modify `routes_generated.go`. Instead, open `cmd/server/main.go` and inject a custom HTTP POST handler for `/library/upload` directly onto the `chi.Router` before `http.ListenAndServe` is called.
2. Accept a `multipart/form-data` ZIP file upload.
3. Using the provided `zipfiles.go` reference logic, extract the ZIP to `/tmp/firmware/`, parse the internal manifest JSON, and map it to a `FirmwareProfile` struct.
4. Use `r.Client.Create(ctx, profile)` to automatically generate the database record.
5. Add a GET handler to serve the extracted `.bin` files from `/tmp/firmware/`.

## Step 3: Implement UpdateJob Reconciler (The Splitter)
**Description:** Write the reconciliation loop in `pkg/reconcilers/updatejob_reconciler.go`.
1. Fetch the incoming `UpdateJob`.
2. Iterate over the `spec.targetNodes` array.
3. **Framework Guardrail (Filtering):** Fabrica's `r.Client.List` does not natively filter by struct fields. You must call `items, err := r.Client.List(ctx, "UpdateTask")` and manually iterate through `items` in memory to check if an `UpdateTask` already exists for this specific `updateJobId` and `targetNode`.
4. If it does not exist, use `r.Client.Create()` to instantiate a new `UpdateTask` resource.
5. Aggregate the status of all child `UpdateTask` resources to determine and update the parent `UpdateJob` status. Call `r.Client.Update(ctx, job)` to persist.

## Step 4: Implement UpdateTask Reconciler (Execution & HMS)
**Description:** Write the reconciliation loop in `pkg/reconcilers/updatetask_reconciler.go`. Ensure all external network calls are wrapped in contexts with strict timeouts (e.g., 30s).
1. **HMS Lock:** Using the provided `hmsInterface.go` logic, execute an HTTP GET to lock the target node. If locked by another service, return `RequeueAfter: 30 * time.Second`.
2. **Pre-flight:** Check `spec.force`. If false, query the node for its current firmware version. If it matches the target version, set State to `Success`, update DB, and halt.
3. **Execution:** Using the provided `redfish.go` reference logic, execute a POST to the node's `SimpleUpdate` endpoint. Pass `{"ImageURI": "http://<FMS_HOST_IP>/library/<firmware-file>"}`. Extract the `Location` header from the response, save it to `Status.TaskURI`, set State to `Running`, and update DB.
4. **Monitoring:** If State is `Running`, GET `Status.TaskURI`. If the task is complete, set State to `Success`. If exception/timeout, set to `Failed`. Persist state changes.

## Step 5: Implement LookupJob Reconciler
**Description:** Write the reconciliation loop in `pkg/reconcilers/lookupjob_reconciler.go`.
1. Fetch the `DeviceProfile` mapped to the `targetNode`.
2. Execute an HTTP GET to `https://<managementIp>/redfish/v1/UpdateService/FirmwareInventory` using a custom `http.Transport` with `InsecureSkipVerify: true`.
3. Store the raw JSON response payload into `Status.FirmwareData`, set State to `Complete`, and explicitly persist by calling `r.Client.Update(ctx, job)`.

## Step 6: System Integration and Validation
**Description:** Verify the system compiles and operates.
1. **Framework Guardrail (SQLite Concurrency):** Open `cmd/server/main.go` or `.fabrica.yaml` (wherever the SQLite connection string is defined) and append `?_busy_timeout=10000` to the database URL to prevent `database is locked` panics during concurrent reconciler writes.
2. Execute `go mod tidy`.
3. Build the server using `go build -o bin/server ./cmd/server`.
4. Start the server.
5. Use `curl` to submit a mock ZIP bundle to the `/library/upload` endpoint. Verify the DB record is created.
6. Use `curl` to submit an `UpdateJob` targeting multiple nodes. Verify the log output shows the splitter generating child tasks and the task execution loop engaging.