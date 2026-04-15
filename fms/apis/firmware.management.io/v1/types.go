package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeviceProfile

type DeviceProfileSpec struct {
	Manufacturer  string `json:"manufacturer" validate:"required"`
	Model         string `json:"model" validate:"required"`
	RedfishPath   string `json:"redfishPath" validate:"required"`
	ManagementIP  string `json:"managementIp" validate:"required"`
}

type DeviceProfile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              DeviceProfileSpec `json:"spec,omitempty"`
}

// FirmwareManifest

type FirmwareManifestSpec struct {
	VersionString   string `json:"versionString" validate:"required"`
	VersionNumber   string `json:"versionNumber" validate:"required"`
	TargetComponent string `json:"targetComponent" validate:"required"`
	PreConditions   string `json:"preConditions,omitempty"`
	PostConditions  string `json:"postConditions,omitempty"`
	SoftwareID      string `json:"softwareId,omitempty"`
	UpdateProfileRef string `json:"updateProfileRef" validate:"required"`
}

type FirmwareManifest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FirmwareManifestSpec `json:"spec,omitempty"`
}

// UpdateProfile

type UpdateProfileSpec struct {
	CommandType    string `json:"commandType" validate:"required,oneof=Redfish SSH"`
	PayloadPath    string `json:"payloadPath,omitempty"`
	SuccessCriteria string `json:"successCriteria,omitempty"`
}

type UpdateProfile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              UpdateProfileSpec `json:"spec,omitempty"`
}

// UpdateJob

type UpdateJobSpec struct {
	TargetNode     string `json:"targetNode" validate:"required"`
	TargetComponent string `json:"targetComponent" validate:"required"`
	FirmwareRef    string `json:"firmwareRef" validate:"required"`
	DryRun         bool   `json:"dryRun,omitempty"`
	Force          bool   `json:"force,omitempty"`
}

type UpdateJobStatus struct {
	State               string `json:"state,omitempty" validate:"oneof=Pending Running Success Failed"`
	StartTime           int64  `json:"startTime,omitempty"`
	TimeoutLimitSeconds int    `json:"timeoutLimitSeconds,omitempty"`
	TaskURI             string `json:"taskUri,omitempty"`
	JobID               string `json:"jobId,omitempty"`
}

type UpdateJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              UpdateJobSpec   `json:"spec,omitempty"`
	Status            UpdateJobStatus `json:"status,omitempty"`
}

// LookupJob

type LookupJobSpec struct {
	TargetNode string `json:"targetNode" validate:"required"`
}

type LookupJobStatus struct {
	State        string `json:"state,omitempty" validate:"oneof=Pending Running Complete Failed"`
	JobID        string `json:"jobId,omitempty"`
	FirmwareData string `json:"firmwareData,omitempty"`
}

type LookupJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              LookupJobSpec   `json:"spec,omitempty"`
	Status            LookupJobStatus `json:"status,omitempty"`
}