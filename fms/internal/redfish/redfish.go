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
	url, _ := url.Parse("https://" + server + path)

	var reqBody io.Reader
	if payload != "" {
		reqBody = bytes.NewBuffer([]byte(payload))
	}

	req, err := http.NewRequest(method, url.String(), reqBody)
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