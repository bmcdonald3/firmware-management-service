// Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package v1

import (
"context"
"github.com/openchami/fabrica/pkg/fabrica"
)

// UpdateProfile represents an updateprofile resource
type UpdateProfile struct {
APIVersion string             `json:"apiVersion"`
Kind       string             `json:"kind"`
Metadata   fabrica.Metadata   `json:"metadata"`
Spec       UpdateProfileSpec  `json:"spec" validate:"required"`
Status     UpdateProfileStatus `json:"status,omitempty"`
}

// UpdateProfileSpec defines the desired state of UpdateProfile
type UpdateProfileSpec struct {
CommandType     string `json:"commandType" validate:"required,oneof=Redfish SSH"`
PayloadPath     string `json:"payloadPath,omitempty"`
SuccessCriteria string `json:"successCriteria,omitempty"`
}

// UpdateProfileStatus defines the observed state of UpdateProfile
type UpdateProfileStatus struct {
Phase   string `json:"phase,omitempty"`
Message string `json:"message,omitempty"`
Ready   bool   `json:"ready"`
}

// Validate implements custom validation logic for UpdateProfile
func (r *UpdateProfile) Validate(ctx context.Context) error {
return nil
}

// GetKind returns the kind of the resource
func (r *UpdateProfile) GetKind() string {
return "UpdateProfile"
}

// GetName returns the name of the resource
func (r *UpdateProfile) GetName() string {
return r.Metadata.Name
}

// GetUID returns the UID of the resource
func (r *UpdateProfile) GetUID() string {
return r.Metadata.UID
}

// IsHub marks this as the hub/storage version
func (r *UpdateProfile) IsHub() {}