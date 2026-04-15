// Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package v1

import (
	"context"
	"github.com/openchami/fabrica/pkg/fabrica"
)

// DeviceProfile represents a deviceprofile resource
type DeviceProfile struct {
	APIVersion string              `json:"apiVersion"`
	Kind       string              `json:"kind"`
	Metadata   fabrica.Metadata    `json:"metadata"`
	Spec       DeviceProfileSpec   `json:"spec" validate:"required"`
	Status     DeviceProfileStatus `json:"status,omitempty"`
}

// DeviceProfileSpec defines the desired state of DeviceProfile
type DeviceProfileSpec struct {
	Manufacturer string `json:"manufacturer" validate:"required"`
	Model        string `json:"model" validate:"required"`
	RedfishPath  string `json:"redfishPath" validate:"required"`
	ManagementIP string `json:"managementIp" validate:"required"`
}

// DeviceProfileStatus defines the observed state of DeviceProfile
type DeviceProfileStatus struct {
	Phase   string `json:"phase,omitempty"`
	Message string `json:"message,omitempty"`
	Ready   bool   `json:"ready"`
}

// Validate implements custom validation logic for DeviceProfile
func (r *DeviceProfile) Validate(ctx context.Context) error {
	return nil
}

// GetKind returns the kind of the resource
func (r *DeviceProfile) GetKind() string {
	return "DeviceProfile"
}

// GetName returns the name of the resource
func (r *DeviceProfile) GetName() string {
	return r.Metadata.Name
}

// GetUID returns the UID of the resource
func (r *DeviceProfile) GetUID() string {
	return r.Metadata.UID
}

// IsHub marks this as the hub/storage version
func (r *DeviceProfile) IsHub() {}