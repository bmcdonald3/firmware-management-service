package v1
import "github.com/openchami/fabrica/pkg/fabrica"
type UpdateJob struct {
	APIVersion string `json:"apiVersion" validate:"required"`
	Kind       string `json:"kind" validate:"required"`
	Metadata   fabrica.Metadata `json:"metadata"`
	Spec       UpdateJobSpec `json:"spec"`
	Status     UpdateJobStatus `json:"status,omitempty"`
}
type UpdateJobSpec struct {
	TargetNodes     []string `json:"targetNodes" validate:"required"`
	TargetComponent string   `json:"targetComponent,omitempty"`
	FirmwareRef     string   `json:"firmwareRef,omitempty"`
	DryRun          bool     `json:"dryRun,omitempty"`
	Force           bool     `json:"force,omitempty"`
}
type UpdateJobStatus struct {
	Phase   string `json:"phase,omitempty"`
	Message string `json:"message,omitempty"`
}
func (r *UpdateJob) GetKind() string { return "UpdateJob" }
func (r *UpdateJob) GetName() string { return r.Metadata.Name }
func (r *UpdateJob) GetUID() string { return r.Metadata.UID }
func (r *UpdateJob) IsHub() {}
