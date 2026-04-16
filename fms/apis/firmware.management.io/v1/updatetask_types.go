// Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package v1

import (
"context"
"github.com/openchami/fabrica/pkg/fabrica"
)

// UpdateTask represents an updatetask resource
type UpdateTask struct {
APIVersion string            `json:"apiVersion"`
Kind       string            `json:"kind"`
Metadata   fabrica.Metadata  `json:"metadata"`
Spec       UpdateTaskSpec    `json:"spec" validate:"required"`
Status     UpdateTaskStatus  `json:"status,omitempty"`
}

// UpdateTaskSpec defines the desired state of UpdateTask
type UpdateTaskSpec struct {
TargetNode     string `json:"targetNode,omitempty"`
TargetComponent string `json:"targetComponent,omitempty"`
FirmwareRef    string `json:"firmwareRef,omitempty"`
UpdateJobId    string `json:"updateJobId,omitempty"`
}

// UpdateTaskStatus defines the observed state of UpdateTask
type UpdateTaskStatus struct {
State     string `json:"state,omitempty" validate:"oneof=Pending Running Success Failed"`
StartTime int64  `json:"startTime,omitempty"`
TaskUri   string `json:"taskUri,omitempty"`
}

// Validate implements custom validation logic for UpdateTask
func (r *UpdateTask) Validate(ctx context.Context) error {
return nil
}

// GetKind returns the kind of the resource
func (r *UpdateTask) GetKind() string {
return "UpdateTask"
}

// GetName returns the name of the resource
func (r *UpdateTask) GetName() string {
return r.Metadata.Name
}

// GetUID returns the UID of the resource
func (r *UpdateTask) GetUID() string {
return r.Metadata.UID
}

// IsHub marks this as the hub/storage version
func (r *UpdateTask) IsHub() {}