package credentials

import (
	"encoding/json"
	"fmt"
	"os"
)

const credentialsFile = "/tmp/credentials.json"

// Creds holds username/password for a node.
type Creds struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type credsFile struct {
	Nodes map[string]Creds `json:"nodes"`
}

// GetCredentials reads the credentials file and returns username/password for the given node.
// Rationale: centralize credential parsing so callers can mock/replace during tests.
func GetCredentials(node string) (string, string, error) {
	data, err := os.ReadFile(credentialsFile)
	if err != nil {
		return "", "", fmt.Errorf("reading credentials file %s: %w", credentialsFile, err)
	}

	var cf credsFile
	if err := json.Unmarshal(data, &cf); err != nil {
		return "", "", fmt.Errorf("parsing credentials JSON: %w", err)
	}

	creds, ok := cf.Nodes[node]
	if !ok {
		return "", "", fmt.Errorf("credentials for node %q not found", node)
	}

	return creds.Username, creds.Password, nil
}