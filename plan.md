# Firmware Management Service (FMS) Logic Implementation Plan

You are an autonomous agent tasked with implementing custom business logic for the Firmware Management Service. The base Fabrica framework is already scaffolded. You MUST execute this plan sequentially. **DO NOT move to a subsequent phase until you have successfully executed the validation step for the current phase.**

## Phase 1: Scaffold External Dependencies
**Goal:** Implement the provided reference snippets from `reference_snippets.md` into the `internal/` directory.
1. Create `internal/redfish/redfish.go` using the Redfish HTTP Client snippet.
2. Create `internal/ziphelper/ziphelper.go` using the ZIP Extraction snippet.
3. Create `internal/hms/hms.go` using the HMS Interface & Mock snippet.
4. **Validation:** Run `go mod tidy` and `go build ./...`. It must compile without errors.

## Phase 2: Implement FMLS Binary and ZIP Parsing
**Goal:** Implement the custom Library Service upload handler.
1. Inject `r.Post("/library/upload", libraryUploadHandler)` into `cmd/server/main.go` right before `http.ListenAndServe`.
2. Create `libraryUploadHandler` to accept a zip file, extract it using `ziphelper`, parse `manifest.json`, map it to `v1.FirmwareProfile`, and save it using `storage.SaveFirmwareProfile(r.Context(), profile)`.
3. Add `r.Get("/library/files/*", ...)` to serve the extracted `/tmp/firmware/` directory.
4. **Validation:** - Start the server (`CGO_ENABLED=1 go run ./cmd/server serve --database-url="file:data.db?cache=shared&_fk=1" &`).
   - Create a dummy zip with a dummy `manifest.json`. Use `curl -F "file=@dummy.zip" http://localhost:8080/library/upload`.
   - Run `curl http://localhost:8080/firmwareprofiles` and assert the database record was created.
   - Kill the server.

## Phase 3: Implement UpdateJob Reconciler (The Splitter)
**Goal:** Implement `pkg/reconcilers/updatejob_reconciler.go`.
1. Ensure idempotency by checking if `res.Status.Phase` is "Complete" or "Failed".
2. Iterate over `res.Spec.TargetNodes`. For each, check if an `UpdateTask` already exists using `r.Client.List(ctx, "UpdateTask")`.
3. If not, generate a UID and call `r.Client.Create(ctx, task)` to spawn the child task.
4. Aggregate child task states to update `res.Status.Phase` and persist via `r.UpdateStatus(ctx, res)`.
5. **Validation:**
   - Start the server.
   - POST an `UpdateJob` targeting `["nodeA", "nodeB"]`.
   - Wait 3 seconds.
   - Run `curl http://localhost:8080/updatetasks`. Assert that exactly two tasks were automatically created.
   - Kill the server.

## Phase 4: Implement UpdateTask Reconciler (Execution)
**Goal:** Implement `pkg/reconcilers/updatetask_reconciler.go`.
1. Ensure idempotency (return if "Success" or "Failed").
2. Initialize `hms.NewLocalHMS()` to get credentials.
3. Use `redfish.SendSecureRedfish` to POST to the node's `SimpleUpdate` endpoint using the `TargetNode` as the host.
4. Update `res.Status.TaskUri` and `res.Status.State` based on the Redfish response or connection error. Call `r.UpdateStatus(ctx, res)`.
5. **Validation:** - Start the server.
   - POST a single `UpdateTask` manually.
   - Wait 3 seconds.
   - GET the task. Assert that its state transitioned to "Failed" (since the target node won't actually exist to receive the Redfish call). This proves the execution loop triggered.
   - Kill the server.

## Phase 5: Implement LookupJob Reconciler
**Goal:** Implement `pkg/reconcilers/lookupjob_reconciler.go`.
1. Return early if `res.Status.State` is "Complete" or "Failed".
2. Search `r.Client.List(ctx, "DeviceProfile")` for the target node to get its IP, fallback to `hmsClient.GetDeviceFQDN()`.
3. Execute an HTTPS GET to the FirmwareInventory endpoint using `InsecureSkipVerify: true`.
4. Store the raw JSON into `res.Status.FirmwareData`, set State to "Complete" (or "Failed" if dial error), and persist via `r.UpdateStatus(ctx, res)`.
5. **Validation:**
   - Start the server.
   - POST a `LookupJob`. 
   - Wait 3 seconds.
   - Assert the state transitioned to "Failed" or "Complete".
   - Kill the server.

## Phase 6: Full System E2E Verification
**Goal:** Run the full integration test.
1. Ensure `cmd/server/main.go` has `_busy_timeout=10000` applied to the SQLite connection string to prevent concurrency locks.
2. Execute the `verify-e2e.sh` script located in the root directory. 
3. The script must report `✅` for all three logic phases. Fix any bugs indicated by the test script until it passes perfectly.