// Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package v1

import (
"context"
"github.com/openchami/fabrica/pkg/fabrica"
)

// UpdateJob represents an updatejob resource
type UpdateJob struct {
APIVersion string          `json:"apiVersion"`
Kind       string          `json:"kind"`
Metadata   fabrica.Metadata `json:"metadata"`
Spec       UpdateJobSpec   `json:"spec" validate:"required"`
Status     UpdateJobStatus `json:"status,omitempty"`
}

// UpdateJobSpec defines the desired state of UpdateJob
type UpdateJobSpec struct {
TargetNodes     []string `json:"targetNodes,omitempty"`
TargetComponent string   `json:"targetComponent,omitempty"`
FirmwareRef     string   `json:"firmwareRef,omitempty"`
DryRun          bool     `json:"dryRun,omitempty"`
Force           bool     `json:"force,omitempty"`
}

// UpdateJobStatus defines the observed state of UpdateJob
type UpdateJobStatus struct {
Phase   string `json:"phase,omitempty"`
Message string `json:"message,omitempty"`
Ready   bool   `json:"ready"`
}

// Validate implements custom validation logic for UpdateJob
func (r *UpdateJob) Validate(ctx context.Context) error {
return nil
}

// GetKind returns the kind of the resource
func (r *UpdateJob) GetKind() string {
return "UpdateJob"
}

// GetName returns the name of the resource
func (r *UpdateJob) GetName() string {
return r.Metadata.Name
}

// GetUID returns the UID of the resource
func (r *UpdateJob) GetUID() string {
return r.Metadata.UID
}

// IsHub marks this as the hub/storage version
func (r *UpdateJob) IsHub() {}