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
"os"
"time"

v1 "github.com/bmcdonald3/fms/apis/firmware.management.io/v1"
"github.com/bmcdonald3/fms/internal/hms"
"github.com/bmcdonald3/fms/internal/redfish"
)

const (
redfishTimeout    = 30
fmsHostEnv        = "FMS_HOST_IP"
defaultFMSHost    = "localhost:8080"
simpleUpdatePath  = "/redfish/v1/UpdateService/SimpleUpdate"
firmwareInventory = "/redfish/v1/UpdateService/FirmwareInventory"
)

// reconcileUpdateTask drives a single firmware update against one BMC node.
// State machine: (empty/Pending) → pre-flight → Running → Success/Failed.
func (r *UpdateTaskReconciler) reconcileUpdateTask(ctx context.Context, res *v1.UpdateTask) error {
r.Logger.Infof("Reconciling UpdateTask %s state=%q node=%s", res.GetUID(), res.Status.State, res.Spec.TargetNode)

hmsClient := hms.NewLocalHMS()
user, pass, err := hmsClient.GetDeviceCredentials(res.Spec.TargetNode)
if err != nil {
return fmt.Errorf("failed to get credentials for %s: %w", res.Spec.TargetNode, err)
}

// Resolve management IP from DeviceProfile, falling back to HMS FQDN.
managementIP, err := r.resolveManagementIP(ctx, res.Spec.TargetNode, hmsClient)
if err != nil {
r.Logger.Warnf("Could not resolve management IP for %s: %v — using node name directly", res.Spec.TargetNode, err)
managementIP = res.Spec.TargetNode
}

switch res.Status.State {
case "", "Pending":
return r.executePreflight(ctx, res, managementIP, user, pass)
case "Running":
return r.monitorTask(ctx, res, managementIP, user, pass)
default:
// Success or Failed — nothing to do.
return nil
}
}

// resolveManagementIP looks up the DeviceProfile for targetNode and returns its
// ManagementIp. It falls back to HMS FQDN resolution when no DeviceProfile exists.
func (r *UpdateTaskReconciler) resolveManagementIP(ctx context.Context, targetNode string, hmsClient *hms.LocalHMS) (string, error) {
items, err := r.Client.List(ctx, "DeviceProfile")
if err != nil {
return hmsClient.GetDeviceFQDN(targetNode)
}
for _, item := range items {
dp, ok := item.(*v1.DeviceProfile)
if !ok {
continue
}
if dp.Metadata.Name == targetNode || dp.Spec.ManagementIp == targetNode {
return dp.Spec.ManagementIp, nil
}
}
return hmsClient.GetDeviceFQDN(targetNode)
}

// executePreflight checks the current firmware version unless force=true,
// then triggers a Redfish SimpleUpdate.
func (r *UpdateTaskReconciler) executePreflight(ctx context.Context, res *v1.UpdateTask, ip, user, pass string) error {
// Fetch the FirmwareProfile to know the target version.
var targetVersion string
if res.Spec.FirmwareRef != "" {
raw, err := r.Client.Get(ctx, "FirmwareProfile", res.Spec.FirmwareRef)
if err == nil {
if fp, ok := raw.(*v1.FirmwareProfile); ok {
targetVersion = fp.Spec.VersionString
}
}
}

// Pre-flight: query current inventory if not forcing.
// We get the parent UpdateJob to check the force flag.
force := false
if res.Spec.UpdateJobId != "" {
raw, err := r.Client.Get(ctx, "UpdateJob", res.Spec.UpdateJobId)
if err == nil {
if job, ok := raw.(*v1.UpdateJob); ok {
force = job.Spec.Force
}
}
}

if !force && targetVersion != "" {
body, err := redfish.SendSecureRedfish(ip, firmwareInventory, "", user, pass, "GET", redfishTimeout)
if err == nil {
if currentVersionMatchesTarget(body, targetVersion, res.Spec.TargetComponent) {
r.Logger.Infof("UpdateTask %s: firmware already at target version %s — marking Success", res.GetUID(), targetVersion)
res.Status.State = "Success"
res.Status.StartTime = time.Now().Unix()
return r.Client.Update(ctx, res)
}
} else {
r.Logger.Warnf("Pre-flight inventory check failed for %s: %v — proceeding with update", ip, err)
}
}

// --- Execute Redfish SimpleUpdate ---
fmsHost := os.Getenv(fmsHostEnv)
if fmsHost == "" {
fmsHost = defaultFMSHost
}
imageURI := fmt.Sprintf("http://%s/library/%s", fmsHost, res.Spec.FirmwareRef)
payload := fmt.Sprintf(`{"ImageURI":%q}`, imageURI)

taskURI, err := redfish.GetLocationHeader(ip, simpleUpdatePath, payload, user, pass, redfishTimeout)
if err != nil {
res.Status.State = "Failed"
if updateErr := r.Client.Update(ctx, res); updateErr != nil {
r.Logger.Errorf("Failed to persist Failed state for UpdateTask %s: %v", res.GetUID(), updateErr)
}
return fmt.Errorf("SimpleUpdate POST failed for %s: %w", ip, err)
}

res.Status.TaskUri = taskURI
res.Status.State = "Running"
res.Status.StartTime = time.Now().Unix()
r.Logger.Infof("UpdateTask %s: SimpleUpdate accepted, taskURI=%s", res.GetUID(), taskURI)
return r.Client.Update(ctx, res)
}

// monitorTask polls the Redfish task URI and transitions to Success or Failed.
func (r *UpdateTaskReconciler) monitorTask(ctx context.Context, res *v1.UpdateTask, ip, user, pass string) error {
if res.Status.TaskUri == "" {
r.Logger.Warnf("UpdateTask %s is Running but has no TaskUri — resetting to Pending", res.GetUID())
res.Status.State = "Pending"
return r.Client.Update(ctx, res)
}

body, err := redfish.SendSecureRedfish(ip, res.Status.TaskUri, "", user, pass, "GET", redfishTimeout)
if err != nil {
r.Logger.Warnf("UpdateTask %s: task poll failed: %v — will retry", res.GetUID(), err)
// Transient error: return error to trigger requeue with backoff.
return err
}

var taskData map[string]interface{}
if err := json.Unmarshal([]byte(body), &taskData); err != nil {
return fmt.Errorf("failed to parse task response: %w", err)
}

taskState, _ := redfish.GetNestedValue(taskData, "TaskState")
taskStatus, _ := redfish.GetNestedValue(taskData, "TaskStatus")

r.Logger.Infof("UpdateTask %s poll: TaskState=%v TaskStatus=%v", res.GetUID(), taskState, taskStatus)

switch fmt.Sprintf("%v", taskState) {
case "Completed":
res.Status.State = "Success"
case "Exception", "Killed", "Interrupted":
res.Status.State = "Failed"
// All other states (Running, New, Pending, etc.) remain Running.
}

if res.Status.State != "Running" {
return r.Client.Update(ctx, res)
}
return nil
}

// currentVersionMatchesTarget parses a FirmwareInventory response body and
// checks whether any member's Version matches targetVersion for the given component.
func currentVersionMatchesTarget(inventoryBody, targetVersion, targetComponent string) bool {
var data map[string]interface{}
if err := json.Unmarshal([]byte(inventoryBody), &data); err != nil {
return false
}
members, ok := data["Members"].([]interface{})
if !ok {
return false
}
for _, m := range members {
entry, ok := m.(map[string]interface{})
if !ok {
continue
}
name, _ := entry["Name"].(string)
version, _ := entry["Version"].(string)
if targetComponent != "" && name != targetComponent {
continue
}
if version == targetVersion {
return true
}
}
return false
}