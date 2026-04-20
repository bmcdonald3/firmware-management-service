package v1
import "github.com/openchami/fabrica/pkg/fabrica"
type UpdateProfile struct {
	APIVersion string `json:"apiVersion" validate:"required"`
	Kind       string `json:"kind" validate:"required"`
	Metadata   fabrica.Metadata `json:"metadata"`
	Spec       UpdateProfileSpec `json:"spec"`
	Status     UpdateProfileStatus `json:"status,omitempty"`
}
type UpdateProfileSpec struct {
	CommandType     string `json:"commandType" validate:"required,oneof=Redfish SSH"`
	PayloadPath     string `json:"payloadPath,omitempty"`
	SuccessCriteria string `json:"successCriteria,omitempty"`
}
type UpdateProfileStatus struct {
	Phase   string `json:"phase,omitempty"`
	Message string `json:"message,omitempty"`
}
func (r *UpdateProfile) GetKind() string { return "UpdateProfile" }
func (r *UpdateProfile) GetName() string { return r.Metadata.Name }
func (r *UpdateProfile) GetUID() string { return r.Metadata.UID }
func (r *UpdateProfile) IsHub() {}
