package hms

import (
"os"
)

type HMSInterface interface {
HMSInit() error
GetDeviceFQDN(deviceName string) (string, error)
GetDeviceCredentials(deviceName string) (string, string, error)
}

type LocalHMS struct{}

func NewLocalHMS() *LocalHMS {
return &LocalHMS{}
}

func (h *LocalHMS) HMSInit() error {
return nil
}

// GetDeviceFQDN mocks FQDN resolution. For Fabrica, use the DeviceProfile's ManagementIP if this fails.
func (h *LocalHMS) GetDeviceFQDN(deviceName string) (string, error) {
return deviceName + ".local", nil
}

// GetDeviceCredentials mocks credential resolution.
func (h *LocalHMS) GetDeviceCredentials(deviceName string) (string, string, error) {
user := os.Getenv("HPEUSER")
if user == "" {
user = "root"
}

pass := os.Getenv("HPEPASS")
if pass == "" {
pass = "secret"
}

return user, pass, nil
}