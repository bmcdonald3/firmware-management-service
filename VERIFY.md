# Verification Steps

1. Start the server in the background: go run ./cmd/server/

2. Create a DeviceProfile:

curl -X POST http://localhost:8080/device-profiles -H "Content-Type: application/json" -d '{"spec": {"profileName": "iLO", "protocol": "redfish", "verification": {"path": "/redfish/v1/UpdateService/FirmwareInventory/0", "filter": ".Name", "value": "iLO*"}, "variables": [{"name": "model", "path": "/redfish/v1/Chassis/1", "filter": ".Model"}]}}'

3. Create an UpdateProfile:

curl -X POST http://localhost:8080/update-profiles -H "Content-Type: application/json" -d '{"spec": {"profileName": "iLO", "protocol": "redfish", "pushPull": "pull", "updatePath": "/redfish/v1/UpdateService/Actions/SimpleUpdate", "payload": "{\"ImageURI\": \"%httpFileName%\"}", "firmwareInventory": "/redfish/v1/UpdateService/FirmwareInventory", "defaultTimeout": 300}}'

4. Create a FirmwareProfile:

curl -X POST http://localhost:8080/firmware-profiles -H "Content-Type: application/json" -d '{"spec": {"deviceProfile": "iLO", "updateProfile": "iLO", "targets": ["System ROM"], "firmwareVersion": "A47 v3.70", "semanticFirmwareVersion": "3.70.0", "fileName": "A47_3.70.flash", "models": ["ProLiant XL675d"]}}'

5. Initiate an UpdateJob:

curl -X POST http://localhost:8080/update-jobs -H "Content-Type: application/json" -d '{"spec": {"targetNode": "node-123", "firmwareProfileId": "A47_3.70.flash", "dryRun": false}}'

6. Check Reconciliation Status:

curl http://localhost:8080/update-jobs | jq

Success Criteria: The status.state field should transition away from an empty or initial state, indicating the reconciler processed the job.