// Package main — library_handler.go
// Provides /library/upload (POST) and /library/{filename} (GET) endpoints.
// The upload endpoint accepts a multipart ZIP bundle, extracts it to
// /tmp/firmware/, parses a manifest.json inside the bundle, and persists a
// FirmwareProfile record via the Fabrica storage client.
package main

import (
"archive/zip"
"encoding/json"
"fmt"
"io"
"log"
"net/http"
"os"
"path/filepath"
"strings"

"github.com/go-chi/chi/v5"
v1 "github.com/bmcdonald3/fms/apis/firmware.management.io/v1"
"github.com/bmcdonald3/fms/internal/storage"
"github.com/bmcdonald3/fms/internal/ziphelper"
"github.com/openchami/fabrica/pkg/resource"
"github.com/openchami/fabrica/pkg/fabrica"
)

const firmwareBasePath = "/tmp/firmware"

// firmwareManifest mirrors the expected manifest.json structure inside a ZIP bundle.
type firmwareManifest struct {
Name            string   `json:"name"`
VersionString   string   `json:"versionString"`
VersionNumber   string   `json:"versionNumber"`
TargetComponent string   `json:"targetComponent"`
SoftwareId      string   `json:"softwareId"`
PreConditions   []string `json:"preConditions"`
PostConditions  []string `json:"postConditions"`
}

// RegisterLibraryRoutes injects the library endpoints onto the provided chi router.
// Called from main.go after RegisterGeneratedRoutes so the custom routes take
// precedence over any generated catch-all handlers.
func RegisterLibraryRoutes(r chi.Router) {
r.Post("/library/upload", handleLibraryUpload)
r.Get("/library/{filename}", handleLibraryServe)
}

// handleLibraryUpload accepts a multipart/form-data POST with a "file" field
// containing a ZIP archive. It extracts the archive to /tmp/firmware/, reads
// manifest.json, and creates a FirmwareProfile resource in the database.
func handleLibraryUpload(w http.ResponseWriter, r *http.Request) {
// 32 MB max in-memory; remainder spills to temp files.
if err := r.ParseMultipartForm(32 << 20); err != nil {
http.Error(w, fmt.Sprintf("failed to parse multipart form: %v", err), http.StatusBadRequest)
return
}

file, header, err := r.FormFile("file")
if err != nil {
http.Error(w, fmt.Sprintf("missing 'file' field: %v", err), http.StatusBadRequest)
return
}
defer file.Close()

// Write the upload to a temp file so we can open it as a zip.Reader.
tmp, err := os.CreateTemp("", "fms-upload-*.zip")
if err != nil {
http.Error(w, "failed to create temp file", http.StatusInternalServerError)
return
}
defer os.Remove(tmp.Name())
defer tmp.Close()

size, err := io.Copy(tmp, file)
if err != nil {
http.Error(w, "failed to buffer upload", http.StatusInternalServerError)
return
}

zr, err := zip.NewReader(tmp, size)
if err != nil {
http.Error(w, fmt.Sprintf("invalid zip file: %v", err), http.StatusBadRequest)
return
}

destDir := filepath.Join(firmwareBasePath, strings.TrimSuffix(header.Filename, ".zip"))
if err := os.MkdirAll(destDir, 0o755); err != nil {
http.Error(w, "failed to create destination directory", http.StatusInternalServerError)
return
}

var manifest *firmwareManifest
for _, f := range zr.File {
if err := ziphelper.ExtractEntry(f, destDir); err != nil {
log.Printf("WARN: skipping zip entry %q: %v", f.Name, err)
continue
}
if f.Name == "manifest.json" {
manifest, err = parseManifest(filepath.Join(destDir, "manifest.json"))
if err != nil {
log.Printf("WARN: failed to parse manifest.json: %v", err)
}
}
}

if manifest == nil {
http.Error(w, "zip bundle must contain a manifest.json", http.StatusBadRequest)
return
}

uid, err := resource.GenerateUIDForResource("FirmwareProfile")
if err != nil {
http.Error(w, fmt.Sprintf("failed to generate UID: %v", err), http.StatusInternalServerError)
return
}

profile := &v1.FirmwareProfile{
APIVersion: "firmware.management.io/v1",
Kind:       "FirmwareProfile",
Metadata: fabrica.Metadata{
Name: manifest.Name,
UID:  uid,
},
Spec: v1.FirmwareProfileSpec{
VersionString:   manifest.VersionString,
VersionNumber:   manifest.VersionNumber,
TargetComponent: manifest.TargetComponent,
SoftwareId:      manifest.SoftwareId,
PreConditions:   manifest.PreConditions,
PostConditions:  manifest.PostConditions,
},
}

client := storage.NewStorageClient()
if err := client.Create(r.Context(), profile); err != nil {
http.Error(w, fmt.Sprintf("failed to persist FirmwareProfile: %v", err), http.StatusInternalServerError)
return
}

log.Printf("INFO: FirmwareProfile %q created (uid=%s)", profile.Metadata.Name, profile.Metadata.UID)

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusCreated)
json.NewEncoder(w).Encode(profile)
}

// handleLibraryServe serves binary files from /tmp/firmware/ by filename.
func handleLibraryServe(w http.ResponseWriter, r *http.Request) {
filename := chi.URLParam(r, "filename")
// Reject any path traversal attempts.
if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
http.Error(w, "invalid filename", http.StatusBadRequest)
return
}

// Walk /tmp/firmware looking for the requested file so bundles in
// sub-directories are reachable via a flat URL namespace.
var found string
err := filepath.Walk(firmwareBasePath, func(path string, info os.FileInfo, err error) error {
if err != nil {
return err
}
if !info.IsDir() && info.Name() == filename {
found = path
return filepath.SkipAll
}
return nil
})
if err != nil || found == "" {
http.Error(w, "file not found", http.StatusNotFound)
return
}

http.ServeFile(w, r, found)
}

// parseManifest reads and unmarshals the manifest JSON file at path.
func parseManifest(path string) (*firmwareManifest, error) {
data, err := os.ReadFile(path)
if err != nil {
return nil, fmt.Errorf("reading manifest: %w", err)
}
var m firmwareManifest
if err := json.Unmarshal(data, &m); err != nil {
return nil, fmt.Errorf("parsing manifest JSON: %w", err)
}
return &m, nil
}