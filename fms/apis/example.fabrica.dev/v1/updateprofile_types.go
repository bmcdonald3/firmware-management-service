// Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package v1

import (
	"context"
	"github.com/openchami/fabrica/pkg/fabrica"
)

// UpdateProfile represents a updateprofile resource
type UpdateProfile struct {
	APIVersion string           `json:"apiVersion"`
	Kind       string           `json:"kind"`
	Metadata   fabrica.Metadata `json:"metadata"`
	Spec       UpdateProfileSpec   `json:"spec" validate:"required"`
	Status     UpdateProfileStatus `json:"status,omitempty"`
}

// UpdateProfileSpec defines the desired state of UpdateProfile
type UpdateProfileSpec struct {
ProfileName            string `json:"profileName" validate:"required"`
Version                string `json:"version,omitempty"`
UpdateProfileVersion   string `json:"updateProfileVersion,omitempty"`
Protocol               string `json:"protocol" validate:"required"`
PushPull               string `json:"pushPull,omitempty"`
UpdatePath             string `json:"updatePath,omitempty"`
Payload                string `json:"payload,omitempty"`
FirmwareInventory      string `json:"firmwareInventory,omitempty"`
FirmwareInventoryExpand string `json:"firmwareInventoryExpand,omitempty"`
DefaultTimeout         int    `json:"defaultTimeout,omitempty"`
}

// UpdateProfileStatus defines the observed state of UpdateProfile
type UpdateProfileStatus struct {
	Phase      string `json:"phase,omitempty"`
	Message    string `json:"message,omitempty"`
	Ready      bool   `json:"ready"`
		// Add your status fields here
}

// Validate implements custom validation logic for UpdateProfile
func (r *UpdateProfile) Validate(ctx context.Context) error {
	// Add custom validation logic here
	// Example:
	// if r.Spec.Description == "forbidden" {
	//     return errors.New("description 'forbidden' is not allowed")
	// }

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
