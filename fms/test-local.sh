#!/bin/bash
SERVER_PID=$!
sleep 5
curl -s -X POST http://localhost:8081/deviceprofiles -H "Content-Type: application/json" -d '{"metadata": {"name": "172.24.0.3"}, "spec": {"managementIp": "172.24.0.3", "manufacturer": "generic", "model": "RootService", "redfishPath": "/redfish/v1"}}' > /dev/null
curl -s -X POST http://localhost:8081/firmwareprofiles -H "Content-Type: application/json" -d '{"metadata": {"name": "generic-fw"}, "spec": {"versionString": "9.9.9", "versionNumber": "999", "targetComponent": "BMC"}}' > /dev/null
curl -s -X POST http://localhost:8081/updatejobs -H "Content-Type: application/json" -d '{"metadata": {"name": "e2e-job-1"}, "spec": {"targetNodes": ["172.24.0.3"], "targetComponent": "BMC", "firmwareRef": "generic-fw", "force": true, "dryRun": true}}' > /dev/null
for i in {1..6}; do
    echo "Task State (Poll $i):"
    curl -s http://localhost:8081/updatetasks
    echo ""
    sleep 2
done
echo "Relevant Reconciler Logs:"
grep "Reconciling\|UpdateTask\|UpdateJob" test_server.log | tail -n 15
kill -9 $SERVER_PID