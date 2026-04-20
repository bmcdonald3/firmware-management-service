package v1
import "github.com/openchami/fabrica/pkg/fabrica"
type FirmwareProfile struct {
	APIVersion string `json:"apiVersion" validate:"required"`
	Kind       string `json:"kind" validate:"required"`
	Metadata   fabrica.Metadata `json:"metadata"`
	Spec       FirmwareProfileSpec `json:"spec"`
	Status     FirmwareProfileStatus `json:"status,omitempty"`
}
type FirmwareProfileSpec struct {
	VersionString   string `json:"versionString" validate:"required"`
	VersionNumber   string `json:"versionNumber" validate:"required"`
	TargetComponent string `json:"targetComponent" validate:"required"`
	PreConditions   string `json:"preConditions,omitempty"`
	PostConditions  string `json:"postConditions,omitempty"`
	SoftwareId      string `json:"softwareId,omitempty"`
}
type FirmwareProfileStatus struct {
	Phase   string `json:"phase,omitempty"`
	Message string `json:"message,omitempty"`
}
func (r *FirmwareProfile) GetKind() string { return "FirmwareProfile" }
func (r *FirmwareProfile) GetName() string { return r.Metadata.Name }
func (r *FirmwareProfile) GetUID() string { return r.Metadata.UID }
func (r *FirmwareProfile) IsHub() {}
