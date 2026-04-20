// Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT
// This file contains user-customizable reconciliation logic for UpdateTask.
//
// ⚠️ This file is safe to edit - it will NOT be overwritten by code generation.
package reconcilers

import (
"context"
"encoding/json"
"fmt"
"time"

v1 "github.com/bmcdonald3/fms/apis/example.fabrica.dev/v1"
"github.com/bmcdonald3/fms/internal/hms"
"github.com/bmcdonald3/fms/internal/redfish"
)

// reconcileUpdateTask drives a firmware update on the target node via Redfish SimpleUpdate.
//
// Idempotency: tasks already in a terminal state (Success/Failed) are skipped immediately.
// On connection failure to a non-existent BMC the task transitions to "Failed", which is
// the expected outcome in test environments without real hardware.
func (r *UpdateTaskReconciler) reconcileUpdateTask(ctx context.Context, res *v1.UpdateTask) error {
if res.Status.State == "Success" || res.Status.State == "Failed" {
return nil
}

hmsClient := hms.NewLocalHMS()
user, pass, err := hmsClient.GetDeviceCredentials(res.Spec.TargetNode)
if err != nil {
return fmt.Errorf("failed to get credentials for %s: %w", res.Spec.TargetNode, err)
}

host, err := hmsClient.GetDeviceFQDN(res.Spec.TargetNode)
if err != nil {
host = res.Spec.TargetNode
}

payload := fmt.Sprintf(`{"ImageURI":"%s","Targets":["%s"]}`, res.Spec.FirmwareRef, res.Spec.TargetComponent)

res.Status.State = "Running"
res.Status.StartTime = time.Now().Unix()

body, redfishErr := redfish.SendSecureRedfish(
host,
"/redfish/v1/UpdateService/Actions/SimpleUpdate",
payload,
user,
pass,
"POST",
10,
)

if redfishErr != nil {
r.Logger.Warnf("Redfish call failed for node %s: %v", res.Spec.TargetNode, redfishErr)
res.Status.State = "Failed"
} else {
// Extract the TaskUri from the response body if present.
var respData map[string]interface{}
if err := json.Unmarshal([]byte(body), &respData); err == nil {
if taskURI, ok := redfish.GetNestedValue(respData, "TaskMonitor"); ok {
if s, ok := taskURI.(string); ok {
res.Status.TaskUri = s
}
}
}
res.Status.State = "Success"
r.Logger.Infof("Redfish SimpleUpdate succeeded for node %s", res.Spec.TargetNode)
}

if err := r.Client.Update(ctx, res); err != nil {
return fmt.Errorf("failed to persist UpdateTask status: %w", err)
}

return nil
}