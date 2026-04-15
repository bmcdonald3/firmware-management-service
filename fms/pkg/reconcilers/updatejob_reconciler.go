// Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT
// This file contains user-customizable reconciliation logic for UpdateJob.
//
// ⚠️ This file is safe to edit - it will NOT be overwritten by code generation.
package reconcilers

import (
"context"
"crypto/tls"
"encoding/json"
"fmt"
"io"
"net/http"
"os"
"strings"
"time"

v1 "github.com/bmcdonald3/firmware-management-service/fms/apis/firmware.management.io/v1"
)

// reconcileUpdateJob implements the full firmware update state machine.
//
// Steps follow the plan specification exactly:
//   0. Load relational data (DeviceProfile, FirmwareManifest, UpdateProfile)
//   1. Locking — prevent concurrent jobs on the same node/component
//   2. Dry-run fast-path
//   3. Pre-flight checks (version match + preconditions)
//   4. Execution initiation (Redfish task or SSH)
//   5. Monitoring an in-progress task until terminal state or timeout
func (r *UpdateJobReconciler) reconcileUpdateJob(ctx context.Context, res *v1.UpdateJob) error {
// ── Step 0: Load relational data ────────────────────────────────────────

deviceProfile, err := getTyped[v1.DeviceProfile](ctx, r.Client, "DeviceProfile", res.Spec.TargetNode)
if err != nil {
return fmt.Errorf("step0: fetch DeviceProfile %q: %w", res.Spec.TargetNode, err)
}

manifest, err := getTyped[v1.FirmwareManifest](ctx, r.Client, "FirmwareManifest", res.Spec.FirmwareRef)
if err != nil {
return fmt.Errorf("step0: fetch FirmwareManifest %q: %w", res.Spec.FirmwareRef, err)
}

updateProfile, err := getTyped[v1.UpdateProfile](ctx, r.Client, "UpdateProfile", manifest.Spec.UpdateProfileRef)
if err != nil {
return fmt.Errorf("step0: fetch UpdateProfile %q: %w", manifest.Spec.UpdateProfileRef, err)
}

// ── Step 1: Locking ─────────────────────────────────────────────────────
// If another job is Running against the same node+component, defer.

rawJobs, err := r.Client.List(ctx, "UpdateJob")
if err != nil {
return fmt.Errorf("step1: listing UpdateJobs: %w", err)
}

for _, raw := range rawJobs {
j, err := marshalTyped[v1.UpdateJob](raw)
if err != nil {
continue
}
if j.Metadata.UID == res.Metadata.UID {
continue
}
if j.Status.State == "Running" &&
j.Spec.TargetNode == res.Spec.TargetNode &&
j.Spec.TargetComponent == res.Spec.TargetComponent {
r.Logger.Infof("step1: conflicting job %s running; requeue in 5s", j.Metadata.UID)
return fmt.Errorf("step1: lock held by job %s", j.Metadata.UID)
}
}

// ── Step 2: Dry-run fast-path ────────────────────────────────────────────

if res.Spec.DryRun {
res.Status.State = "Success"
if err := r.Client.Update(ctx, res); err != nil {
return fmt.Errorf("step2: persist dry-run success: %w", err)
}
r.Logger.Infof("step2: dry-run complete for job %s", res.Metadata.UID)
return nil
}

// ── Step 3: Pre-flight ───────────────────────────────────────────────────

httpClient := &http.Client{
Timeout: 15 * time.Second,
Transport: &http.Transport{
TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
},
}

if !res.Spec.Force {
// Check whether the installed firmware version already matches the desired one.
versionURL := fmt.Sprintf("https://%s%s/FirmwareVersion",
deviceProfile.Spec.ManagementIP, deviceProfile.Spec.RedfishPath)

vCtx, vCancel := context.WithTimeout(ctx, 15*time.Second)
vReq, _ := http.NewRequestWithContext(vCtx, http.MethodGet, versionURL, nil)
vResp, vErr := httpClient.Do(vReq)
vCancel()

if vErr == nil {
defer vResp.Body.Close()
body, _ := io.ReadAll(vResp.Body)
var payload struct {
Version string `json:"Version"`
}
if json.Unmarshal(body, &payload) == nil && payload.Version == manifest.Spec.VersionString {
r.Logger.Infof("step3: version already %s; marking Success", manifest.Spec.VersionString)
res.Status.State = "Success"
if err := r.Client.Update(ctx, res); err != nil {
return fmt.Errorf("step3: persist version-match success: %w", err)
}
return nil
}
}

// Evaluate PreConditions (e.g. power-state endpoint).
if manifest.Spec.PreConditions != "" {
pcURL := fmt.Sprintf("https://%s%s/%s",
deviceProfile.Spec.ManagementIP, deviceProfile.Spec.RedfishPath, manifest.Spec.PreConditions)

pcCtx, pcCancel := context.WithTimeout(ctx, 15*time.Second)
pcReq, _ := http.NewRequestWithContext(pcCtx, http.MethodGet, pcURL, nil)
pcResp, pcErr := httpClient.Do(pcReq)
pcCancel()

if pcErr != nil || pcResp.StatusCode != http.StatusOK {
r.Logger.Warnf("step3: precondition %q not met; requeue in 10s", manifest.Spec.PreConditions)
return fmt.Errorf("step3: precondition not met: %v", pcErr)
}
pcResp.Body.Close()
}
}

// ── Step 4: Execution ────────────────────────────────────────────────────
// Only initiate if no outgoing task is already tracked.

if res.Status.TaskURI == "" {
fmsHostIP := os.Getenv("FMS_HOST_IP")

switch updateProfile.Spec.CommandType {
case "Redfish":
updateURL := fmt.Sprintf("https://%s%s/Actions/UpdateService.SimpleUpdate",
deviceProfile.Spec.ManagementIP, deviceProfile.Spec.RedfishPath)

payloadBody := fmt.Sprintf(
`{"ImageURI":"http://%s/library/%s","Targets":["%s"]}`,
fmsHostIP, updateProfile.Spec.PayloadPath, res.Spec.TargetComponent)

postCtx, postCancel := context.WithTimeout(ctx, 30*time.Second)
postReq, _ := http.NewRequestWithContext(postCtx, http.MethodPost, updateURL,
strings.NewReader(payloadBody))
postReq.Header.Set("Content-Type", "application/json")
postResp, postErr := httpClient.Do(postReq)
postCancel()

if postErr != nil {
return fmt.Errorf("step4: redfish POST failed: %w", postErr)
}
postResp.Body.Close()

taskURI := postResp.Header.Get("Location")
if taskURI == "" {
return fmt.Errorf("step4: redfish response missing Location header")
}
res.Status.TaskURI = taskURI

case "SSH":
sshCtx, sshCancel := context.WithTimeout(ctx, 60*time.Second)
defer sshCancel()

if err := runSSHUpdate(sshCtx, *deviceProfile, *updateProfile, fmsHostIP); err != nil {
return fmt.Errorf("step4: SSH update failed: %w", err)
}
res.Status.TaskURI = "ssh://done"

default:
return fmt.Errorf("step4: unknown CommandType %q", updateProfile.Spec.CommandType)
}

res.Status.State = "Running"
res.Status.StartTime = time.Now().Unix()
if err := r.Client.Update(ctx, res); err != nil {
return fmt.Errorf("step4: persist Running state: %w", err)
}
}

// ── Step 5: Monitoring ───────────────────────────────────────────────────

if res.Status.TimeoutLimitSeconds > 0 &&
time.Now().Unix()-res.Status.StartTime > int64(res.Status.TimeoutLimitSeconds) {
r.Logger.Warnf("step5: job %s timed out", res.Metadata.UID)
res.Status.State = "Failed"
if err := r.Client.Update(ctx, res); err != nil {
return fmt.Errorf("step5: persist timeout failure: %w", err)
}
return nil
}

// Poll the Redfish task monitor (SSH completes synchronously).
if res.Status.TaskURI != "" && res.Status.TaskURI != "ssh://done" {
pollCtx, pollCancel := context.WithTimeout(ctx, 15*time.Second)
pollReq, _ := http.NewRequestWithContext(pollCtx, http.MethodGet, res.Status.TaskURI, nil)
pollResp, pollErr := httpClient.Do(pollReq)
pollCancel()

if pollErr != nil {
r.Logger.Warnf("step5: poll error %v; will requeue", pollErr)
return pollErr
}
defer pollResp.Body.Close()

body, _ := io.ReadAll(pollResp.Body)
var task struct {
TaskState string `json:"TaskState"`
}
if err := json.Unmarshal(body, &task); err != nil {
return fmt.Errorf("step5: unmarshal task response: %w", err)
}

switch task.TaskState {
case "Completed":
res.Status.State = "Success"
case "Exception", "Killed":
res.Status.State = "Failed"
default:
if err := r.Client.Update(ctx, res); err != nil {
return fmt.Errorf("step5: persist in-progress state: %w", err)
}
return fmt.Errorf("step5: task still %s; requeue", task.TaskState)
}

if err := r.Client.Update(ctx, res); err != nil {
return fmt.Errorf("step5: persist terminal state: %w", err)
}
}

return nil
}

// getTyped retrieves a resource by kind+uid and deserialises it into type T.
func getTyped[T any](ctx context.Context, client interface {
Get(context.Context, string, string) (interface{}, error)
}, kind, uid string) (*T, error) {
raw, err := client.Get(ctx, kind, uid)
if err != nil {
return nil, err
}
b, err := json.Marshal(raw)
if err != nil {
return nil, fmt.Errorf("marshal %s: %w", kind, err)
}
var out T
if err := json.Unmarshal(b, &out); err != nil {
return nil, fmt.Errorf("unmarshal %s: %w", kind, err)
}
return &out, nil
}

// marshalTyped converts an interface{} value (from ClientInterface.List) into a typed struct.
func marshalTyped[T any](raw interface{}) (*T, error) {
b, err := json.Marshal(raw)
if err != nil {
return nil, err
}
var out T
if err := json.Unmarshal(b, &out); err != nil {
return nil, err
}
return &out, nil
}