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
"fmt"
"io"
"net/http"
"time"

v1 "github.com/bmcdonald3/fms/apis/firmware.management.io/v1"
"github.com/bmcdonald3/fms/internal/hms"
)

// reconcileLookupJob queries the Redfish FirmwareInventory endpoint on the
// target node and stores the raw JSON response in Status.FirmwareData.
func (r *LookupJobReconciler) reconcileLookupJob(ctx context.Context, res *v1.LookupJob) error {
if res.Status.State == "Complete" || res.Status.State == "Failed" {
return nil
}

r.Logger.Infof("Reconciling LookupJob %s for node %s", res.GetUID(), res.Spec.TargetNode)

// Resolve management IP from DeviceProfile; fall back to HMS FQDN.
managementIP, err := r.resolveLookupIP(ctx, res.Spec.TargetNode)
if err != nil {
r.Logger.Warnf("Could not resolve IP for %s: %v — using node name", res.Spec.TargetNode, err)
managementIP = res.Spec.TargetNode
}

hmsClient := hms.NewLocalHMS()
user, pass, err := hmsClient.GetDeviceCredentials(res.Spec.TargetNode)
if err != nil {
return fmt.Errorf("failed to get credentials: %w", err)
}

inventoryURL := fmt.Sprintf("https://%s/redfish/v1/UpdateService/FirmwareInventory", managementIP)

// Use a dedicated http.Transport with InsecureSkipVerify so self-signed BMC
// certificates do not cause the request to fail.
transport := &http.Transport{
TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}
httpClient := &http.Client{
Transport: transport,
Timeout:   30 * time.Second,
}

req, err := http.NewRequestWithContext(ctx, http.MethodGet, inventoryURL, nil)
if err != nil {
return fmt.Errorf("failed to build inventory request: %w", err)
}
if user != "" && pass != "" {
req.SetBasicAuth(user, pass)
}
req.Header.Set("Accept", "application/json")

resp, err := httpClient.Do(req)
if err != nil {
res.Status.State = "Failed"
if updateErr := r.Client.Update(ctx, res); updateErr != nil {
r.Logger.Errorf("Failed to persist Failed state for LookupJob %s: %v", res.GetUID(), updateErr)
}
return fmt.Errorf("inventory GET failed for %s: %w", managementIP, err)
}
defer resp.Body.Close()

body, err := io.ReadAll(resp.Body)
if err != nil {
return fmt.Errorf("failed to read inventory response: %w", err)
}

res.Status.FirmwareData = string(body)
res.Status.State = "Complete"
res.Status.JobId = res.GetUID()

if err := r.Client.Update(ctx, res); err != nil {
return fmt.Errorf("failed to persist LookupJob status: %w", err)
}
r.Logger.Infof("LookupJob %s complete, %d bytes stored", res.GetUID(), len(body))
return nil
}

// resolveLookupIP finds the DeviceProfile whose name matches targetNode and
// returns its ManagementIp. Falls back to HMS FQDN when no profile is found.
func (r *LookupJobReconciler) resolveLookupIP(ctx context.Context, targetNode string) (string, error) {
items, err := r.Client.List(ctx, "DeviceProfile")
if err != nil {
hmsClient := hms.NewLocalHMS()
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
hmsClient := hms.NewLocalHMS()
return hmsClient.GetDeviceFQDN(targetNode)
}