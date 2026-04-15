// Package credentials provides parsing and lookup of node credentials
// from a JSON file. The file schema is:
//   {"nodes": {"<nodeName>": {"username": "...", "password": "..."}}}
package credentials

import (
	"encoding/json"
	"fmt"
	"os"
)

// NodeCredential holds the username and password for a single node.
type NodeCredential struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// credentialsFile mirrors the top-level structure of credentials.json.
type credentialsFile struct {
	Nodes map[string]NodeCredential `json:"nodes"`
}

// Parser loads and exposes node credentials from a JSON file on disk.
type Parser struct {
	path  string
	store map[string]NodeCredential
}

// NewParser creates a Parser and eagerly loads credentials from path.
func NewParser(path string) (*Parser, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("credentials: reading file %q: %w", path, err)
	}

	var cf credentialsFile
	if err := json.Unmarshal(data, &cf); err != nil {
		return nil, fmt.Errorf("credentials: parsing JSON from %q: %w", path, err)
	}

	if cf.Nodes == nil {
		cf.Nodes = make(map[string]NodeCredential)
	}

	return &Parser{path: path, store: cf.Nodes}, nil
}

// Get returns the credential for nodeName, or an error if it is not found.
func (p *Parser) Get(nodeName string) (NodeCredential, error) {
	cred, ok := p.store[nodeName]
	if !ok {
		return NodeCredential{}, fmt.Errorf("credentials: no entry for node %q", nodeName)
	}
	return cred, nil
}