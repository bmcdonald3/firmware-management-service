package v1
import "github.com/openchami/fabrica/pkg/fabrica"
type UpdateTask struct {
	APIVersion string `json:"apiVersion" validate:"required"`
	Kind       string `json:"kind" validate:"required"`
	Metadata   fabrica.Metadata `json:"metadata"`
	Spec       UpdateTaskSpec `json:"spec"`
	Status     UpdateTaskStatus `json:"status,omitempty"`
}
type UpdateTaskSpec struct {
	TargetNode      string `json:"targetNode" validate:"required"`
	TargetComponent string `json:"targetComponent,omitempty"`
	FirmwareRef     string `json:"firmwareRef,omitempty"`
	UpdateJobId     string `json:"updateJobId" validate:"required"`
}
type UpdateTaskStatus struct {
	State     string `json:"state,omitempty" validate:"omitempty,oneof=Pending Running Success Failed"`
	StartTime int64  `json:"startTime,omitempty"`
	TaskUri   string `json:"taskUri,omitempty"`
}
func (r *UpdateTask) GetKind() string { return "UpdateTask" }
func (r *UpdateTask) GetName() string { return r.Metadata.Name }
func (r *UpdateTask) GetUID() string { return r.Metadata.UID }
func (r *UpdateTask) IsHub() {}
