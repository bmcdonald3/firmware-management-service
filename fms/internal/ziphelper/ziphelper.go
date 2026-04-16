// Package ziphelper provides safe ZIP extraction utilities.
// Zip-slip prevention is implemented by verifying that every extracted path
// remains within the intended destination directory before writing.
package ziphelper

import (
"archive/zip"
"fmt"
"io"
"os"
"path/filepath"
"strings"
)

// ExtractEntry safely extracts a single zip entry into destDir, preventing zip-slip.
func ExtractEntry(f *zip.File, destDir string) error {
destPath, err := filepath.Abs(filepath.Join(destDir, f.Name))
if err != nil {
return err
}
baseDir, err := filepath.Abs(destDir)
if err != nil {
return err
}
if !strings.HasPrefix(destPath, baseDir+string(os.PathSeparator)) {
return fmt.Errorf("illegal path in zip entry %q", f.Name)
}

if f.FileInfo().IsDir() {
return os.MkdirAll(destPath, 0o755)
}

if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
return err
}

rc, err := f.Open()
if err != nil {
return err
}
defer rc.Close()

out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
if err != nil {
return err
}
defer out.Close()

_, err = io.Copy(out, rc)
return err
}