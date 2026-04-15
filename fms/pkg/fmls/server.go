// Package fmls implements the Firmware Management Library Server, which
// exposes firmware binaries stored under /tmp/firmware via HTTP. It supports:
//   - GET  /library/<filename>  — download a firmware binary
//   - POST /library/            — upload a new binary (multipart/form-data, field "file")
//   - DELETE /library/<filename> — remove a binary
package fmls

import (
"fmt"
"io"
"log"
"net/http"
"os"
"path/filepath"
"strings"
)

const firmwareDir = "/tmp/firmware"

// Server is an HTTP handler for the firmware library endpoints.
type Server struct {
mux *http.ServeMux
}

// NewServer constructs and registers all /library/ routes.
func NewServer() (*Server, error) {
if err := os.MkdirAll(firmwareDir, 0o755); err != nil {
return nil, fmt.Errorf("fmls: creating firmware dir: %w", err)
}

s := &Server{mux: http.NewServeMux()}
s.mux.HandleFunc("/library/", s.libraryHandler)
return s, nil
}

// ServeHTTP satisfies http.Handler so Server can be mounted into any mux.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
s.mux.ServeHTTP(w, r)
}

// libraryHandler dispatches GET / POST / DELETE for /library/<filename>.
func (s *Server) libraryHandler(w http.ResponseWriter, r *http.Request) {
switch r.Method {
case http.MethodGet:
s.handleGet(w, r)
case http.MethodPost:
s.handlePost(w, r)
case http.MethodDelete:
s.handleDelete(w, r)
default:
http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}
}

// handleGet serves a firmware binary by filename.
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
name := strings.TrimPrefix(r.URL.Path, "/library/")
if name == "" {
http.Error(w, "filename required", http.StatusBadRequest)
return
}

target := filepath.Join(firmwareDir, filepath.Base(name))
f, err := os.Open(target)
if err != nil {
if os.IsNotExist(err) {
http.Error(w, "not found", http.StatusNotFound)
return
}
log.Printf("fmls: GET open %q: %v", target, err)
http.Error(w, "internal error", http.StatusInternalServerError)
return
}
defer f.Close()

w.Header().Set("Content-Type", "application/octet-stream")
if _, err := io.Copy(w, f); err != nil {
log.Printf("fmls: GET copy %q: %v", target, err)
}
}

// handlePost accepts a multipart upload and writes the binary to firmwareDir.
func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
// Limit upload size to 500 MB to prevent unbounded reads.
if err := r.ParseMultipartForm(500 << 20); err != nil {
http.Error(w, "failed to parse multipart form", http.StatusBadRequest)
return
}

file, header, err := r.FormFile("file")
if err != nil {
http.Error(w, "field 'file' required", http.StatusBadRequest)
return
}
defer file.Close()

// Sanitise the filename: strip any path components the client may supply.
safeName := filepath.Base(header.Filename)
if safeName == "." || safeName == "/" {
http.Error(w, "invalid filename", http.StatusBadRequest)
return
}

dest := filepath.Join(firmwareDir, safeName)
out, err := os.Create(dest)
if err != nil {
log.Printf("fmls: POST create %q: %v", dest, err)
http.Error(w, "internal error", http.StatusInternalServerError)
return
}
defer out.Close()

if _, err := io.Copy(out, file); err != nil {
log.Printf("fmls: POST copy %q: %v", dest, err)
http.Error(w, "internal error", http.StatusInternalServerError)
return
}

w.WriteHeader(http.StatusCreated)
fmt.Fprintf(w, `{"stored":"%s"}`, safeName)
}

// handleDelete removes a firmware binary by filename.
func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
name := strings.TrimPrefix(r.URL.Path, "/library/")
if name == "" {
http.Error(w, "filename required", http.StatusBadRequest)
return
}

target := filepath.Join(firmwareDir, filepath.Base(name))
if err := os.Remove(target); err != nil {
if os.IsNotExist(err) {
http.Error(w, "not found", http.StatusNotFound)
return
}
log.Printf("fmls: DELETE remove %q: %v", target, err)
http.Error(w, "internal error", http.StatusInternalServerError)
return
}

w.WriteHeader(http.StatusNoContent)
}