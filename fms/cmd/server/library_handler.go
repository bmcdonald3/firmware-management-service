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

v1 "github.com/bmcdonald3/fms/apis/example.fabrica.dev/v1"
"github.com/bmcdonald3/fms/internal/storage"
"github.com/bmcdonald3/fms/internal/ziphelper"
"github.com/google/uuid"
"github.com/openchami/fabrica/pkg/fabrica"
)

const firmwareDestDir = "/tmp/firmware"

// manifest represents the expected structure of manifest.json inside a firmware zip bundle.
type manifest struct {
Name            string `json:"name"`
VersionString   string `json:"versionString"`
VersionNumber   string `json:"versionNumber"`
TargetComponent string `json:"targetComponent"`
SoftwareId      string `json:"softwareId"`
PreConditions   string `json:"preConditions"`
PostConditions  string `json:"postConditions"`
}

// libraryUploadHandler accepts a multipart zip upload, extracts it, parses manifest.json,
// and persists a FirmwareProfile record so the firmware is discoverable by the reconcilers.
func libraryUploadHandler(w http.ResponseWriter, r *http.Request) {
if err := r.ParseMultipartForm(64 << 20); err != nil {
http.Error(w, fmt.Sprintf("failed to parse multipart form: %v", err), http.StatusBadRequest)
return
}

file, _, err := r.FormFile("file")
if err != nil {
http.Error(w, fmt.Sprintf("missing 'file' field: %v", err), http.StatusBadRequest)
return
}
defer file.Close()

// Write uploaded zip to a temp file so we can pass it to zip.OpenReader.
tmp, err := os.CreateTemp("", "firmware-*.zip")
if err != nil {
http.Error(w, "failed to create temp file", http.StatusInternalServerError)
return
}
defer os.Remove(tmp.Name())
defer tmp.Close()

if _, err := io.Copy(tmp, file); err != nil {
http.Error(w, "failed to write temp file", http.StatusInternalServerError)
return
}
tmp.Close()

zr, err := zip.OpenReader(tmp.Name())
if err != nil {
http.Error(w, fmt.Sprintf("failed to open zip: %v", err), http.StatusBadRequest)
return
}
defer zr.Close()

if err := os.MkdirAll(firmwareDestDir, 0o755); err != nil {
http.Error(w, "failed to create firmware directory", http.StatusInternalServerError)
return
}

var mf manifest
foundManifest := false

for _, f := range zr.File {
if filepath.Base(f.Name) == "manifest.json" {
rc, err := f.Open()
if err != nil {
http.Error(w, "failed to open manifest.json", http.StatusInternalServerError)
return
}
if err := json.NewDecoder(rc).Decode(&mf); err != nil {
rc.Close()
http.Error(w, fmt.Sprintf("failed to parse manifest.json: %v", err), http.StatusBadRequest)
return
}
rc.Close()
foundManifest = true
continue
}
if err := ziphelper.ExtractEntry(f, firmwareDestDir); err != nil {
log.Printf("warning: failed to extract %s: %v", f.Name, err)
}
}

if !foundManifest {
http.Error(w, "zip does not contain manifest.json", http.StatusBadRequest)
return
}

name := mf.Name
if name == "" {
name = fmt.Sprintf("firmware-%s", uuid.NewString()[:8])
}

profile := &v1.FirmwareProfile{
APIVersion: "example.fabrica.dev/v1",
Kind:       "FirmwareProfile",
Metadata: fabrica.Metadata{
Name: name,
UID:  uuid.NewString(),
},
Spec: v1.FirmwareProfileSpec{
VersionString:   mf.VersionString,
VersionNumber:   mf.VersionNumber,
TargetComponent: mf.TargetComponent,
SoftwareId:      mf.SoftwareId,
PreConditions:   mf.PreConditions,
PostConditions:  mf.PostConditions,
},
}

if err := storage.SaveFirmwareProfile(r.Context(), profile); err != nil {
http.Error(w, fmt.Sprintf("failed to save firmware profile: %v", err), http.StatusInternalServerError)
return
}

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusCreated)
json.NewEncoder(w).Encode(profile)
}