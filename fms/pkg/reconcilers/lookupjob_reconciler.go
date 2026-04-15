// Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT
// This file contains user-customizable reconciliation logic for LookupJob.
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
"time"

v1 "github.com/bmcdonald3/firmware-management-service/fms/apis/firmware.management.io/v1"
)

// reconcileLookupJob performs a Redfish firmware inventory lookup for the
// target node and persists the result to Status.FirmwareData.
//
// Steps:
//   0. Fetch the DeviceProfile to obtain ManagementIP and RedfishPath
//   1. HTTP GET to the Redfish FirmwareInventory endpoint (15s timeout, TLS insecure)
//   2. Store raw JSON in Status.FirmwareData, set State = "Complete", persist
func (r *LookupJobReconciler) reconcileLookupJob(ctx context.Context, res *v1.LookupJob) error {
// ── Step 0: Fetch DeviceProfile ──────────────────────────────────────────

deviceProfile, err := getTyped[v1.DeviceProfile](ctx, r.Client, "DeviceProfile", res.Spec.TargetNode)
if err != nil {
return fmt.Errorf("step0: fetch DeviceProfile %q: %w", res.Spec.TargetNode, err)
}

// ── Step 1: Query Redfish FirmwareInventory ──────────────────────────────

inventoryURL := fmt.Sprintf("https://%s/redfish/v1/UpdateService/FirmwareInventory",
deviceProfile.Spec.ManagementIP)

httpClient := &http.Client{
Timeout: 15 * time.Second,
Transport: &http.Transport{
TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
},
}

reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, inventoryURL, nil)
cancel()
if err != nil {
return fmt.Errorf("step1: build request: %w", err)
}

resp, err := httpClient.Do(req)
if err != nil {
return fmt.Errorf("step1: GET %s: %w", inventoryURL, err)
}
defer resp.Body.Close()

body, err := io.ReadAll(resp.Body)
if err != nil {
return fmt.Errorf("step1: read response body: %w", err)
}

// Validate that the response is parseable JSON before storing it.
var check json.RawMessage
if err := json.Unmarshal(body, &check); err != nil {
return fmt.Errorf("step1: response is not valid JSON: %w", err)
}

// ── Step 2: Persist results ──────────────────────────────────────────────

res.Status.FirmwareData = string(body)
res.Status.State = "Complete"

if err := r.Client.Update(ctx, res); err != nil {
return fmt.Errorf("step2: persist LookupJob status: %w", err)
}

r.Logger.Infof("LookupJob %s complete; stored %d bytes of firmware inventory for node %s",
res.Metadata.UID, len(body), res.Spec.TargetNode)

return nil
}