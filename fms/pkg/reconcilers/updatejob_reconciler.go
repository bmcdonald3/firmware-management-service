// Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT
// This file contains user-customizable reconciliation logic for UpdateJob.
//
// ⚠️ This file is safe to edit - it will NOT be overwritten by code generation.
package reconcilers

import (
"context"
"time"

"github.com/user/fms/apis/example.fabrica.dev/v1"
)

// reconcileUpdateJob contains custom reconciliation logic.
//
// This method is called by the generated Reconcile() orchestration method.
// Implement UpdateJob-specific reconciliation logic here.
//
// Guidelines:
//  1. Keep this method idempotent (safe to call multiple times)
//  2. Update Status fields to reflect observed state
//  3. Emit events for significant state changes using r.EmitEvent()
//  4. Use r.Logger for debugging (Infof, Warnf, Errorf, Debugf)
//  5. Return errors for transient failures (will retry with backoff)
//  6. Access storage via r.Client (Get, List, Update, Create, Delete)
//
// Example implementation patterns:
//
// For hardware resources (BMC, Node):
//   - Connect to hardware endpoint
//   - Query current state
//   - Update Status.Connected, Status.Version, Status.Health
//   - Emit events when state changes
//
// For hierarchical resources (Rack, Chassis):
//   - Create/reconcile child resources
//   - Update Status with child counts and references
//   - Emit events when topology changes
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - res: The UpdateJob resource to reconcile
//
// Returns:
//   - error: If reconciliation failed (will trigger retry with backoff)
func (r *UpdateJobReconciler) reconcileUpdateJob(ctx context.Context, res *v1.UpdateJob) error {
	// Fetch FirmwareProfile
	var firmwareProfile v1.FirmwareProfile
	if err := r.Client.Get(ctx, res.Spec.FirmwareProfileId, &firmwareProfile); err != nil {
		r.Logger.Errorf("Failed to fetch FirmwareProfile: %v", err)
		res.Status.State = "failed"
		res.Status.Message = "FirmwareProfile not found"
		return err
	}
	// Fetch UpdateProfile
	var updateProfile v1.UpdateProfile
	if err := r.Client.Get(ctx, firmwareProfile.Spec.UpdateProfile, &updateProfile); err != nil {
		r.Logger.Errorf("Failed to fetch UpdateProfile: %v", err)
		res.Status.State = "failed"
		res.Status.Message = "UpdateProfile not found"
		return err
	}
	// Fetch DeviceProfile
	var deviceProfile v1.DeviceProfile
	if err := r.Client.Get(ctx, firmwareProfile.Spec.DeviceProfile, &deviceProfile); err != nil {
		r.Logger.Errorf("Failed to fetch DeviceProfile: %v", err)
		res.Status.State = "failed"
		res.Status.Message = "DeviceProfile not found"
		return err
	}

	// Retrieve credentials (stubbed or from env)
	username := getenvDefault("REDFISH_USER", "admin")
	password := getenvDefault("REDFISH_PASS", "password")

	// Build firmware binary URL
	binaryURL := "http://localhost:8080/firmware-binaries/" + firmwareProfile.Spec.FileName

	// Replace %httpFileName% in payload
	payload := updateProfile.Spec.Payload
	payload = replaceHTTPFileName(payload, binaryURL)

	// Handle job state machine
	switch res.Status.State {
	case "", "pending":
		// Construct Redfish HTTP request
		targetURL := "http://" + res.Spec.TargetNode + updateProfile.Spec.UpdatePath
		r.Logger.Infof("Sending Redfish update to %s with payload: %s", targetURL, payload)
		req, err := http.NewRequestWithContext(ctx, "POST", targetURL, strings.NewReader(payload))
		if err != nil {
			r.Logger.Errorf("Failed to create Redfish request: %v", err)
			res.Status.State = "failed"
			res.Status.Message = "Failed to create Redfish request"
			return err
		}
		req.SetBasicAuth(username, password)
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			r.Logger.Errorf("Redfish request failed: %v", err)
			res.Status.State = "failed"
			res.Status.Message = "Redfish request failed"
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			res.Status.State = "running"
			res.Status.Message = "Job started"
			res.Status.StartTime = time.Now().Format(time.RFC3339)
		} else {
			res.Status.State = "failed"
			res.Status.Message = "Redfish update rejected"
			return nil
		}
	case "running":
		// Poll firmwareInventory
		targetURL := "http://" + res.Spec.TargetNode + updateProfile.Spec.FirmwareInventory
		r.Logger.Infof("Polling firmware inventory at %s", targetURL)
		req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
		if err != nil {
			r.Logger.Errorf("Failed to create inventory request: %v", err)
			return err
		}
		req.SetBasicAuth(username, password)
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			r.Logger.Errorf("Inventory request failed: %v", err)
			return err
		}
		defer resp.Body.Close()
		// For demo: assume version matches if HTTP 200
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			res.Status.State = "complete"
			res.Status.Message = "Job completed successfully"
			res.Status.EndTime = time.Now().Format(time.RFC3339)
		} else {
			// Check timeout
			start, _ := time.Parse(time.RFC3339, res.Status.StartTime)
			timeout := updateProfile.Spec.DefaultTimeout
			if timeout == 0 {
				timeout = 300
			}
			if time.Since(start) > time.Duration(timeout)*time.Second {
				res.Status.State = "failed"
				res.Status.Message = "Timeout waiting for firmware update"
				res.Status.EndTime = time.Now().Format(time.RFC3339)
			} else {
				return &RequeueAfterError{Delay: 30 * time.Second}
			}
		}
	}
	return nil
}

// getenvDefault returns the value of the environment variable or a default.
func getenvDefault(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

// replaceHTTPFileName replaces %httpFileName% in the payload string.
func replaceHTTPFileName(payload, url string) string {
	return strings.ReplaceAll(payload, "%httpFileName%", url)
}

// RequeueAfterError signals the controller to requeue after a delay.
type RequeueAfterError struct {
	Delay time.Duration
}

func (e *RequeueAfterError) Error() string {
	return "requeue after"
}
