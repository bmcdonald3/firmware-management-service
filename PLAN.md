# Firmware Management Service (FMS) Demo Plan

1. **Initialize Project**: Use the `fabrica_init` MCP tool on the `fabrica` MCP server to create `fms`.
   - Enable metrics, events, and reconciliation.
   - Use absolute path to the target directory (e.g., `/Users/username/demo-fabrica/fms`).
   - The `working_dir` parameter MUST change to the newly created `fms` subdirectory for all subsequent steps.

2. **Add FMLS Resources**: Use the MCP tool to add the library resources.
   - Resource 1: `DeviceProfile`. Fields in Spec: `profileName`, `version`, `deviceProfileVersion`, `protocol`, `verification` (struct/map), `variables` (slice).
   - Resource 2: `UpdateProfile`. Fields in Spec: `profileName`, `version`, `updateProfileVersion`, `protocol`, `pushPull`, `updatePath`, `payload`, `firmwareInventory`, `firmwareInventoryExpand`, `defaultTimeout`.
   - Resource 3: `FirmwareProfile`. Fields in Spec: `deviceProfile`, `updateProfile`, `targets`, `firmwareVersion`, `semanticFirmwareVersion`, `fileName`, `timeout`, `models`, `softwareIds`, `pollingSpeed`.

3. **Add FMUS Resource**: Use the MCP tool to add an `UpdateJob` resource to track node updates.
   - Fields in Spec: `targetNode`, `firmwareProfileId`, `dryRun`.
   - Fields in Status: `state` (e.g., pending, running, complete, failed), `taskId`, `message`, `startTime`, `endTime`.

4. **Generate**: Use the `fabrica_generate` MCP tool and run `go mod tidy`.

5. **Implement Reconciliation**: Modify the generated reconciler for `UpdateJob`.
   - Implement logic to read the requested `FirmwareProfile`, look up the associated `DeviceProfile` and `UpdateProfile`.
   - Stub the out-of-band communication logic (e.g., log the intended Redfish payload and path).
   - Update `Status.State` to simulate the job lifecycle transitioning from "pending" to "complete".

6. **Verify**: Follow the test plan in `VERIFY.md`.