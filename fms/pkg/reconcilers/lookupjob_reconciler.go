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

v1 "github.com/bmcdonald3/fms/apis/example.fabrica.dev/v1"
"github.com/bmcdonald3/fms/internal/hms"
)

// reconcileLookupJob fetches the firmware inventory from a target node via Redfish.
//
// Resolution order for the node address:
//  1. DeviceProfile with matching name (uses ManagementIP field)
//  2. HMS mock FQDN (<node>.local)
//
// The raw JSON response is stored in Status.FirmwareData. On any network error
// the task transitions to "Failed" so callers can observe the attempted execution.
func (r *LookupJobReconciler) reconcileLookupJob(ctx context.Context, res *v1.LookupJob) error {
if res.Status.State == "Complete" || res.Status.State == "Failed" {
return nil
}

host := ""

// Try to resolve the host from a DeviceProfile first.
items, err := r.Client.List(ctx, "DeviceProfile")
if err == nil {
for _, item := range items {
data, err := json.Marshal(item)
if err != nil {
continue
}
var dp v1.DeviceProfile
if err := json.Unmarshal(data, &dp); err != nil {
continue
}
if dp.Metadata.Name == res.Spec.TargetNode && dp.Spec.ManagementIp != "" {
host = dp.Spec.ManagementIp
break
}
}
}

// Fall back to HMS FQDN resolution.
if host == "" {
hmsClient := hms.NewLocalHMS()
fqdn, err := hmsClient.GetDeviceFQDN(res.Spec.TargetNode)
if err != nil {
host = res.Spec.TargetNode
} else {
host = fqdn
}
}

url := fmt.Sprintf("https://%s/redfish/v1/UpdateService/FirmwareInventory", host)

httpClient := &http.Client{
Transport: &http.Transport{
TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
},
}

req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
if err != nil {
res.Status.State = "Failed"
r.Logger.Errorf("Failed to build request for %s: %v", host, err)
} else {
resp, err := httpClient.Do(req)
if err != nil {
r.Logger.Warnf("FirmwareInventory GET failed for %s: %v", host, err)
res.Status.State = "Failed"
} else {
defer resp.Body.Close()
body, readErr := io.ReadAll(resp.Body)
if readErr != nil {
res.Status.State = "Failed"
r.Logger.Errorf("Failed to read FirmwareInventory response: %v", readErr)
} else {
res.Status.FirmwareData = string(body)
res.Status.State = "Complete"
r.Logger.Infof("FirmwareInventory fetched for %s (%d bytes)", host, len(body))
}
}
}

if err := r.Client.Update(ctx, res); err != nil {
return fmt.Errorf("failed to persist LookupJob status: %w", err)
}

return nil
}