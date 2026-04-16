// Package redfish provides HTTP client utilities for communicating with BMCs.
// TLS verification is intentionally disabled because BMC certificates are
// typically self-signed in lab and production environments.
package redfish

import (
"bytes"
"context"
"crypto/tls"
"fmt"
"io"
"net/http"
"net/url"
"strings"
"time"
)

// SendSecureRedfish executes HTTP requests against BMCs, bypassing TLS verification.
func SendSecureRedfish(server string, path string, payload string, user string, password string, method string, timeoutSeconds int) (string, error) {
u, _ := url.Parse("https://" + server + path)

var reqBody io.Reader
if payload != "" {
reqBody = bytes.NewBuffer([]byte(payload))
}

req, err := http.NewRequest(method, u.String(), reqBody)
if err != nil {
return "", fmt.Errorf("failed to create HTTP request: %v", err)
}

if user != "" && password != "" {
req.SetBasicAuth(user, password)
}

reqContext, reqCtxCancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeoutSeconds))
defer reqCtxCancel()

req = req.WithContext(reqContext)
req.Header.Add("Content-Type", "application/json")
req.Header.Add("cache-control", "no-cache")

client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
resp, err := client.Do(req)
if err != nil {
return "", fmt.Errorf("HTTP request failed: %v", err)
}
defer resp.Body.Close()

body, err := io.ReadAll(resp.Body)
if err != nil {
return "", fmt.Errorf("failed to read response body: %v", err)
}
if resp.StatusCode >= 400 {
return string(body), fmt.Errorf("redfish request failed with HTTP %d %s: %s", resp.StatusCode, resp.Status, string(body))
}
return string(body), nil
}

// GetLocationHeader executes a POST and returns the Location header value.
// Used to obtain the task URI from a Redfish SimpleUpdate response.
func GetLocationHeader(server string, path string, payload string, user string, password string, timeoutSeconds int) (string, error) {
u, _ := url.Parse("https://" + server + path)

var reqBody io.Reader
if payload != "" {
reqBody = bytes.NewBuffer([]byte(payload))
}

req, err := http.NewRequest(http.MethodPost, u.String(), reqBody)
if err != nil {
return "", fmt.Errorf("failed to create HTTP request: %v", err)
}
if user != "" && password != "" {
req.SetBasicAuth(user, password)
}

reqCtx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeoutSeconds))
defer cancel()
req = req.WithContext(reqCtx)
req.Header.Add("Content-Type", "application/json")
req.Header.Add("cache-control", "no-cache")

httpClient := &http.Client{
CheckRedirect: func(req *http.Request, via []*http.Request) error {
return http.ErrUseLastResponse
},
Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
}
resp, err := httpClient.Do(req)
if err != nil {
return "", fmt.Errorf("HTTP request failed: %v", err)
}
defer resp.Body.Close()

loc := resp.Header.Get("Location")
if loc == "" {
return "", fmt.Errorf("no Location header in response (HTTP %d)", resp.StatusCode)
}
return loc, nil
}

// GetNestedValue traverses a nested map using a dot-separated path string.
func GetNestedValue(data map[string]interface{}, path string) (interface{}, bool) {
keys := strings.Split(path, ".")
var current interface{} = data

for _, key := range keys {
if key == "" {
continue
}
if m, ok := current.(map[string]interface{}); ok {
if val, exists := m[key]; exists {
current = val
} else {
return nil, false
}
} else {
return nil, false
}
}
return current, true
}