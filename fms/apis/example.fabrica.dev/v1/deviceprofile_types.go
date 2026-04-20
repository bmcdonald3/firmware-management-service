package v1
import "github.com/openchami/fabrica/pkg/fabrica"
type DeviceProfile struct {
	APIVersion string `json:"apiVersion" validate:"required"`
	Kind       string `json:"kind" validate:"required"`
	Metadata   fabrica.Metadata `json:"metadata"`
	Spec       DeviceProfileSpec `json:"spec"`
	Status     DeviceProfileStatus `json:"status,omitempty"`
}
type DeviceProfileSpec struct {
	Manufacturer string `json:"manufacturer" validate:"required"`
	Model        string `json:"model" validate:"required"`
	RedfishPath  string `json:"redfishPath" validate:"required"`
	ManagementIp string `json:"managementIp" validate:"required"`
}
type DeviceProfileStatus struct {
	Phase   string `json:"phase,omitempty"`
	Message string `json:"message,omitempty"`
}
func (r *DeviceProfile) GetKind() string { return "DeviceProfile" }
func (r *DeviceProfile) GetName() string { return r.Metadata.Name }
func (r *DeviceProfile) GetUID() string { return r.Metadata.UID }
func (r *DeviceProfile) IsHub() {}
