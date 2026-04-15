[
    {
        "id": "step_1_init_project",
        "name": "Initialize Fabrica Project via MCP",
        "description": "Use the fabrica_init MCP tool on the fabrica MCP server to create a project named 'fms'. Set the parameters to use the 'firmware.management.io' API group, 'ent' storage-type, 'sqlite' database, and enable the reconciliation framework. Use an absolute path to the target directory. Ensure the working_dir parameter is changed to the newly created 'fms' subdirectory for all subsequent tool calls.",
        "requires_code": true,
        "expected_outputs": [
            "Project directory structure",
            ".fabrica.yaml configuration file",
            "Root apis.yaml file"
        ],
        "success_criteria": [
            "MCP tool execution returns success",
            "Project configuration reflects SQLite and reconciliation enabled"
        ]
    },
    {
        "id": "step_2_define_resources",
        "name": "Generate API Resources via MCP",
        "description": "Use the fabrica_add_resource MCP tool (or equivalent Fabrica MCP tool for adding resources) to add the 5 core resources to the working_dir: DeviceProfile, FirmwareManifest, UpdateProfile, UpdateJob, and LookupJob.",
        "requires_code": true,
        "expected_outputs": [
            "Generated resource struct files in apis/firmware.management.io/v1/"
        ],
        "success_criteria": [
            "All 5 resource definition files exist in the API directory"
        ]
    },
    {
        "id": "step_3_configure_structs",
        "name": "Configure Resource Structs and Validations",
        "description": "Modify the generated *_types.go files to strictly match these definitions. DeviceProfileSpec: Manufacturer (string `json:\"manufacturer\" validate:\"required\"`), Model (string `json:\"model\" validate:\"required\"`), RedfishPath (string `json:\"redfishPath\" validate:\"required\"`), ManagementIP (string `json:\"managementIp\" validate:\"required\"`). FirmwareManifestSpec: VersionString (string `json:\"versionString\" validate:\"required\"`), VersionNumber (string `json:\"versionNumber\" validate:\"required\"`), TargetComponent (string `json:\"targetComponent\" validate:\"required\"`), PreConditions (string `json:\"preConditions\"`), PostConditions (string `json:\"postConditions\"`), SoftwareID (string `json:\"softwareId\"`), UpdateProfileRef (string `json:\"updateProfileRef\" validate:\"required\"`). UpdateProfileSpec: CommandType (string `json:\"commandType\" validate:\"required,oneof=Redfish SSH\"`), PayloadPath (string `json:\"payloadPath\"`), SuccessCriteria (string `json:\"successCriteria\"`). UpdateJobSpec: TargetNode (string `json:\"targetNode\" validate:\"required\"`), TargetComponent (string `json:\"targetComponent\" validate:\"required\"`), FirmwareRef (string `json:\"firmwareRef\" validate:\"required\"`), DryRun (bool `json:\"dryRun\"`), Force (bool `json:\"force\"`). UpdateJobStatus: State (string `json:\"state\" validate:\"oneof=Pending Running Success Failed\"`), StartTime (int64 `json:\"startTime\"`), TimeoutLimitSeconds (int `json:\"timeoutLimitSeconds\"`), TaskURI (string `json:\"taskUri\"`), JobID (string `json:\"jobId\"`). LookupJobSpec: TargetNode (string `json:\"targetNode\" validate:\"required\"`). LookupJobStatus: State (string `json:\"state\" validate:\"oneof=Pending Running Complete Failed\"`), JobID (string `json:\"jobId\"`), FirmwareData (string `json:\"firmwareData\"`).",
        "requires_code": true,
        "expected_outputs": [
            "Updated Go struct definitions",
            "Struct tags applied for validation"
        ],
        "success_criteria": [
            "Code compiles without errors",
            "Validation rules strictly enforce the defined enumerations"
        ]
    },
    {
        "id": "step_4_generate_layer",
        "name": "Generate API and Storage Layers via MCP",
        "description": "Use the fabrica_generate MCP tool on the fabrica MCP server to generate the API handlers, SQLite storage layer functions, and OpenAPI specifications. Afterwards, execute a standard shell command to run `go mod tidy`.",
        "requires_code": true,
        "expected_outputs": [
            "Generated handler code",
            "Generated SQLite ent schema and storage functions",
            "OpenAPI documentation"
        ],
        "success_criteria": [
            "MCP generation completes without errors",
            "Dependencies are cleanly resolved"
        ]
    },
    {
        "id": "step_5_database_init",
        "name": "Initialize Database Schema",
        "description": "Run the initial database setup command in the shell to create the required SQLite tables for the generated resources. Execute: go run ./cmd/server init-db",
        "requires_code": true,
        "expected_outputs": [
            "Local SQLite database file created",
            "Schema tables instantiated for all API resources"
        ],
        "success_criteria": [
            "Database file exists and is accessible",
            "Tables match the defined structs"
        ]
    },
    {
        "id": "step_6_implement_fmls",
        "name": "Implement Firmware Library and Credentials Reader",
        "description": "Create pkg/fmls/server.go and pkg/credentials/parser.go. Create a file at '/tmp/credentials.json' matching this schema: `{\"nodes\": {\"node1\": {\"username\": \"root\", \"password\": \"secret\"}}}`. Create a parser to retrieve credentials by node name. Write an HTTP server exposing `/library/` with a GET handler to serve binaries from `/tmp/firmware`, a POST handler (multipart/form-data) to upload new binaries, and a DELETE handler.",
        "requires_code": true,
        "expected_outputs": [
            "Credentials parser package",
            "Firmware binary file server package"
        ],
        "success_criteria": [
            "Parser returns correct credentials for 'node1'",
            "HTTP server correctly writes a new file via POST and removes it via DELETE"
        ]
    },
    {
        "id": "step_7_implement_fmus_reconciler",
        "name": "Implement UpdateJob Reconciler Logic",
        "description": "Write reconciliation loop in pkg/reconcilers/updatejob_reconciler.go. Step 0 (Load Relational Data): Use `r.Client.Get(ctx, Kind, UID)` to fetch the DeviceProfile, FirmwareManifest, and UpdateProfile mapped to this job. Step 1 (Locking): Use `r.Client.List(ctx, \"UpdateJob\")`. If a conflicting running job exists, return requeue 5s. Step 2 (Dry Run): If `spec.DryRun`, set `Status.State = \"Success\"`, call `r.Client.Update()`, and halt. Step 3 (Pre-flight): Use HTTP with a 15-second timeout context. If `spec.Force` is false, check current version. If match, set `Status.State = \"Success\"`, update, and halt. Evaluate `PreConditions` (e.g. power state); if not met, requeue 10s. Step 4 (Execution): If `Status.TaskURI` is empty, read `os.Getenv(\"FMS_HOST_IP\")`. Switch on `CommandType`. For 'Redfish', POST to ManagementIP with 30s timeout context. Parse Task Monitor URI from Location header into `Status.TaskURI`. For 'SSH', apply 60s timeout context for SCP and execution. Set `Status.State = \"Running\"` and `Status.StartTime`, then call `r.Client.Update(ctx, job)`. Step 5 (Monitoring): If `time.Now().Unix() - Status.StartTime > Status.TimeoutLimitSeconds`, set Failed, update, and halt. Otherwise, GET TaskURI. Update State to Success/Failed based on payload, call `r.Client.Update(ctx, job)`. Requeue if still running.",
        "requires_code": true,
        "expected_outputs": [
            "Reconciliation logic handling state machine transitions with explicit DB updates",
            "Context timeouts applied to all network calls"
        ],
        "success_criteria": [
            "Database state correctly persists across reconciliation loops",
            "Missing network endpoints do not hang the worker pool"
        ]
    },
    {
        "id": "step_8_implement_lookupjob_reconciler",
        "name": "Implement LookupJob Reconciler Logic",
        "description": "Write the reconciliation loop for the LookupJob resource. Step 0: Fetch the DeviceProfile via `r.Client.Get(ctx, \"DeviceProfile\", job.Spec.TargetNode)`. Step 1: Initiate HTTP GET to `https://<DeviceProfile.ManagementIP>/redfish/v1/UpdateService/FirmwareInventory` using a 15-second timeout context and `InsecureSkipVerify: true`. Step 2: Store JSON into `Status.FirmwareData`, set `Status.State = \"Complete\"`, and explicitly persist by calling `r.Client.Update(ctx, job)`.",
        "requires_code": true,
        "expected_outputs": [
            "Reconciliation logic with database updates and context timeouts"
        ],
        "success_criteria": [
            "Lookup queries execute safely and persist correctly to SQLite"
        ]
    },
    {
        "id": "step_9_validation",
        "name": "System Integration and API Testing",
        "description": "Execute `go get golang.org/x/crypto/ssh` and `go mod tidy` to resolve dependencies. Build the server: `FMS_HOST_IP=127.0.0.1 go build -o bin/server ./cmd/server`. Run `./bin/server`. Use a REST client to submit mock records for all 5 resources. Verify logs and database states.",
        "requires_code": true,
        "expected_outputs": [
            "Resolved dependency tree",
            "Running server binary",
            "Successful API integration tests"
        ],
        "success_criteria": [
            "Build completes with zero missing module errors",
            "UpdateJob traverses full state machine from Pending to Success/Failed"
        ]
    }
]