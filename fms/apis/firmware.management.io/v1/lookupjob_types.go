// Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package v1

import (
"context"
"github.com/openchami/fabrica/pkg/fabrica"
)

// LookupJob represents a lookupjob resource
type LookupJob struct {
APIVersion string           `json:"apiVersion"`
Kind       string           `json:"kind"`
Metadata   fabrica.Metadata `json:"metadata"`
Spec       LookupJobSpec    `json:"spec" validate:"required"`
Status     LookupJobStatus  `json:"status,omitempty"`
}

// LookupJobSpec defines the desired state of LookupJob
type LookupJobSpec struct {
TargetNode string `json:"targetNode,omitempty"`
}

// LookupJobStatus defines the observed state of LookupJob
type LookupJobStatus struct {
State        string `json:"state,omitempty" validate:"oneof=Pending Running Complete Failed"`
JobId        string `json:"jobId,omitempty"`
FirmwareData string `json:"firmwareData,omitempty"`
}

// Validate implements custom validation logic for LookupJob
func (r *LookupJob) Validate(ctx context.Context) error {
return nil
}

// GetKind returns the kind of the resource
func (r *LookupJob) GetKind() string {
return "LookupJob"
}

// GetName returns the name of the resource
func (r *LookupJob) GetName() string {
return r.Metadata.Name
}

// GetUID returns the UID of the resource
func (r *LookupJob) GetUID() string {
return r.Metadata.UID
}

// IsHub marks this as the hub/storage version
func (r *LookupJob) IsHub() {}