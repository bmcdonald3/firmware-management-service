// Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package v1

import (
"context"
"github.com/openchami/fabrica/pkg/fabrica"
)

// FirmwareManifest represents a firmwaremanifest resource
type FirmwareManifest struct {
APIVersion string                 `json:"apiVersion"`
Kind       string                 `json:"kind"`
Metadata   fabrica.Metadata       `json:"metadata"`
Spec       FirmwareManifestSpec   `json:"spec" validate:"required"`
Status     FirmwareManifestStatus `json:"status,omitempty"`
}

// FirmwareManifestSpec defines the desired state of FirmwareManifest
type FirmwareManifestSpec struct {
VersionString    string `json:"versionString" validate:"required"`
VersionNumber    string `json:"versionNumber" validate:"required"`
TargetComponent  string `json:"targetComponent" validate:"required"`
PreConditions    string `json:"preConditions"`
PostConditions   string `json:"postConditions"`
SoftwareID       string `json:"softwareId"`
UpdateProfileRef string `json:"updateProfileRef" validate:"required"`
}

// FirmwareManifestStatus defines the observed state of FirmwareManifest
type FirmwareManifestStatus struct {
Phase   string `json:"phase,omitempty"`
Message string `json:"message,omitempty"`
Ready   bool   `json:"ready"`
}

// Validate implements custom validation logic for FirmwareManifest
func (r *FirmwareManifest) Validate(ctx context.Context) error {
return nil
}

// GetKind returns the kind of the resource
func (r *FirmwareManifest) GetKind() string {
return "FirmwareManifest"
}

// GetName returns the name of the resource
func (r *FirmwareManifest) GetName() string {
return r.Metadata.Name
}

// GetUID returns the UID of the resource
func (r *FirmwareManifest) GetUID() string {
return r.Metadata.UID
}

// IsHub marks this as the hub/storage version
func (r *FirmwareManifest) IsHub() {}