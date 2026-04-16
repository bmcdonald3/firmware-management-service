// Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package v1

import (
"context"
"github.com/openchami/fabrica/pkg/fabrica"
)

// FirmwareProfile represents a firmwareprofile resource
type FirmwareProfile struct {
APIVersion string              `json:"apiVersion"`
Kind       string              `json:"kind"`
Metadata   fabrica.Metadata    `json:"metadata"`
Spec       FirmwareProfileSpec `json:"spec" validate:"required"`
Status     FirmwareProfileStatus `json:"status,omitempty"`
}

// FirmwareProfileSpec defines the desired state of FirmwareProfile
type FirmwareProfileSpec struct {
VersionString   string   `json:"versionString" validate:"required"`
VersionNumber   string   `json:"versionNumber" validate:"required"`
TargetComponent string   `json:"targetComponent" validate:"required"`
PreConditions   []string `json:"preConditions,omitempty"`
PostConditions  []string `json:"postConditions,omitempty"`
SoftwareId      string   `json:"softwareId,omitempty"`
}

// FirmwareProfileStatus defines the observed state of FirmwareProfile
type FirmwareProfileStatus struct {
Phase   string `json:"phase,omitempty"`
Message string `json:"message,omitempty"`
Ready   bool   `json:"ready"`
}

// Validate implements custom validation logic for FirmwareProfile
func (r *FirmwareProfile) Validate(ctx context.Context) error {
return nil
}

// GetKind returns the kind of the resource
func (r *FirmwareProfile) GetKind() string {
return "FirmwareProfile"
}

// GetName returns the name of the resource
func (r *FirmwareProfile) GetName() string {
return r.Metadata.Name
}

// GetUID returns the UID of the resource
func (r *FirmwareProfile) GetUID() string {
return r.Metadata.UID
}

// IsHub marks this as the hub/storage version
func (r *FirmwareProfile) IsHub() {}