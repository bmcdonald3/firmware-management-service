package fmls

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const firmwareDir = "/tmp/firmware"

// RegisterHandlers registers /library/ handlers on the provided mux.
func RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/library/", libraryHandler)
}

// libraryHandler handles:
// - GET  /library/           -> JSON list of files
// - GET  /library/{name}     -> download file
// - POST /library/           -> multipart/form-data upload (field "file")
// - DELETE /library/{name}   -> delete file
func libraryHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure firmware directory exists
	if err := os.MkdirAll(firmwareDir, 0o755); err != nil {
		http.Error(w, "internal: cannot create firmware dir", http.StatusInternalServerError)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/library/")
	switch r.Method {
	case http.MethodGet:
		if name == "" {
			// list files
			files, err := os.ReadDir(firmwareDir)
			if err != nil {
				http.Error(w, "failed to read firmware dir", http.StatusInternalServerError)
				return
			}
			names := make([]string, 0, len(files))
			for _, f := range files {
				if !f.IsDir() {
					names = append(names, f.Name())
				}
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string][]string{"files": names})
			return
		}
		// serve file
		fp := filepath.Join(firmwareDir, filepath.Clean(name))
		f, err := os.Open(fp)
		if err != nil {
			if os.IsNotExist(err) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "failed to open file", http.StatusInternalServerError)
			return
		}
		defer f.Close()
		http.ServeContent(w, r, name, time.Now(), f)
	case http.MethodPost:
		// multipart upload
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "invalid multipart form", http.StatusBadRequest)
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "missing file field", http.StatusBadRequest)
			return
		}
		defer file.Close()

		dstPath := filepath.Join(firmwareDir, filepath.Clean(header.Filename))
		out, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			http.Error(w, "failed to create file", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		if _, err := io.Copy(out, file); err != nil {
			http.Error(w, "failed to write file", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "filename": header.Filename})
	case http.MethodDelete:
		if name == "" {
			http.Error(w, "filename required", http.StatusBadRequest)
			return
		}
		fp := filepath.Join(firmwareDir, filepath.Clean(name))
		if err := os.Remove(fp); err != nil {
			if os.IsNotExist(err) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, fmt.Sprintf("failed to delete: %v", err), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "GET,POST,DELETE")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}