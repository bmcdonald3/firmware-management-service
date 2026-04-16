// Package hms provides HMS (Hardware Management Service) interface and local mock.
// In production this would resolve device FQDNs and credentials from a real HMS.
// For local development the LocalHMS implementation returns static mock values
// so tests and local runs do not require external dependencies.
package hms

import (
"os"
)

// HMSInterface abstracts device resolution and credential lookup.
type HMSInterface interface {
HMSInit() error
GetDeviceFQDN(deviceName string) (string, error)
GetDeviceCredentials(deviceName string) (string, string, error)
}

// LocalHMS is a no-op HMS implementation for local development.
type LocalHMS struct{}

// NewLocalHMS returns a LocalHMS instance.
func NewLocalHMS() *LocalHMS {
return &LocalHMS{}
}

func (h *LocalHMS) HMSInit() error {
return nil
}

// GetDeviceFQDN mocks FQDN resolution. Falls back to "<deviceName>.local" if
// no real HMS is available; callers should prefer DeviceProfile.ManagementIp
// when this returns an unresolvable hostname.
func (h *LocalHMS) GetDeviceFQDN(deviceName string) (string, error) {
return deviceName + ".local", nil
}

// GetDeviceCredentials returns BMC credentials from environment variables,
// falling back to static mock values so unit tests never fail due to missing env.
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