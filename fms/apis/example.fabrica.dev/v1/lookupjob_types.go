package v1
import "github.com/openchami/fabrica/pkg/fabrica"
type LookupJob struct {
	APIVersion string `json:"apiVersion" validate:"required"`
	Kind       string `json:"kind" validate:"required"`
	Metadata   fabrica.Metadata `json:"metadata"`
	Spec       LookupJobSpec `json:"spec"`
	Status     LookupJobStatus `json:"status,omitempty"`
}
type LookupJobSpec struct {
	TargetNode string `json:"targetNode" validate:"required"`
}
type LookupJobStatus struct {
	State        string `json:"state,omitempty" validate:"omitempty,oneof=Pending Running Complete Failed"`
	JobId        string `json:"jobId,omitempty"`
	FirmwareData string `json:"firmwareData,omitempty"`
}
func (r *LookupJob) GetKind() string { return "LookupJob" }
func (r *LookupJob) GetName() string { return r.Metadata.Name }
func (r *LookupJob) GetUID() string { return r.Metadata.UID }
func (r *LookupJob) IsHub() {}
